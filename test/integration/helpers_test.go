//go:build integration

package integration

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../")
	require.NoError(t, err)
	return root
}

func manifestFiles(root string) []testcontainers.ContainerFile {
	manifests := []string{
		"install-pkg.yaml",
		"remove-pkg.yaml",
		"start-svc.yaml",
		"stop-svc.yaml",
	}
	files := make([]testcontainers.ContainerFile, len(manifests))
	for i, m := range manifests {
		files[i] = testcontainers.ContainerFile{
			HostFilePath:      filepath.Join(root, "test", "testdata", m),
			ContainerFilePath: "/etc/maddock/" + m,
			FileMode:          0o644,
		}
	}
	return files
}

func buildAgent(t *testing.T, ctx context.Context, containerfile string) testcontainers.Container {
	t.Helper()
	root := projectRoot(t)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    root,
				Dockerfile: containerfile,
			},
			Files:      manifestFiles(root),
			Entrypoint: []string{"sleep", "infinity"},
			WaitingFor: wait.ForExec([]string{"true"}).WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	return container
}

func buildSystemdAgent(t *testing.T, ctx context.Context, containerfile string) testcontainers.Container {
	t.Helper()
	root := projectRoot(t)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    root,
				Dockerfile: containerfile,
			},
			Files:      manifestFiles(root),
			Privileged: true,
			WaitingFor: wait.ForExec([]string{"systemctl", "is-system-running", "--wait"}).
				WithStartupTimeout(60 * time.Second),
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
