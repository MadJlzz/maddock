package kernel

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

func newTestResource(t *testing.T, values map[string]string, cmder util.Commander) *SysctlResource {
	t.Helper()
	return &SysctlResource{
		ResourceName: "test",
		Values:       values,
		Filename:     "99-test.conf",
		dir:          t.TempDir(),
		cmder:        cmder,
	}
}

func TestSysctlResource_Check_AllInSync(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "1\n", ExitCode: 0},
	}}
	sr := newTestResource(t, values, cmder)
	require.NoError(t, os.WriteFile(sr.path(), []byte("net.ipv4.ip_forward = 1\n"), 0644))

	result, err := sr.Check(context.Background())
	assert.NoError(t, err)
	assert.False(t, result.Changed)
	assert.Empty(t, result.Differences)
}

func TestSysctlResource_Check_RuntimeDrift(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "0\n", ExitCode: 0},
	}}
	sr := newTestResource(t, values, cmder)
	require.NoError(t, os.WriteFile(sr.path(), []byte("net.ipv4.ip_forward = 1\n"), 0644))

	result, err := sr.Check(context.Background())
	assert.NoError(t, err)
	assert.True(t, result.Changed)
	assert.Equal(t, []resource.Difference{
		{Attribute: "net.ipv4.ip_forward", Current: "0", Desired: "1"},
	}, result.Differences)
}

func TestSysctlResource_Check_FileMissing(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "1\n", ExitCode: 0},
	}}
	sr := newTestResource(t, values, cmder)

	result, err := sr.Check(context.Background())
	assert.NoError(t, err)
	assert.True(t, result.Changed)
	assert.Len(t, result.Differences, 1)
	assert.Equal(t, "file", result.Differences[0].Attribute)
	assert.Equal(t, "", result.Differences[0].Current)
	assert.Equal(t, "net.ipv4.ip_forward = 1\n", result.Differences[0].Desired)
}

func TestSysctlResource_Check_FileContentDiffers(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "1\n", ExitCode: 0},
	}}
	sr := newTestResource(t, values, cmder)
	require.NoError(t, os.WriteFile(sr.path(), []byte("# stale comment\n"), 0644))

	result, err := sr.Check(context.Background())
	assert.NoError(t, err)
	assert.True(t, result.Changed)
	assert.Len(t, result.Differences, 1)
	assert.Equal(t, "file", result.Differences[0].Attribute)
}

func TestSysctlResource_Check_MultipleKeysSorted(t *testing.T) {
	values := map[string]string{
		"net.ipv6.conf.all.disable_ipv6": "1",
		"net.ipv4.ip_forward":            "1",
	}
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward":            {Output: "0\n", ExitCode: 0},
		"sysctl --values net.ipv6.conf.all.disable_ipv6": {Output: "0\n", ExitCode: 0},
	}}
	sr := newTestResource(t, values, cmder)
	require.NoError(t, os.WriteFile(sr.path(), []byte(sr.renderFile()), 0644))

	result, err := sr.Check(context.Background())
	assert.NoError(t, err)
	assert.True(t, result.Changed)
	require.Len(t, result.Differences, 2)
	assert.Equal(t, "net.ipv4.ip_forward", result.Differences[0].Attribute)
	assert.Equal(t, "net.ipv6.conf.all.disable_ipv6", result.Differences[1].Attribute)
}

func TestSysctlResource_Check_ReadFails(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "", ExitCode: 1},
	}}
	sr := newTestResource(t, values, cmder)

	result, err := sr.Check(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSysctlResource_Apply_AlreadyOk(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "1\n", ExitCode: 0},
	}}
	sr := newTestResource(t, values, cmder)
	require.NoError(t, os.WriteFile(sr.path(), []byte(sr.renderFile()), 0644))

	result, err := sr.Apply(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, resource.Ok, result.Result)
}

func TestSysctlResource_Apply_WritesAndReloads(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	sr := newTestResource(t, values, nil)
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "0\n", ExitCode: 0},
		"sysctl -p " + sr.path():               {ExitCode: 0},
	}}
	sr.cmder = cmder

	result, err := sr.Apply(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, resource.Changed, result.Result)

	written, err := os.ReadFile(sr.path())
	require.NoError(t, err)
	assert.Equal(t, "net.ipv4.ip_forward = 1\n", string(written))
}

func TestSysctlResource_Apply_ReloadFails(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	sr := newTestResource(t, values, nil)
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "0\n", ExitCode: 0},
		"sysctl -p " + sr.path():               {ExitCode: 1},
	}}
	sr.cmder = cmder

	result, err := sr.Apply(context.Background())
	assert.Error(t, err)
	assert.Equal(t, resource.Failed, result.Result)
}

func TestSysctlResource_Apply_WriteFails(t *testing.T) {
	values := map[string]string{"net.ipv4.ip_forward": "1"}
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "0\n", ExitCode: 0},
	}}
	sr := newTestResource(t, values, cmder)
	sr.dir = filepath.Join(sr.dir, "does-not-exist")

	result, err := sr.Apply(context.Background())
	assert.Error(t, err)
	assert.Equal(t, resource.Failed, result.Result)
}

func TestSysctlResource_RenderFile_Deterministic(t *testing.T) {
	values := map[string]string{
		"c.b.a": "3",
		"a.b.c": "1",
		"b.c.a": "2",
	}
	sr := &SysctlResource{Values: values}
	expected := "a.b.c = 1\nb.c.a = 2\nc.b.a = 3\n"
	assert.Equal(t, expected, sr.renderFile())
}
