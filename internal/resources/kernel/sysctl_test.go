package kernel

import (
	"context"
	"testing"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestSysctlResource_Check(t *testing.T) {
	tests := []struct {
		name         string
		desiredValue string
		currentValue string
		wantChanged  bool
		wantDiffs    int
	}{
		{"already correct", "1", "1\n", false, 0},
		{"needs update", "1", "0\n", true, 1},
		{"trailing whitespace trimmed", "1", "  1  \n", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmder := util.MockCommander{Commands: map[string]util.MockCommand{
				"sysctl --values net.ipv4.ip_forward": {Output: tt.currentValue, ExitCode: 0},
			}}
			sr := SysctlResource{Key: "net.ipv4.ip_forward", DesiredValue: tt.desiredValue, cmder: cmder}

			result, err := sr.Check(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.wantChanged, result.Changed)
			assert.Len(t, result.Differences, tt.wantDiffs)
		})
	}
}

func TestSysctlResource_Check_Diff(t *testing.T) {
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "0\n", ExitCode: 0},
	}}
	sr := SysctlResource{Key: "net.ipv4.ip_forward", DesiredValue: "1", cmder: cmder}

	result, err := sr.Check(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, resource.Difference{
		Attribute: "value",
		Current:   "0",
		Desired:   "1",
	}, result.Differences[0])
}

func TestSysctlResource_Check_ReadFails(t *testing.T) {
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "", ExitCode: 1},
	}}
	sr := SysctlResource{Key: "net.ipv4.ip_forward", DesiredValue: "1", cmder: cmder}

	result, err := sr.Check(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSysctlResource_Apply(t *testing.T) {
	tests := []struct {
		name         string
		desiredValue string
		currentValue string
		setExitCode  int
		wantResult   resource.State
		wantErr      bool
	}{
		{"changes value", "1", "0\n", 0, resource.Changed, false},
		{"already correct", "1", "1\n", 0, resource.Ok, false},
		{"set fails", "1", "0\n", 1, resource.Failed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmder := util.MockCommander{Commands: map[string]util.MockCommand{
				"sysctl --values net.ipv4.ip_forward":              {Output: tt.currentValue, ExitCode: 0},
				"sysctl --write net.ipv4.ip_forward=" + tt.desiredValue: {ExitCode: tt.setExitCode},
			}}
			sr := SysctlResource{Key: "net.ipv4.ip_forward", DesiredValue: tt.desiredValue, cmder: cmder}

			result, err := sr.Apply(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantResult, result.Result)
		})
	}
}

func TestSysctlResource_Apply_ReadFails(t *testing.T) {
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"sysctl --values net.ipv4.ip_forward": {Output: "", ExitCode: 1},
	}}
	sr := SysctlResource{Key: "net.ipv4.ip_forward", DesiredValue: "1", cmder: cmder}

	result, err := sr.Apply(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
}
