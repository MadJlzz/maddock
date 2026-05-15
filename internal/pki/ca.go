package pki

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	CertificateAuthorityCertName = "ca.crt"
	CertificateAuthorityKeyName  = "ca.key"
)

type CA struct {
	Cert *x509.Certificate
	Key  *ecdsa.PrivateKey
}

func GenerateCA(commonName string) (*CA, error) {
	pKey, err := GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("pki: generate CA key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, pKey.Public(), pKey)
	if err != nil {
		return nil, fmt.Errorf("pki: sign CA certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, fmt.Errorf("pki: parse CA certificate: %w", err)
	}

	return &CA{
		Cert: cert,
		Key:  pKey,
	}, nil
}

func Load(dir string) (*CA, error) {
	certBytes, err := os.ReadFile(filepath.Join(dir, CertificateAuthorityCertName))
	if err != nil {
		return nil, fmt.Errorf("pki: load CA certificate: %w", err)
	}
	cert, err := DecodeCertPEM(certBytes)
	if err != nil {
		return nil, fmt.Errorf("pki: load CA certificate: %w", err)
	}
	if !cert.IsCA {
		return nil, fmt.Errorf("pki: CA certificate is not a CA")
	}

	keyBytes, err := os.ReadFile(filepath.Join(dir, CertificateAuthorityKeyName))
	if err != nil {
		return nil, fmt.Errorf("pki: load CA certificate: %w", err)
	}
	key, err := DecodeKeyPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("pki: load CA certificate: %w", err)
	}

	return &CA{
		Cert: cert,
		Key:  key,
	}, nil
}

func (c *CA) Save(dir string) error {
	certPem := EncodeCertPEM(c.Cert.Raw)
	keyPem, err := EncodeKeyPEM(c.Key)
	if err != nil {
		return fmt.Errorf("pki: CA encode key: %w", err)
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("pki: CA directory: %w", err)
	}

	err = os.WriteFile(filepath.Join(dir, CertificateAuthorityCertName), certPem, 0644)
	if err != nil {
		return fmt.Errorf("pki: CA certificate: %w", err)
	}

	err = os.WriteFile(filepath.Join(dir, CertificateAuthorityKeyName), keyPem, 0600)
	if err != nil {
		return fmt.Errorf("pki: CA key: %w", err)
	}

	return nil
}

// SignAgentCert signs a cert with CN = hostname, SAN DNS = hostname.
// EKU includes both ServerAuth (the agent listens) and ClientAuth
// (so the same cert can be reused if agents ever initiate calls).
// Returns DER-encoded certificate bytes.
func (c *CA) SignAgentCert(hostname string, pub crypto.PublicKey, ttl time.Duration) ([]byte, error) {
	template := &x509.Certificate{
		Subject:     pkix.Name{CommonName: hostname},
		DNSNames:    []string{hostname},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(ttl),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	return c.signLeaf(template, pub)
}

// SignControlPlaneCert signs a cert with CN = "controlplane".
// EKU = ClientAuth (CP dials agents) + ServerAuth (CP listens for bootstrap).
func (c *CA) SignControlPlaneCert(pub crypto.PublicKey, ttl time.Duration) ([]byte, error) {
	template := &x509.Certificate{
		Subject:     pkix.Name{CommonName: "controlplane"},
		DNSNames:    []string{"controlplane"},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(ttl),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	return c.signLeaf(template, pub)
}

func (c *CA) signLeaf(template *x509.Certificate, pub crypto.PublicKey) ([]byte, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("pki: generate serial: %w", err)
	}
	template.SerialNumber = serial

	der, err := x509.CreateCertificate(rand.Reader, template, c.Cert, pub, c.Key)
	if err != nil {
		return nil, fmt.Errorf("pki: sign leaf certificate: %w", err)
	}
	return der, nil
}
