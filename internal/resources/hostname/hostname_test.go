package hostname

import (
	"context"
	"testing"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestHostnameResource_Check(t *testing.T) {
	tests := []struct {
		name            string
		desiredHostname string
		currentHostname string
		wantChanged     bool
		wantDiffs       int
	}{
		{"already correct", "web1", "web1\n", false, 0},
		{"needs update", "web1", "db1\n", true, 1},
		{"trailing whitespace trimmed", "web1", "  web1  \n", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmder := util.MockCommander{Commands: map[string]util.MockCommand{
				"hostnamectl --static": {Output: tt.currentHostname, ExitCode: 0},
			}}
			hr := HostnameResource{name: tt.desiredHostname, cmder: cmder}

			result, err := hr.Check(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.wantChanged, result.Changed)
			assert.Len(t, result.Differences, tt.wantDiffs)
		})
	}
}

func TestHostnameResource_Check_Diff(t *testing.T) {
	cmder := util.MockCommander{Commands: map[string]util.MockCommand{
		"hostnamectl --static": {Output: "old-host\n", ExitCode: 0},
	}}
	hr := HostnameResource{name: "new-host", cmder: cmder}

	result, err := hr.Check(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, resource.Difference{
		Attribute: "name",
		Current:   "old-host",
		Desired:   "new-host",
	}, result.Differences[0])
}

func TestHostnameResource_Apply(t *testing.T) {
	tests := []struct {
		name            string
		desiredHostname string
		currentHostname string
		setExitCode     int
		wantResult      resource.State
		wantErr         bool
	}{
		{"changes hostname", "web1", "db1\n", 0, resource.Changed, false},
		{"already correct", "web1", "web1\n", 0, resource.Ok, false},
		{"set fails", "web1", "db1\n", 1, resource.Failed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmder := util.MockCommander{Commands: map[string]util.MockCommand{
				"hostnamectl --static":                       {Output: tt.currentHostname, ExitCode: 0},
				"hostnamectl hostname " + tt.desiredHostname: {ExitCode: tt.setExitCode},
			}}
			hr := HostnameResource{name: tt.desiredHostname, cmder: cmder}

			result, err := hr.Apply(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantResult, result.Result)
		})
	}
}
