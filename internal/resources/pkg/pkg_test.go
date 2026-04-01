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
	mp, ok := m.Results[pkg]
	if !ok {
		return false, "", fmt.Errorf("package %s not found", pkg)
	}
	return mp.installed, "", nil
}

func (m *MockManager) Install(ctx context.Context, pkg string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockManager) Remove(ctx context.Context, pkg string) error {
	//TODO implement me
	panic("implement me")
}

func TestPackageResource_Apply(t *testing.T) {

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
