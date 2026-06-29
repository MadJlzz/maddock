package token

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	// idHexLen and secretHexLen are the expected lengths of the hex-encoded
	// id and secret — twice the raw byte counts used in Generate (3 and 8).
	idHexLen     = 6
	secretHexLen = 16

	separator = "."
)

type Token struct {
	ID            string    `json:"id"`
	Description   string    `json:"description"`
	SecretHash    []byte    `json:"secret_hash"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	RemainingUses int       `json:"remaining_uses"`
}

// Generate creates a fresh token. Returns the raw "<id>.<secret>" string
// (which the operator must capture - it is not recoverable) and the Token
// struct ready to persist.
func Generate(ttl time.Duration, uses int, desc string) (raw string, t Token, err error) {
	rawId := make([]byte, 3)
	if _, err = rand.Read(rawId); err != nil {
		return "", Token{}, err
	}

	rawSecret := make([]byte, 8)
	if _, err = rand.Read(rawSecret); err != nil {
		return "", Token{}, err
	}

	hexId, hexSecret := hex.EncodeToString(rawId), hex.EncodeToString(rawSecret)
	bcryptSecret, err := bcrypt.GenerateFromPassword([]byte(hexSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", Token{}, err
	}

	rawToken := fmt.Sprintf("%s.%s", hexId, hexSecret)
	now := time.Now()
	return rawToken, Token{
		ID:            hexId,
		Description:   desc,
		SecretHash:    bcryptSecret,
		CreatedAt:     now,
		ExpiresAt:     now.Add(ttl),
		RemainingUses: uses,
	}, nil
}

// Parse splits a raw token string into id and secret. Returns an error if
// the format is wrong (no dot, wrong lengths, non-hex characters). The id and
// secret are returned as their original hex strings — Parse validates the
// encoding but does not decode them.
func Parse(raw string) (id, secret string, err error) {
	parts := strings.Split(raw, separator)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("token: expected format <id>%s<secret>", separator)
	}
	id, secret = parts[0], parts[1]

	if len(id) != idHexLen {
		return "", "", fmt.Errorf("token: id must be %d hex chars, got %d", idHexLen, len(id))
	}
	if len(secret) != secretHexLen {
		return "", "", fmt.Errorf("token: secret must be %d hex chars, got %d", secretHexLen, len(secret))
	}
	if _, err = hex.DecodeString(id); err != nil {
		return "", "", fmt.Errorf("token: id is not valid hex: %w", err)
	}
	if _, err = hex.DecodeString(secret); err != nil {
		return "", "", fmt.Errorf("token: secret is not valid hex: %w", err)
	}
	return id, secret, nil
}

// MatchAndConsume checks whether secret matches the bcrypt hash, the
// token isn't expired, and uses_remaining > 0. On success it decrements
// UsesRemaining and returns (true, false). On failure it returns
// (false, expired).
func (t *Token) MatchAndConsume(secret string, now time.Time) (ok bool, expired bool) {
	return false, false
}
