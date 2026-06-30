package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestGenerate(t *testing.T) {
	raw, tok, err := Generate(time.Hour, 1, "a one hour token, usable once")
	require.NoError(t, err)

	// The raw token must be a well-formed <id>.<secret> pair...
	id, secret, err := Parse(raw)
	require.NoError(t, err)

	// ...whose id matches the struct and whose secret verifies against the
	// stored hash (and is therefore the plaintext, not the hash).
	assert.Equal(t, tok.ID, id)
	assert.NoError(t, bcrypt.CompareHashAndPassword(tok.SecretHash, []byte(secret)))

	assert.Equal(t, "a one hour token, usable once", tok.Description)
	assert.Equal(t, 1, tok.RemainingUses)
	assert.Equal(t, tok.CreatedAt.Add(time.Hour), tok.ExpiresAt)
}

func TestGenerate_Unique(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for range 100 {
		raw, _, err := Generate(time.Hour, 1, "")
		require.NoError(t, err)
		_, exists := seen[raw]
		assert.False(t, exists, "Generate produced a duplicate raw token: %s", raw)
		seen[raw] = struct{}{}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantID     string
		wantSecret string
		wantErr    bool
	}{
		{
			name:       "valid token",
			raw:        "abc123.def456789abcdef0",
			wantID:     "abc123",
			wantSecret: "def456789abcdef0",
		},
		{
			name:    "missing separator",
			raw:     "abc123def456789abcdef0",
			wantErr: true,
		},
		{
			name:    "too many separators",
			raw:     "abc123.def456.789abcdef",
			wantErr: true,
		},
		{
			name:    "empty string",
			raw:     "",
			wantErr: true,
		},
		{
			name:    "id too short",
			raw:     "abc1.def456789abcdef0",
			wantErr: true,
		},
		{
			name:    "id too long",
			raw:     "abc1234.def456789abcdef0",
			wantErr: true,
		},
		{
			name:    "secret too short",
			raw:     "abc123.def456",
			wantErr: true,
		},
		{
			name:    "secret too long",
			raw:     "abc123.def456789abcdef0123",
			wantErr: true,
		},
		{
			name:    "non-hex id",
			raw:     "abcxyz.def456789abcdef0",
			wantErr: true,
		},
		{
			// 'g' is non-hex but the length is still exactly 16, so this
			// proves hex validation is independent of the length check.
			name:    "non-hex secret",
			raw:     "abc123.def456789abcdeg0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, secret, err := Parse(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, id)
				assert.Empty(t, secret)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantSecret, secret)
		})
	}
}

// newToken returns a freshly generated token alongside its plaintext secret,
// so tests can exercise MatchAndConsume with a secret that verifies against
// the stored hash.
func newToken(t *testing.T, ttl time.Duration, uses int) (Token, string) {
	t.Helper()
	raw, tok, err := Generate(ttl, uses, "")
	require.NoError(t, err)
	_, secret, err := Parse(raw)
	require.NoError(t, err)
	return tok, secret
}

func TestMatchAndConsume(t *testing.T) {
	t.Run("valid secret consumes one use", func(t *testing.T) {
		tok, secret := newToken(t, time.Hour, 3)

		ok, expired := tok.MatchAndConsume(secret, time.Now())
		assert.True(t, ok)
		assert.False(t, expired)
		assert.Equal(t, 2, tok.RemainingUses, "a successful match must decrement RemainingUses")
	})

	t.Run("wrong secret fails without consuming", func(t *testing.T) {
		tok, _ := newToken(t, time.Hour, 3)

		ok, expired := tok.MatchAndConsume("0000000000000000", time.Now())
		assert.False(t, ok)
		assert.False(t, expired, "a wrong secret must not be reported as expired")
		assert.Equal(t, 3, tok.RemainingUses, "a failed match must not decrement RemainingUses")
	})

	t.Run("expired token reports expired without consuming", func(t *testing.T) {
		tok, secret := newToken(t, time.Hour, 3)

		ok, expired := tok.MatchAndConsume(secret, tok.ExpiresAt.Add(time.Second))
		assert.False(t, ok)
		assert.True(t, expired)
		assert.Equal(t, 3, tok.RemainingUses, "an expired match must not decrement RemainingUses")
	})

	t.Run("expiry is inclusive at the boundary", func(t *testing.T) {
		tok, secret := newToken(t, time.Hour, 3)

		// Exactly at ExpiresAt the token is still valid: now.After(ExpiresAt)
		// is false, so this consumes a use.
		ok, expired := tok.MatchAndConsume(secret, tok.ExpiresAt)
		assert.True(t, ok)
		assert.False(t, expired)
		assert.Equal(t, 2, tok.RemainingUses)
	})

	t.Run("wrong secret takes precedence over expiry", func(t *testing.T) {
		tok, _ := newToken(t, time.Hour, 3)

		// The token is expired, but the secret is also wrong. Because the
		// secret is checked first, the caller learns nothing about expiry.
		ok, expired := tok.MatchAndConsume("0000000000000000", tok.ExpiresAt.Add(time.Second))
		assert.False(t, ok)
		assert.False(t, expired)
	})

	t.Run("exhausted token fails closed", func(t *testing.T) {
		tok, secret := newToken(t, time.Hour, 0)

		ok, expired := tok.MatchAndConsume(secret, time.Now())
		assert.False(t, ok)
		assert.False(t, expired)
		assert.Equal(t, 0, tok.RemainingUses)
	})

	t.Run("unlimited token never decrements", func(t *testing.T) {
		tok, secret := newToken(t, time.Hour, -1)

		for range 5 {
			ok, expired := tok.MatchAndConsume(secret, time.Now())
			assert.True(t, ok)
			assert.False(t, expired)
		}
		assert.Equal(t, -1, tok.RemainingUses, "an unlimited token must keep RemainingUses at -1")
	})

	t.Run("unlimited token still respects expiry", func(t *testing.T) {
		tok, secret := newToken(t, time.Hour, -1)

		ok, expired := tok.MatchAndConsume(secret, tok.ExpiresAt.Add(time.Second))
		assert.False(t, ok)
		assert.True(t, expired)
	})

	t.Run("negative uses other than -1 fail closed", func(t *testing.T) {
		// A corrupted counter must not be mistaken for unlimited use.
		tok, secret := newToken(t, time.Hour, -2)

		ok, expired := tok.MatchAndConsume(secret, time.Now())
		assert.False(t, ok)
		assert.False(t, expired)
		assert.Equal(t, -2, tok.RemainingUses)
	})

	t.Run("token is consumable until exhausted", func(t *testing.T) {
		tok, secret := newToken(t, time.Hour, 2)
		now := time.Now()

		ok, _ := tok.MatchAndConsume(secret, now)
		assert.True(t, ok)
		ok, _ = tok.MatchAndConsume(secret, now)
		assert.True(t, ok)

		// Third attempt: uses are exhausted.
		ok, expired := tok.MatchAndConsume(secret, now)
		assert.False(t, ok)
		assert.False(t, expired)
		assert.Equal(t, 0, tok.RemainingUses)
	})
}
