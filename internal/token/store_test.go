package token

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newStore returns a Store rooted at a fresh temp dir.
func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(t.TempDir())
	require.NoError(t, err)
	return s
}

func TestStore_CreateAndValidate(t *testing.T) {
	s := newStore(t)

	raw, created, err := s.Create(time.Hour, 1, "join web-1")
	require.NoError(t, err)

	got, err := s.Validate(raw)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "join web-1", got.Description)
}

func TestStore_ValidateWrongSecret(t *testing.T) {
	s := newStore(t)

	raw, _, err := s.Create(time.Hour, 2, "")
	require.NoError(t, err)

	// Same id, bogus secret.
	id, _, err := Parse(raw)
	require.NoError(t, err)
	_, err = s.Validate(id + "." + "0000000000000000")
	require.ErrorIs(t, err, ErrInvalid)

	// The valid secret must still have both its uses: no state was mutated.
	_, err = s.Validate(raw)
	require.NoError(t, err)
	tokens, err := s.List()
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, 1, tokens[0].RemainingUses)
}

func TestStore_ValidateExpired(t *testing.T) {
	s := newStore(t)

	// TTL already in the past.
	raw, _, err := s.Create(-time.Minute, 1, "")
	require.NoError(t, err)

	_, err = s.Validate(raw)
	require.ErrorIs(t, err, ErrExpired)

	// Documented choice: an expired token is removed from the store.
	tokens, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestStore_ValidateSingleUseRemovesToken(t *testing.T) {
	s := newStore(t)

	raw, _, err := s.Create(time.Hour, 1, "")
	require.NoError(t, err)

	got, err := s.Validate(raw)
	require.NoError(t, err)
	assert.Equal(t, 0, got.RemainingUses)

	// Token is gone, so a second use reports not found.
	_, err = s.Validate(raw)
	require.ErrorIs(t, err, ErrNotFound)

	tokens, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestStore_ValidateUnlimitedNeverDecrements(t *testing.T) {
	s := newStore(t)

	raw, _, err := s.Create(time.Hour, -1, "")
	require.NoError(t, err)

	for range 5 {
		got, err := s.Validate(raw)
		require.NoError(t, err)
		assert.Equal(t, -1, got.RemainingUses)
	}

	tokens, err := s.List()
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, -1, tokens[0].RemainingUses)
}

func TestStore_Revoke(t *testing.T) {
	s := newStore(t)

	_, created, err := s.Create(time.Hour, 1, "")
	require.NoError(t, err)

	require.NoError(t, s.Revoke(created.ID))

	tokens, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, tokens)

	// Revoking again is a not-found, not a panic.
	require.ErrorIs(t, s.Revoke(created.ID), ErrNotFound)
}

func TestStore_ListHidesSecretHash(t *testing.T) {
	s := newStore(t)

	_, _, err := s.Create(time.Hour, 1, "")
	require.NoError(t, err)

	tokens, err := s.List()
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Nil(t, tokens[0].SecretHash, "List must zero SecretHash")
}

func TestStore_FilePermissions(t *testing.T) {
	s := newStore(t)

	_, _, err := s.Create(time.Hour, 1, "")
	require.NoError(t, err)

	info, err := os.Stat(filepath.Join(s.path, tokensFile))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

// TestStore_ConcurrentValidateNoDoubleSpend asserts that a single-use token
// validated from many goroutines succeeds exactly once. Run with -race.
func TestStore_ConcurrentValidateNoDoubleSpend(t *testing.T) {
	s := newStore(t)

	raw, _, err := s.Create(time.Hour, 1, "")
	require.NoError(t, err)

	const goroutines = 20
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			if _, err := s.Validate(raw); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, 1, successes, "a single-use token must be consumable exactly once")

	tokens, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, tokens)
}
