package command

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_RequiresCommand(t *testing.T) {
	_, err := resource.Parse("command", "missing", map[string]any{})
	assert.Error(t, err)
}

func TestParse_CommandMustBeString(t *testing.T) {
	_, err := resource.Parse("command", "bad", map[string]any{"command": 42})
	assert.Error(t, err)
}

func TestParse_AcceptsGuards(t *testing.T) {
	r, err := resource.Parse("command", "guarded", map[string]any{
		"command": "echo hi",
		"creates": "/tmp/x",
		"unless":  "test -f /tmp/x",
		"onlyif":  "true",
	})
	require.NoError(t, err)

	cr := r.(*CommandResource)
	assert.Equal(t, "echo hi", cr.command)
	assert.Equal(t, "/tmp/x", cr.creates)
	assert.Equal(t, "test -f /tmp/x", cr.unless)
	assert.Equal(t, "true", cr.onlyif)
}

func TestCheck_NoGuardsAlwaysChanges(t *testing.T) {
	cr := &CommandResource{command: "echo hi", cmder: util.MockCommander{}}
	result, err := cr.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Changed)
	require.Len(t, result.Differences, 1)
	assert.Equal(t, "command", result.Differences[0].Attribute)
	assert.Equal(t, "echo hi", result.Differences[0].Desired)
}

func TestCheck_CreatesExisting(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	require.NoError(t, os.WriteFile(marker, []byte("x"), 0644))

	cr := &CommandResource{command: "touch " + marker, creates: marker, cmder: util.MockCommander{}}
	result, err := cr.Check(context.Background())
	require.NoError(t, err)
	assert.False(t, result.Changed)
}

func TestCheck_CreatesMissing(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "does-not-exist")

	cr := &CommandResource{command: "touch " + marker, creates: marker, cmder: util.MockCommander{}}
	result, err := cr.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Changed)
}

func TestCheck_UnlessExitZeroSkips(t *testing.T) {
	cr := &CommandResource{
		command: "echo run",
		unless:  "test -f /marker",
		cmder: util.MockCommander{Commands: map[string]util.MockCommand{
			"/bin/sh -c test -f /marker": {ExitCode: 0},
		}},
	}
	result, err := cr.Check(context.Background())
	require.NoError(t, err)
	assert.False(t, result.Changed)
}

func TestCheck_UnlessExitNonZeroRuns(t *testing.T) {
	cr := &CommandResource{
		command: "echo run",
		unless:  "test -f /marker",
		cmder: util.MockCommander{Commands: map[string]util.MockCommand{
			"/bin/sh -c test -f /marker": {ExitCode: 1},
		}},
	}
	result, err := cr.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Changed)
}

func TestCheck_OnlyifExitZeroRuns(t *testing.T) {
	cr := &CommandResource{
		command: "echo run",
		onlyif:  "test -f /marker",
		cmder: util.MockCommander{Commands: map[string]util.MockCommand{
			"/bin/sh -c test -f /marker": {ExitCode: 0},
		}},
	}
	result, err := cr.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Changed)
}

func TestCheck_OnlyifExitNonZeroSkips(t *testing.T) {
	cr := &CommandResource{
		command: "echo run",
		onlyif:  "test -f /marker",
		cmder: util.MockCommander{Commands: map[string]util.MockCommand{
			"/bin/sh -c test -f /marker": {ExitCode: 1},
		}},
	}
	result, err := cr.Check(context.Background())
	require.NoError(t, err)
	assert.False(t, result.Changed)
}

func TestCheck_GuardsAreAndCombined(t *testing.T) {
	// unless says "run" (non-zero) but onlyif says "skip" (non-zero).
	// AND-combination means the command must NOT run.
	cr := &CommandResource{
		command: "echo run",
		unless:  "false",
		onlyif:  "false",
		cmder: util.MockCommander{Commands: map[string]util.MockCommand{
			"/bin/sh -c false": {ExitCode: 1},
		}},
	}
	result, err := cr.Check(context.Background())
	require.NoError(t, err)
	assert.False(t, result.Changed)
}

func TestApply_SuccessReturnsChanged(t *testing.T) {
	cr := &CommandResource{
		command: "echo hi",
		cmder: util.MockCommander{Commands: map[string]util.MockCommand{
			"/bin/sh -c echo hi": {ExitCode: 0},
		}},
	}
	result, err := cr.Apply(context.Background())
	require.NoError(t, err)
	assert.Equal(t, resource.Changed, result.Result)
}

func TestApply_NonZeroExitReturnsFailed(t *testing.T) {
	cr := &CommandResource{
		command: "broken",
		cmder: util.MockCommander{Commands: map[string]util.MockCommand{
			"/bin/sh -c broken": {ExitCode: 1, Output: "some stderr"},
		}},
	}
	result, err := cr.Apply(context.Background())
	require.Error(t, err)
	assert.Equal(t, resource.Failed, result.Result)
}
