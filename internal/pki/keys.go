package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// GenerateKey returns a fresh ECDSA P-256 private key.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func EncodeCertPEM(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Bytes: der, Type: "CERTIFICATE"})
}

func EncodeKeyPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	kb, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("pki: marshal private key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Bytes: kb, Type: "PRIVATE KEY"}), nil
}

func DecodeCertPEM(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("pki: not a PEM block")
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("pki: expected CERTIFICATE PEM, got %q", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("pki: parse certificate: %w", err)
	}

	return cert, nil
}

func DecodeKeyPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("pki: not a PEM block")
	}
	if block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("pki: expected PRIVATE KEY PEM, got %q", block.Type)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("pki: parse private key: %w", err)
	}

	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("pki: expected ECDSA private key, got %T", key)
	}

	return ecdsaKey, nil
}
