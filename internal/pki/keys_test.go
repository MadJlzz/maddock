package pki

import (
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundTripKey(t *testing.T) {
	key, err := GenerateKey()
	assert.NoError(t, err)

	pemKey, err := EncodeKeyPEM(key)
	assert.NoError(t, err)

	decodedKey, err := DecodeKeyPEM(pemKey)
	assert.NoError(t, err)

	assert.True(t, decodedKey.Equal(key))
}

func TestDecodeKeyPEM_Errors(t *testing.T) {
	wrongType := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("ignored")})
	malformed := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not pkcs8")})

	tests := []struct {
		name  string
		input []byte
	}{
		{"empty input", nil},
		{"not a pem block", []byte("just some random bytes")},
		{"wrong block type", wrongType},
		{"malformed pkcs8 payload", malformed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeKeyPEM(tc.input)
			assert.Error(t, err)
		})
	}
}

func TestDecodeCertPEM_Errors(t *testing.T) {
	wrongType := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("ignored")})
	malformed := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not a der cert")})

	tests := []struct {
		name  string
		input []byte
	}{
		{"empty input", nil},
		{"not a pem block", []byte("just some random bytes")},
		{"wrong block type", wrongType},
		{"malformed der payload", malformed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeCertPEM(tc.input)
			assert.Error(t, err)
		})
	}
}
