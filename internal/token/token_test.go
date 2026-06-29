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
