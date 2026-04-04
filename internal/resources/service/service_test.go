package service

import (
	"context"
	"testing"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestServiceResource_Check(t *testing.T) {
	tests := []struct {
		name           string
		desiredState   string
		desiredEnabled bool
		isActive       string
		isEnabled      string
		wantChanged    bool
		wantDiffs      int
	}{
		{"running+enabled, want running+enabled", "running", true, "active", "enabled", false, 0},
		{"stopped+disabled, want stopped+disabled", "stopped", false, "inactive", "disabled", false, 0},
		{"stopped+disabled, want running+enabled", "running", true, "inactive", "disabled", true, 2},
		{"running+enabled, want stopped+disabled", "stopped", false, "active", "enabled", true, 2},
		{"stopped+disabled, want running+disabled", "running", false, "inactive", "disabled", true, 1},
		{"running+enabled, want running+disabled", "running", false, "active", "enabled", true, 1},
		{"stopped+disabled, want stopped+enabled", "stopped", true, "inactive", "disabled", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmder := util.MockCommander{Commands: map[string]util.MockCommand{
				"systemctl is-active test-svc":  {Output: tt.isActive, ExitCode: 0},
				"systemctl is-enabled test-svc": {Output: tt.isEnabled, ExitCode: 0},
			}}
			sr := ServiceResource{
				service:        "test-svc",
				desiredState:   tt.desiredState,
				desiredEnabled: tt.desiredEnabled,
				cmder:          cmder,
			}
			result, err := sr.Check(context.Background())
			assert.Nil(t, err)
			assert.Equal(t, tt.wantChanged, result.Changed)
			assert.Equal(t, tt.wantDiffs, len(result.Differences))
		})
	}
}

func TestServiceResource_Apply(t *testing.T) {
	tests := []struct {
		name           string
		desiredState   string
		desiredEnabled bool
		isActive       string
		isEnabled      string
		wantResult     resource.State
		wantErr        bool
	}{
		{"no changes needed", "running", true, "active", "enabled", resource.Changed, false},
		{"start service", "running", true, "inactive", "enabled", resource.Changed, false},
		{"stop service", "stopped", false, "active", "disabled", resource.Changed, false},
		{"enable service", "running", true, "active", "disabled", resource.Changed, false},
		{"disable service", "stopped", false, "inactive", "enabled", resource.Changed, false},
		{"start and enable", "running", true, "inactive", "disabled", resource.Changed, false},
		{"start fails", "running", false, "inactive", "disabled", resource.Failed, true},
		{"enable fails", "running", true, "active", "disabled", resource.Failed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := map[string]util.MockCommand{
				"systemctl is-active test-svc":  {Output: tt.isActive, ExitCode: 0},
				"systemctl is-enabled test-svc": {Output: tt.isEnabled, ExitCode: 0},
				"systemctl start test-svc":      {ExitCode: 0},
				"systemctl stop test-svc":       {ExitCode: 0},
				"systemctl enable test-svc":     {ExitCode: 0},
				"systemctl disable test-svc":    {ExitCode: 0},
			}

			// Override to make specific commands fail
			if tt.name == "start fails" {
				commands["systemctl start test-svc"] = util.MockCommand{ExitCode: 1}
			}
			if tt.name == "enable fails" {
				commands["systemctl enable test-svc"] = util.MockCommand{ExitCode: 1}
			}

			sr := ServiceResource{
				service:        "test-svc",
				desiredState:   tt.desiredState,
				desiredEnabled: tt.desiredEnabled,
				cmder:          util.MockCommander{Commands: commands},
			}
			result, err := sr.Apply(context.Background())
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.wantResult, result.Result)
		})
	}
}
