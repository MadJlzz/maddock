package transport

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"testing"
	"time"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/pki"
	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/transport/proto"

	// Register the file resource with the global registry.
	_ "github.com/MadJlzz/maddock/internal/resources/file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// testCerts holds a freshly minted CA and the server/client material derived
// from it, mirroring the real mTLS setup: the agent presents an agent cert,
// the control plane presents a control-plane cert, both trust the same CA.
type testCerts struct {
	serverCreds credentials.TransportCredentials
	caPool      *x509.CertPool
	cpCert      tls.Certificate
	agentName   string
}

// newTestCerts builds an in-memory PKI for a test: a CA, an agent server cert
// (SAN = agentName), and a control-plane client cert.
func newTestCerts(t *testing.T) testCerts {
	t.Helper()

	ca, err := pki.GenerateCA("test-ca")
	require.NoError(t, err)

	const agentName = "agent-test"
	agentKey, err := pki.GenerateKey()
	require.NoError(t, err)
	agentDER, err := ca.SignAgentCert(agentName, agentKey.Public(), time.Hour)
	require.NoError(t, err)
	agentCert := mustKeyPair(t, agentDER, agentKey)

	cpKey, err := pki.GenerateKey()
	require.NoError(t, err)
	cpDER, err := ca.SignControlPlaneCert(cpKey.Public(), time.Hour)
	require.NoError(t, err)
	cpCert := mustKeyPair(t, cpDER, cpKey)

	caPool := x509.NewCertPool()
	caPool.AddCert(ca.Cert)

	serverCfg := &tls.Config{
		Certificates: []tls.Certificate{agentCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS13,
	}

	return testCerts{
		serverCreds: credentials.NewTLS(serverCfg),
		caPool:      caPool,
		cpCert:      cpCert,
		agentName:   agentName,
	}
}

func mustKeyPair(t *testing.T, der []byte, key *ecdsa.PrivateKey) tls.Certificate {
	t.Helper()
	keyPEM, err := pki.EncodeKeyPEM(key)
	require.NoError(t, err)
	cert, err := tls.X509KeyPair(pki.EncodeCertPEM(der), keyPEM)
	require.NoError(t, err)
	return cert
}

// clientConfig returns a control-plane client tls.Config verifying the agent
// cert against serverName.
func (tc testCerts) clientConfig(serverName string) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{tc.cpCert},
		RootCAs:      tc.caPool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}
}

// startServer brings up a TLS gRPC agent server on a random local port and
// returns its address. It is stopped when the test ends.
func startServer(t *testing.T, creds credentials.TransportCredentials) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer(grpc.Creds(creds))
	proto.RegisterAgentServiceServer(grpcServer, &AgentServer{Version: "test"})

	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(func() { grpcServer.Stop() })

	return lis.Addr().String()
}

// setupAgentServer starts a TLS gRPC server on a random local port and returns
// a connected mTLS client. Both are cleaned up when the test ends.
func setupAgentServer(t *testing.T) *Client {
	t.Helper()

	tc := newTestCerts(t)
	addr := startServer(t, tc.serverCreds)

	client, err := NewClient(addr, tc.clientConfig(tc.agentName))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	return client
}

// currentOwnership returns the current user's username and primary group name.
// Used to build file-resource attrs that won't trigger ownership changes.
func currentOwnership(t *testing.T) (string, string) {
	t.Helper()
	u, err := user.Current()
	require.NoError(t, err)
	g, err := user.LookupGroupId(u.Gid)
	require.NoError(t, err)
	return u.Username, g.Name
}

func TestPing(t *testing.T) {
	client := setupAgentServer(t)

	resp, err := client.Ping(context.Background())
	require.NoError(t, err)

	hostname, _ := os.Hostname()
	assert.Equal(t, hostname, resp.Hostname)
	assert.Equal(t, "test", resp.AgentVersion)
}

func TestPing_RejectsSANMismatch(t *testing.T) {
	tc := newTestCerts(t)
	addr := startServer(t, tc.serverCreds)

	// ServerName does not match the agent cert SAN, so verification must fail.
	client, err := NewClient(addr, tc.clientConfig("wrong-host"))
	require.NoError(t, err) // dial is lazy; the error surfaces on the first RPC
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.Ping(context.Background())
	require.Error(t, err)
}

func TestPing_RejectsClientWithoutCert(t *testing.T) {
	tc := newTestCerts(t)
	addr := startServer(t, tc.serverCreds)

	// Client trusts the CA but presents no client cert; the agent requires one.
	cfg := &tls.Config{
		RootCAs:    tc.caPool,
		ServerName: tc.agentName,
		MinVersion: tls.VersionTLS13,
	}
	client, err := NewClient(addr, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.Ping(context.Background())
	require.Error(t, err)
}

func TestApplyCatalog_CreatesFile(t *testing.T) {
	client := setupAgentServer(t)
	owner, group := currentOwnership(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	rc := &catalog.RawCatalog{
		Name: "transport-test",
		Resources: []catalog.RawResource{{
			Type: "file",
			Name: path,
			Attributes: map[string]any{
				"content": "hello from grpc",
				"owner":   owner,
				"group":   group,
				"mode":    "0644",
			},
		}},
	}

	r, err := client.ApplyCatalog(context.Background(), rc, false)
	require.NoError(t, err)
	require.Len(t, r.ResourceReports, 1)
	assert.Equal(t, resource.Changed, r.ResourceReports[0].State)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello from grpc", string(data))
}

func TestApplyCatalog_DryRunDoesNotCreateFile(t *testing.T) {
	client := setupAgentServer(t)
	owner, group := currentOwnership(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	rc := &catalog.RawCatalog{
		Name: "transport-test-dry",
		Resources: []catalog.RawResource{{
			Type: "file",
			Name: path,
			Attributes: map[string]any{
				"content": "should not exist",
				"owner":   owner,
				"group":   group,
				"mode":    "0644",
			},
		}},
	}

	r, err := client.ApplyCatalog(context.Background(), rc, true)
	require.NoError(t, err)
	require.Len(t, r.ResourceReports, 1)
	assert.Equal(t, resource.Skipped, r.ResourceReports[0].State)

	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestApplyCatalog_Idempotent(t *testing.T) {
	client := setupAgentServer(t)
	owner, group := currentOwnership(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	rc := &catalog.RawCatalog{
		Name: "transport-test-idempotent",
		Resources: []catalog.RawResource{{
			Type: "file",
			Name: path,
			Attributes: map[string]any{
				"content": "idempotent content",
				"owner":   owner,
				"group":   group,
				"mode":    "0644",
			},
		}},
	}

	// First apply: changes expected.
	r, err := client.ApplyCatalog(context.Background(), rc, false)
	require.NoError(t, err)
	assert.Equal(t, resource.Changed, r.ResourceReports[0].State)

	// Second apply: everything should already be in desired state.
	r, err = client.ApplyCatalog(context.Background(), rc, false)
	require.NoError(t, err)
	assert.Equal(t, resource.Ok, r.ResourceReports[0].State)
}
