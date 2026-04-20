//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRunCommand(t *testing.T, containerfile string) {
	ctx := context.Background()
	container := buildAgent(t, ctx, containerfile)
	defer func() { _ = container.Terminate(ctx) }()

	// First run: command should execute and create the marker file.
	code, output := execAgent(t, ctx, container, "apply", "/etc/maddock/run-command.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "CHANGED")
	t.Log(output)

	// Verify the marker file was created with expected content.
	verifyCode, _, err := container.Exec(ctx, []string{
		"sh", "-c", `[ "$(cat /tmp/maddock-cmd-test)" = "hello" ]`,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, verifyCode, "marker file should contain 'hello'")

	// Second run: the creates guard should skip execution — OK, not CHANGED.
	code, output = execAgent(t, ctx, container, "apply", "/etc/maddock/run-command.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "OK")
	assert.NotContains(t, output, "CHANGED")
	t.Log(output)
}

func testDryRunCommand(t *testing.T, containerfile string) {
	ctx := context.Background()
	container := buildAgent(t, ctx, containerfile)
	defer func() { _ = container.Terminate(ctx) }()

	// Dry-run should report SKIPPED and NOT execute the command.
	code, output := execAgent(t, ctx, container, "apply", "--dry-run", "/etc/maddock/run-command.yaml")
	assert.Equal(t, 3, code)
	assert.Contains(t, output, "SKIPPED")
	t.Log(output)

	// Verify the marker file was NOT created.
	verifyCode, _, err := container.Exec(ctx, []string{"test", "-f", "/tmp/maddock-cmd-test"})
	require.NoError(t, err)
	assert.NotEqual(t, 0, verifyCode, "marker file should not exist after dry-run")
}

func TestFedora_RunCommand(t *testing.T) {
	testRunCommand(t, "Containerfile")
}

func TestFedora_DryRunCommand(t *testing.T) {
	testDryRunCommand(t, "Containerfile")
}

func TestUbuntu_RunCommand(t *testing.T) {
	testRunCommand(t, "test/Containerfile.ubuntu")
}

func TestUbuntu_DryRunCommand(t *testing.T) {
	testDryRunCommand(t, "test/Containerfile.ubuntu")
}
