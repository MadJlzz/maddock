//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInstallPackage(t *testing.T, containerfile string) {
	ctx := context.Background()
	container := buildAgent(t, ctx, containerfile)
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

func testRemovePackage(t *testing.T, containerfile string) {
	ctx := context.Background()
	container := buildAgent(t, ctx, containerfile)
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

func testDryRunPackage(t *testing.T, containerfile string, queryCmd []string) {
	ctx := context.Background()
	container := buildAgent(t, ctx, containerfile)
	defer func() { _ = container.Terminate(ctx) }()

	// Dry-run should report SKIPPED, not install anything (exit code 3 = changes pending)
	code, output := execAgent(t, ctx, container, "apply", "--dry-run", "/etc/maddock/install-pkg.yaml")
	assert.Equal(t, 3, code)
	assert.Contains(t, output, "SKIPPED")
	t.Log(output)

	// Verify the package was NOT actually installed
	verifyCode, _, err := container.Exec(ctx, queryCmd)
	require.NoError(t, err)
	assert.NotEqual(t, 0, verifyCode, "package should not be installed after dry-run")
}

func TestFedora_InstallPackage(t *testing.T) {
	testInstallPackage(t, "Containerfile")
}

func TestFedora_RemovePackage(t *testing.T) {
	testRemovePackage(t, "Containerfile")
}

func TestFedora_DryRunPackage(t *testing.T) {
	testDryRunPackage(t, "Containerfile", []string{"rpm", "--query", "htop"})
}

func TestUbuntu_InstallPackage(t *testing.T) {
	testInstallPackage(t, "test/Containerfile.ubuntu")
}

func TestUbuntu_RemovePackage(t *testing.T) {
	testRemovePackage(t, "test/Containerfile.ubuntu")
}

func TestUbuntu_DryRunPackage(t *testing.T) {
	testDryRunPackage(t, "test/Containerfile.ubuntu", []string{"dpkg-query", "--status", "htop"})
}
