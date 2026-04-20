package transport

import (
	"context"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/transport/proto"

	// Register the file resource with the global registry.
	_ "github.com/MadJlzz/maddock/internal/resources/file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// setupServer starts a gRPC server on a random local port and returns a
// connected client. Both are cleaned up when the test ends.
func setupServer(t *testing.T) *Client {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	proto.RegisterAgentServiceServer(grpcServer, &Server{Version: "test"})

	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(func() { grpcServer.Stop() })

	client, err := NewClient(lis.Addr().String())
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
	client := setupServer(t)

	resp, err := client.Ping(context.Background())
	require.NoError(t, err)

	hostname, _ := os.Hostname()
	assert.Equal(t, hostname, resp.Hostname)
	assert.Equal(t, "test", resp.AgentVersion)
}

func TestApplyCatalog_CreatesFile(t *testing.T) {
	client := setupServer(t)
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
	client := setupServer(t)
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
	client := setupServer(t)
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
