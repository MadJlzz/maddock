//go:build integration

package integration

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func buildAgent(t *testing.T, ctx context.Context) testcontainers.Container {
	t.Helper()
	projectRoot, err := filepath.Abs("../../")
	require.NoError(t, err)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    projectRoot,
				Dockerfile: "Containerfile",
			},
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      filepath.Join(projectRoot, "test", "testdata", "install-pkg.yaml"),
					ContainerFilePath: "/etc/maddock/install-pkg.yaml",
					FileMode:          0o644,
				},
				{
					HostFilePath:      filepath.Join(projectRoot, "test", "testdata", "remove-pkg.yaml"),
					ContainerFilePath: "/etc/maddock/remove-pkg.yaml",
					FileMode:          0o644,
				},
			},
			Entrypoint: []string{"sleep", "infinity"},
			WaitingFor: wait.ForExec([]string{"true"}).WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	return container
}

func execAgent(t *testing.T, ctx context.Context, container testcontainers.Container, args ...string) (int, string) {
	t.Helper()
	cmd := append([]string{"maddock-agent"}, args...)
	code, reader, err := container.Exec(ctx, cmd)
	require.NoError(t, err)
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, reader)
	require.NoError(t, err)
	return code, stdout.String()
}

func TestAgentInstallPackage(t *testing.T) {
	ctx := context.Background()
	container := buildAgent(t, ctx)
	defer func() { _ = container.Terminate(ctx) }()

	// First run: package should be installed
	code, output := execAgent(t, ctx, container, "apply", "/etc/maddock/install-pkg.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "CHANGED")
	t.Log(output)

	// Second run: idempotency — nothing should change
	code, output = execAgent(t, ctx, container, "apply", "/etc/maddock/install-pkg.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "OK")
	assert.NotContains(t, output, "CHANGED")
	t.Log(output)
}

func TestAgentRemovePackage(t *testing.T) {
	ctx := context.Background()
	container := buildAgent(t, ctx)
	defer func() { _ = container.Terminate(ctx) }()

	// First install the package
	code, _ := execAgent(t, ctx, container, "apply", "/etc/maddock/install-pkg.yaml")
	assert.Equal(t, 0, code)

	// Now remove it
	code, output := execAgent(t, ctx, container, "apply", "/etc/maddock/remove-pkg.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "CHANGED")
	t.Log(output)

	// Idempotency: removing again should be OK
	code, output = execAgent(t, ctx, container, "apply", "/etc/maddock/remove-pkg.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "OK")
	assert.NotContains(t, output, "CHANGED")
	t.Log(output)
}

func TestAgentDryRun(t *testing.T) {
	ctx := context.Background()
	container := buildAgent(t, ctx)
	defer func() { _ = container.Terminate(ctx) }()

	// Dry-run should report SKIPPED, not install anything
	code, output := execAgent(t, ctx, container, "apply", "--dry-run", "/etc/maddock/install-pkg.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "SKIPPED")
	t.Log(output)

	// Verify the package was NOT actually installed
	verifyCode, _, err := container.Exec(ctx, []string{"rpm", "--query", "htop"})
	require.NoError(t, err)
	assert.NotEqual(t, 0, verifyCode, "package should not be installed after dry-run")
}
