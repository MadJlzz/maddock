//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCreateFile(t *testing.T, containerfile string) {
	ctx := context.Background()
	container := buildAgent(t, ctx, containerfile)
	defer func() { _ = container.Terminate(ctx) }()

	// First run: file should be created
	code, output := execAgent(t, ctx, container, "apply", "/etc/maddock/create-file.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "CHANGED")
	t.Log(output)

	// Verify file content
	verifyCode, _, err := container.Exec(ctx, []string{
		"sh", "-c", `[ "$(cat /tmp/maddock-test.txt)" = "hello from maddock" ]`,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, verifyCode, "file content should match")

	// Verify file mode
	verifyCode, _, err = container.Exec(ctx, []string{
		"sh", "-c", `[ "$(stat -c %a /tmp/maddock-test.txt)" = "644" ]`,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, verifyCode, "file mode should be 644")

	// Verify ownership
	verifyCode, _, err = container.Exec(ctx, []string{
		"sh", "-c", `[ "$(stat -c %U:%G /tmp/maddock-test.txt)" = "root:root" ]`,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, verifyCode, "file should be owned by root:root")

	// Second run: idempotency
	code, output = execAgent(t, ctx, container, "apply", "/etc/maddock/create-file.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "OK")
	assert.NotContains(t, output, "CHANGED")
	t.Log(output)
}

func testDryRunFile(t *testing.T, containerfile string) {
	ctx := context.Background()
	container := buildAgent(t, ctx, containerfile)
	defer func() { _ = container.Terminate(ctx) }()

	// Dry-run should not create the file
	code, output := execAgent(t, ctx, container, "apply", "--dry-run", "/etc/maddock/create-file.yaml")
	assert.Equal(t, 3, code)
	assert.Contains(t, output, "SKIPPED")
	t.Log(output)

	// Verify the file was NOT created
	verifyCode, _, err := container.Exec(ctx, []string{"test", "-f", "/tmp/maddock-test.txt"})
	require.NoError(t, err)
	assert.NotEqual(t, 0, verifyCode, "file should not exist after dry-run")
}

func testCreateFileFromTemplate(t *testing.T, containerfile string) {
	ctx := context.Background()
	container := buildAgent(t, ctx, containerfile)
	defer func() { _ = container.Terminate(ctx) }()

	// First run: file should be created from rendered template
	code, output := execAgent(t, ctx, container, "apply", "/etc/maddock/create-file-template.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "CHANGED")
	t.Log(output)

	// Verify rendered content
	verifyCode, _, err := container.Exec(ctx, []string{
		"sh", "-c", `grep -q "worker_connections 1024;" /tmp/maddock-nginx.conf`,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, verifyCode, "rendered template should contain worker_connections")

	verifyCode, _, err = container.Exec(ctx, []string{
		"sh", "-c", `grep -q "server_name localhost;" /tmp/maddock-nginx.conf`,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, verifyCode, "rendered template should contain server_name")

	// Second run: idempotency
	code, output = execAgent(t, ctx, container, "apply", "/etc/maddock/create-file-template.yaml")
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "OK")
	assert.NotContains(t, output, "CHANGED")
	t.Log(output)
}

func TestFedora_CreateFile(t *testing.T) {
	testCreateFile(t, "Containerfile")
}

func TestFedora_DryRunFile(t *testing.T) {
	testDryRunFile(t, "Containerfile")
}

func TestUbuntu_CreateFile(t *testing.T) {
	testCreateFile(t, "test/Containerfile.ubuntu")
}

func TestUbuntu_DryRunFile(t *testing.T) {
	testDryRunFile(t, "test/Containerfile.ubuntu")
}

func TestFedora_CreateFileFromTemplate(t *testing.T) {
	testCreateFileFromTemplate(t, "Containerfile")
}

func TestUbuntu_CreateFileFromTemplate(t *testing.T) {
	testCreateFileFromTemplate(t, "test/Containerfile.ubuntu")
}
