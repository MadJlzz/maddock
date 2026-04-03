package pkg

import (
	"context"
	"fmt"
	"testing"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/stretchr/testify/assert"
)

func findDiff(diffs []resource.Difference, attr string) *resource.Difference {
	var diff *resource.Difference
	for _, d := range diffs {
		if d.Attribute == attr {
			diff = &d
			break
		}
	}
	return diff
}

type MockManagerResult struct {
	installed  bool
	installErr error
	removeErr  error
}

type MockManager struct {
	Results map[string]MockManagerResult
}

func (m *MockManager) IsInstalled(ctx context.Context, pkg string) (bool, string, error) {
	mmr, ok := m.Results[pkg]
	if !ok {
		return false, "", fmt.Errorf("mock results not found for package %s", pkg)
	}
	return mmr.installed, "", nil
}

func (m *MockManager) Install(ctx context.Context, pkg string) error {
	mmr, ok := m.Results[pkg]
	if !ok {
		return fmt.Errorf("mock results not found for package %s", pkg)
	}
	return mmr.installErr
}

func (m *MockManager) Remove(ctx context.Context, pkg string) error {
	mmr, ok := m.Results[pkg]
	if !ok {
		return fmt.Errorf("mock results not found for package %s", pkg)
	}
	return mmr.removeErr
}

func TestPackageResource_Apply(t *testing.T) {
	tests := []struct {
		name         string
		pkg          string
		wantErr      bool
		desiredState string
		result       resource.State
	}{
		{"installed and present, want Ok", "installedAndPresent", false, "present", resource.Ok},
		{"not installed and present, want Changed", "notInstalledAndPresent", false, "present", resource.Changed},
		{"installed and absent, want Changed", "installedAndAbsent", false, "absent", resource.Changed},
		{"not installed and absent, want Ok", "notInstalledAndAbsent", false, "absent", resource.Ok},
		{"not installed and present but Err, want Failed", "notInstalledAndPresentButErr", true, "present", resource.Failed},
		{"installed and absent but Err, want Failed", "installedAndAbsentButErr", true, "absent", resource.Failed},
	}

	mockManager := &MockManager{
		Results: map[string]MockManagerResult{
			"installedAndPresent":          {installed: true},
			"notInstalledAndPresent":       {installed: false},
			"installedAndAbsent":           {installed: true},
			"notInstalledAndAbsent":        {installed: false},
			"notInstalledAndPresentButErr": {installed: false, installErr: fmt.Errorf("mock install pkg error")},
			"installedAndAbsentButErr":     {installed: true, removeErr: fmt.Errorf("mock install pkg error")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := PackageResource{
				pkg:          tt.pkg,
				desiredState: tt.desiredState,
				manager:      mockManager,
			}
			results, err := pr.Apply(context.Background())
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.result, results.Result)
		})
	}

}

func TestPackageResource_Check(t *testing.T) {
	tests := []struct {
		name         string
		pkg          string
		currentState string
		desiredState string
		wantChanged  bool
		wantDiffs    int
	}{
		{"installed, want present", "existingPkg", "present", "present", false, 0},
		{"missing, want present", "toBeInstalledPkg", "absent", "present", true, 1},
		{"installed, want absent", "existingPkg", "present", "absent", true, 1},
		{"missing, want absent", "toBeInstalledPkg", "absent", "absent", false, 0},
	}

	mockManager := &MockManager{Results: map[string]MockManagerResult{
		"existingPkg": {
			installed: true,
		},
		"toBeInstalledPkg": {
			installed: false,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := PackageResource{
				pkg:          tt.pkg,
				desiredState: tt.desiredState,
				manager:      mockManager,
			}
			result, err := pr.Check(context.Background())
			assert.Nil(t, err, tt.name)
			assert.Equal(t, tt.wantChanged, result.Changed, tt.name)
			assert.Equal(t, tt.wantDiffs, len(result.Differences), tt.name)
			if len(result.Differences) > 0 {
				stateDiff := findDiff(result.Differences, "state")
				assert.NotNil(t, stateDiff, tt.name)
				assert.Equal(t, tt.currentState, stateDiff.Current, tt.name)
				assert.Equal(t, tt.desiredState, stateDiff.Desired, tt.name)
			}
		})
	}
}
