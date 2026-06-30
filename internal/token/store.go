package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// tokensFile is the name of the JSON array persisted under the store's path.
const tokensFile = "tokens.json"

var (
	// ErrNotFound is returned when no token with the given id exists.
	ErrNotFound = errors.New("token: not found")
	// ErrInvalid is returned when a token exists but the secret does not
	// match or it has no uses remaining.
	ErrInvalid = errors.New("token: invalid secret or no uses remaining")
	// ErrExpired is returned when the secret matched but the token's TTL has
	// passed.
	ErrExpired = errors.New("token: expired")
)

// Store persists tokens as a single JSON array at <path>/tokens.json.
type Store struct {
	path string
	mu   sync.Mutex
}

// NewStore returns a Store rooted at path, creating the directory (mode 0700)
// if it does not exist.
func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, fmt.Errorf("token: create state dir: %w", err)
	}
	return &Store{path: path}, nil
}

func (s *Store) tokensPath() string {
	return filepath.Join(s.path, tokensFile)
}

// load reads the full token list. A missing or empty file is an empty list.
// The caller must hold s.mu.
func (s *Store) load() ([]Token, error) {
	data, err := os.ReadFile(s.tokensPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("token: read store: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	var tokens []Token
	if err = json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("token: parse store: %w", err)
	}
	return tokens, nil
}

// save atomically replaces tokens.json: write to a temp file, fsync, then
// rename over the real file. The file is created with mode 0600. The caller
// must hold s.mu.
func (s *Store) save(tokens []Token) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("token: encode store: %w", err)
	}

	tmp := s.tokensPath() + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("token: open temp store: %w", err)
	}
	// On any error after this point, drop the temp file rather than leak it.
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("token: write temp store: %w", err)
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("token: fsync temp store: %w", err)
	}
	if err = f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("token: close temp store: %w", err)
	}
	if err = os.Rename(tmp, s.tokensPath()); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("token: commit store: %w", err)
	}
	return nil
}

// Create generates a new token, persists it, and returns the raw
// "<id>.<secret>" string. The raw string is the only place the secret ever
// appears — it is not recoverable from the store.
func (s *Store) Create(ttl time.Duration, uses int, desc string) (raw string, t Token, err error) {
	raw, t, err = Generate(ttl, uses, desc)
	if err != nil {
		return "", Token{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return "", Token{}, err
	}
	tokens = append(tokens, t)
	if err = s.save(tokens); err != nil {
		return "", Token{}, err
	}
	return raw, t, nil
}

// Validate looks up the token by id, consumes one use, and persists the
// result. On success it returns the consumed token (with its decremented
// counter). A successful consume that exhausts the token (RemainingUses
// reaches 0) removes it from the store; an expired token is also removed.
//
// The whole lookup-consume-persist sequence runs under the mutex so two
// concurrent Validate calls cannot double-spend a single-use token.
func (s *Store) Validate(raw string) (*Token, error) {
	id, secret, err := Parse(raw)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return nil, err
	}

	idx := indexOf(tokens, id)
	if idx < 0 {
		return nil, ErrNotFound
	}

	// Operate on a copy; only write it back once we know what happened.
	tok := tokens[idx]
	ok, expired := tok.MatchAndConsume(secret, time.Now())
	if !ok {
		if expired {
			// Self-clean: a token whose TTL passed can never succeed again.
			tokens = append(tokens[:idx], tokens[idx+1:]...)
			if err = s.save(tokens); err != nil {
				return nil, err
			}
			return nil, ErrExpired
		}
		// Wrong secret or exhausted — no state change.
		return nil, ErrInvalid
	}

	if tok.RemainingUses == 0 {
		tokens = append(tokens[:idx], tokens[idx+1:]...)
	} else {
		tokens[idx] = tok
	}
	if err = s.save(tokens); err != nil {
		return nil, err
	}
	return &tok, nil
}

// List returns all stored tokens with SecretHash zeroed, so callers cannot
// accidentally leak the hash.
func (s *Store) List() ([]Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return nil, err
	}
	for i := range tokens {
		tokens[i].SecretHash = nil
	}
	return tokens, nil
}

// Revoke removes the token with the given id. It returns ErrNotFound if no
// such token exists.
func (s *Store) Revoke(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return err
	}

	idx := indexOf(tokens, id)
	if idx < 0 {
		return ErrNotFound
	}
	tokens = append(tokens[:idx], tokens[idx+1:]...)
	return s.save(tokens)
}

// indexOf returns the position of the token with the given id, or -1.
func indexOf(tokens []Token, id string) int {
	for i := range tokens {
		if tokens[i].ID == id {
			return i
		}
	}
	return -1
}
