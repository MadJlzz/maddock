package main

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"

	"github.com/MadJlzz/maddock/internal/pki"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cp")

	require.NoError(t, runInit(dir))

	cases := []struct {
		name     string
		wantMode os.FileMode
	}{
		{pki.CertificateAuthorityCertName, 0644},
		{pki.CertificateAuthorityKeyName, 0600},
		{ControlPlaneCertName, 0644},
		{ControlPlaneKeyName, 0600},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			info, err := os.Stat(filepath.Join(dir, tc.name))
			require.NoError(t, err)
			assert.Equal(t, tc.wantMode, info.Mode().Perm())
		})
	}

	dirInfo, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())

	caCertPEM, err := os.ReadFile(filepath.Join(dir, pki.CertificateAuthorityCertName))
	require.NoError(t, err)
	caCert, err := pki.DecodeCertPEM(caCertPEM)
	require.NoError(t, err)

	cpCertPEM, err := os.ReadFile(filepath.Join(dir, ControlPlaneCertName))
	require.NoError(t, err)
	cpCert, err := pki.DecodeCertPEM(cpCertPEM)
	require.NoError(t, err)

	pool := x509.NewCertPool()
	pool.AddCert(caCert)
	_, err = cpCert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	assert.NoError(t, err)
}

func TestRunInit_RefusesIfAlreadyInitialized(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cp")

	require.NoError(t, runInit(dir))

	err := runInit(dir)
	assert.ErrorContains(t, err, "already initialized")
}
