package pkg

import (
	"context"
	"os/exec"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
)

func init() {
	resource.Register("package", func(name string, attrs map[string]any) (resource.Resource, error) {
		manager := detectManager()
		state, _ := attrs["state"].(string)
		return &PackageResource{
			pkg:          name,
			desiredState: state,
			manager:      manager,
		}, nil
	})
}

func detectManager() Manager {
	if _, err := exec.LookPath("dnf"); err == nil {
		return newDnfManager(util.RealCommander{})
	} else if _, err := exec.LookPath("apt"); err == nil {
		return newAptManager(util.RealCommander{})
	}
	panic("no dependency manager found - this should never happen")
}

type Manager interface {
	IsInstalled(ctx context.Context, pkg string) (bool, string, error)
	Install(ctx context.Context, pkg string) error
	Remove(ctx context.Context, pkg string) error
}

type PackageResource struct {
	pkg          string
	desiredState string
	manager      Manager
}

func (pr *PackageResource) Type() string {
	return "package"
}

func (pr *PackageResource) Name() string {
	return pr.pkg
}

func (pr *PackageResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	installed, _, err := pr.manager.IsInstalled(ctx, pr.pkg)
	if err != nil {
		return nil, err
	}

	shouldBeInstalled := pr.desiredState == "present"
	changed := shouldBeInstalled != installed

	var diffs []resource.Difference
	if changed {
		current := "absent"
		if installed {
			current = "present"
		}
		diffs = append(diffs, resource.Difference{
			Attribute: "state",
			Current:   current,
			Desired:   pr.desiredState,
		})
	}

	return &resource.CheckResult{
		Changed:     changed,
		Differences: diffs,
	}, nil
}

func (pr *PackageResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	installed, _, err := pr.manager.IsInstalled(ctx, pr.pkg)
	if err != nil {
		return nil, err
	}

	if pr.desiredState == "present" && !installed {
		if err = pr.manager.Install(ctx, pr.pkg); err != nil {
			return &resource.ApplyResult{Result: resource.Failed}, err
		}
		return &resource.ApplyResult{Result: resource.Changed}, nil
	} else if pr.desiredState == "absent" && installed {
		if err = pr.manager.Remove(ctx, pr.pkg); err != nil {
			return &resource.ApplyResult{Result: resource.Failed}, err
		}
		return &resource.ApplyResult{Result: resource.Changed}, nil
	}

	return &resource.ApplyResult{Result: resource.Ok}, nil
}
