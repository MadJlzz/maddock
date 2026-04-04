//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFedora_StartService(t *testing.T) {
	ctx := context.Background()
	container := buildSystemdAgent(t, ctx, "test/Containerfile.fedora-systemd")
	defer func() { _ = container.Terminate(ctx) }()

	// First run: service should be started and enabled
	code, output := execAgent(t, ctx, container, "apply", "/etc/maddock/start-svc.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "CHANGED")
	t.Log(output)

	// Second run: idempotency
	code, output = execAgent(t, ctx, container, "apply", "/etc/maddock/start-svc.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "OK")
	assert.NotContains(t, output, "CHANGED")
	t.Log(output)
}

func TestFedora_StopService(t *testing.T) {
	ctx := context.Background()
	container := buildSystemdAgent(t, ctx, "test/Containerfile.fedora-systemd")
	defer func() { _ = container.Terminate(ctx) }()

	// First start the service
	code, _ := execAgent(t, ctx, container, "apply", "/etc/maddock/start-svc.yaml")
	assert.Equal(t, 0, code)

	// Now stop and disable it
	code, output := execAgent(t, ctx, container, "apply", "/etc/maddock/stop-svc.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "CHANGED")
	t.Log(output)

	// Idempotency
	code, output = execAgent(t, ctx, container, "apply", "/etc/maddock/stop-svc.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "OK")
	assert.NotContains(t, output, "CHANGED")
	t.Log(output)
}

func TestFedora_ServiceDryRun(t *testing.T) {
	ctx := context.Background()
	container := buildSystemdAgent(t, ctx, "test/Containerfile.fedora-systemd")
	defer func() { _ = container.Terminate(ctx) }()

	// Dry-run should not start the service
	code, output := execAgent(t, ctx, container, "apply", "--dry-run", "/etc/maddock/start-svc.yaml")
	assert.Equal(t, 3, code)
	assert.Contains(t, output, "SKIPPED")
	t.Log(output)

	// Verify the service is NOT running
	verifyCode, _, err := container.Exec(ctx, []string{"systemctl", "is-active", "maddock-test"})
	require.NoError(t, err)
	assert.NotEqual(t, 0, verifyCode, "service should not be running after dry-run")
}
