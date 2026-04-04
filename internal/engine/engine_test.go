package engine

import (
	"context"
	"fmt"
	"testing"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/stretchr/testify/assert"
)

type mockResource struct {
	name        string
	checkResult *resource.CheckResult
	checkErr    error
	applyResult *resource.ApplyResult
	applyErr    error
}

func (m *mockResource) Type() string {
	return "mock"
}

func (m *mockResource) Name() string {
	return m.name
}

func (m *mockResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	return m.checkResult, m.checkErr
}
func (m *mockResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	return m.applyResult, m.applyErr
}

func TestRun_NoChanges(t *testing.T) {
	c := &catalog.Catalog{
		Name: "test",
		Resources: []resource.Resource{
			&mockResource{
				name:        "already-ok",
				checkResult: &resource.CheckResult{Changed: false},
			},
		},
	}

	r := Run(context.Background(), c, false)
	assert.Equal(t, 1, len(r.ResourceReports))
	assert.Equal(t, resource.Ok, r.ResourceReports[0].State)
}

func TestRun_ApplyChanges(t *testing.T) {
	c := &catalog.Catalog{
		Name: "test",
		Resources: []resource.Resource{
			&mockResource{
				name: "needs-change",
				checkResult: &resource.CheckResult{
					Changed: true,
					Differences: []resource.Difference{
						{Attribute: "state", Current: "absent", Desired: "present"},
					},
				},
				applyResult: &resource.ApplyResult{Result: resource.Changed},
			},
		},
	}

	r := Run(context.Background(), c, false)
	assert.Equal(t, 1, len(r.ResourceReports))
	assert.Equal(t, resource.Changed, r.ResourceReports[0].State)
	assert.Equal(t, 1, len(r.ResourceReports[0].Differences))
}

func TestRun_DryRunSkipsApply(t *testing.T) {
	c := &catalog.Catalog{
		Name: "test",
		Resources: []resource.Resource{
			&mockResource{
				name: "would-change",
				checkResult: &resource.CheckResult{
					Changed: true,
					Differences: []resource.Difference{
						{Attribute: "state", Current: "absent", Desired: "present"},
					},
				},
				applyResult: &resource.ApplyResult{Result: resource.Changed},
			},
		},
	}

	r := Run(context.Background(), c, true)
	assert.Equal(t, 1, len(r.ResourceReports))
	assert.Equal(t, resource.Skipped, r.ResourceReports[0].State)
	assert.Equal(t, 1, len(r.ResourceReports[0].Differences))
}

func TestRun_CheckError(t *testing.T) {
	c := &catalog.Catalog{
		Name: "test",
		Resources: []resource.Resource{
			&mockResource{
				name:     "check-fails",
				checkErr: fmt.Errorf("check failed"),
			},
		},
	}

	r := Run(context.Background(), c, false)
	assert.Equal(t, 1, len(r.ResourceReports))
	assert.Equal(t, resource.Failed, r.ResourceReports[0].State)
	assert.NotNil(t, r.ResourceReports[0].Error)
}

func TestRun_ApplyError(t *testing.T) {
	c := &catalog.Catalog{
		Name: "test",
		Resources: []resource.Resource{
			&mockResource{
				name:        "apply-fails",
				checkResult: &resource.CheckResult{Changed: true},
				applyResult: &resource.ApplyResult{Result: resource.Failed},
				applyErr:    fmt.Errorf("apply failed"),
			},
		},
	}

	r := Run(context.Background(), c, false)
	assert.Equal(t, 1, len(r.ResourceReports))
	assert.Equal(t, resource.Failed, r.ResourceReports[0].State)
	assert.NotNil(t, r.ResourceReports[0].Error)
}

func TestRun_MultipleResources(t *testing.T) {
	c := &catalog.Catalog{
		Name: "test",
		Resources: []resource.Resource{
			&mockResource{
				name:        "ok-resource",
				checkResult: &resource.CheckResult{Changed: false},
			},
			&mockResource{
				name:        "changed-resource",
				checkResult: &resource.CheckResult{Changed: true},
				applyResult: &resource.ApplyResult{Result: resource.Changed},
			},
			&mockResource{
				name:     "failed-resource",
				checkErr: fmt.Errorf("broken"),
			},
		},
	}

	r := Run(context.Background(), c, false)
	assert.Equal(t, 3, len(r.ResourceReports))
	assert.Equal(t, resource.Ok, r.ResourceReports[0].State)
	assert.Equal(t, resource.Changed, r.ResourceReports[1].State)
	assert.Equal(t, resource.Failed, r.ResourceReports[2].State)
}
