package pki

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"time"
)

type CA struct {
	Cert *x509.Certificate
	Key  *ecdsa.PrivateKey
}

func GenerateCA(commonName string) (*CA, error) {

	panic("implement me")
}

func Load(dir string) (*CA, error) {
	panic("implement me")
}

func (c *CA) Save(dir string) error {
	panic("implement me")
}

// SignAgentCert signs a cert with CN = hostname, SAN DNS = hostname.
// EKU includes both ServerAuth (the agent listens) and ClientAuth
// (so the same cert can be reused if agents ever initiate calls).
// Returns DER-encoded certificate bytes.
func (*CA) SignAgentCert(hostname string, pub crypto.PublicKey, ttl time.Duration) ([]byte, error) {
	panic("implement me")
}

// SignControlPlaneCert signs a cert with CN = "controlplane".
// EKU = ClientAuth (CP dials agents) + ServerAuth (CP listens for bootstrap).
func (*CA) SignControlPlaneCert(pub crypto.PublicKey, ttl time.Duration) ([]byte, error) {
	panic("implement me")
}
