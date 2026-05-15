package pki

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCA(t *testing.T) {
	ca, err := GenerateCA("foo.bar")
	assert.NoError(t, err)

	assert.True(t, ca.Cert.IsCA)
	assert.True(t, ca.Cert.BasicConstraintsValid)
	assert.Equal(t, "foo.bar", ca.Cert.Subject.CommonName)
	assert.True(t, ca.Cert.NotAfter.After(time.Now()))
	assert.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, ca.Cert.KeyUsage)
	assert.NoError(t, ca.Cert.CheckSignatureFrom(ca.Cert))
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()

	original, err := GenerateCA("test.local")
	require.NoError(t, err)

	require.NoError(t, original.Save(dir))

	certStat, err := os.Stat(filepath.Join(dir, CertificateAuthorityCertName))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), certStat.Mode().Perm())

	keyStat, err := os.Stat(filepath.Join(dir, CertificateAuthorityKeyName))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), keyStat.Mode().Perm())

	loaded, err := Load(dir)
	require.NoError(t, err)

	assert.True(t, loaded.Cert.Equal(original.Cert))
	assert.True(t, loaded.Key.Equal(original.Key))
}

func TestSignAgentCert(t *testing.T) {
	ca, err := GenerateCA("test.local")
	require.NoError(t, err)

	agentKey, err := GenerateKey()
	require.NoError(t, err)

	der, err := ca.SignAgentCert("agent01.example.com", agentKey.Public(), time.Hour)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	assert.Equal(t, "agent01.example.com", cert.Subject.CommonName)
	assert.Contains(t, cert.DNSNames, "agent01.example.com")
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	assert.False(t, cert.IsCA)

	pool := x509.NewCertPool()
	pool.AddCert(ca.Cert)
	_, err = cert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	assert.NoError(t, err)
}

func TestSignControlPlaneCert(t *testing.T) {
	ca, err := GenerateCA("test.local")
	require.NoError(t, err)

	cpKey, err := GenerateKey()
	require.NoError(t, err)

	der, err := ca.SignControlPlaneCert(cpKey.Public(), time.Hour)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	assert.Equal(t, "controlplane", cert.Subject.CommonName)
	assert.Contains(t, cert.DNSNames, "controlplane")
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	assert.False(t, cert.IsCA)

	pool := x509.NewCertPool()
	pool.AddCert(ca.Cert)
	_, err = cert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	assert.NoError(t, err)
}

func TestSignAgentCert_RejectsExpired(t *testing.T) {
	ca, err := GenerateCA("test.local")
	require.NoError(t, err)

	agentKey, err := GenerateKey()
	require.NoError(t, err)

	der, err := ca.SignAgentCert("agent01.example.com", agentKey.Public(), -time.Hour)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	pool := x509.NewCertPool()
	pool.AddCert(ca.Cert)
	_, err = cert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	assert.Error(t, err)
}
