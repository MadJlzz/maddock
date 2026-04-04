package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
)

func init() {
	resource.Register("service", func(name string, attrs map[string]any) (resource.Resource, error) {
		val, ok := attrs["state"]
		if !ok {
			return nil, fmt.Errorf("missing attr 'state'")
		}
		state, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("expected string attr 'state'")
		}
		if state != "running" && state != "stopped" {
			return nil, fmt.Errorf("state is expected to be 'running' or 'stopped'")
		}
		val, ok = attrs["enabled"]
		if !ok {
			return nil, fmt.Errorf("missing attr 'enabled'")
		}
		enabled, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool attr 'enabled'")
		}
		return &ServiceResource{
			service:        name,
			desiredState:   state,
			desiredEnabled: enabled,
			cmder:          util.RealCommander{},
		}, nil
	})
}

type ServiceResource struct {
	service        string
	desiredState   string
	desiredEnabled bool
	cmder          util.Commander
}

func (sr *ServiceResource) Type() string {
	return "service"
}

func (sr *ServiceResource) Name() string {
	return sr.service
}

func (sr *ServiceResource) Status(ctx context.Context) (string, error) {
	stdout, _, _, err := sr.cmder.Run(ctx, "systemctl", []string{"is-active", sr.service})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

func (sr *ServiceResource) Enabled(ctx context.Context) (bool, error) {
	stdout, _, _, err := sr.cmder.Run(ctx, "systemctl", []string{"is-enabled", sr.service})
	if err != nil {
		return false, err
	}
	trimmed := strings.TrimSpace(stdout)
	return trimmed == "enabled" || trimmed == "enabled-runtime", nil
}

func (sr *ServiceResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	status, err := sr.Status(ctx)
	if err != nil {
		return nil, err
	}
	enabled, err := sr.Enabled(ctx)
	if err != nil {
		return nil, err
	}

	var diffs []resource.Difference

	currentState := "stopped"
	if status == "active" {
		currentState = "running"
	}
	if currentState != sr.desiredState {
		diffs = append(diffs, resource.Difference{
			Attribute: "state",
			Current:   currentState,
			Desired:   sr.desiredState,
		})
	}

	currentEnabled := "false"
	if enabled {
		currentEnabled = "true"
	}
	desiredEnabled := "false"
	if sr.desiredEnabled {
		desiredEnabled = "true"
	}
	if currentEnabled != desiredEnabled {
		diffs = append(diffs, resource.Difference{
			Attribute: "enabled",
			Current:   currentEnabled,
			Desired:   desiredEnabled,
		})
	}

	return &resource.CheckResult{
		Changed:     len(diffs) > 0,
		Differences: diffs,
	}, nil
}

func (sr *ServiceResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	status, err := sr.Status(ctx)
	if err != nil {
		return nil, err
	}
	enabled, err := sr.Enabled(ctx)
	if err != nil {
		return nil, err
	}

	currentState := "stopped"
	if status == "active" {
		currentState = "running"
	}

	if currentState != sr.desiredState {
		action := "start"
		if sr.desiredState == "stopped" {
			action = "stop"
		}
		_, stderr, exitCode, err := sr.cmder.Run(ctx, "systemctl", []string{action, sr.service})
		if err != nil {
			return &resource.ApplyResult{Result: resource.Failed}, err
		}
		if exitCode != 0 {
			return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("systemctl %s failed with status %d and err %s", action, exitCode, stderr)
		}
	}

	if enabled != sr.desiredEnabled {
		action := "enable"
		if !sr.desiredEnabled {
			action = "disable"
		}
		_, stderr, exitCode, err := sr.cmder.Run(ctx, "systemctl", []string{action, sr.service})
		if err != nil {
			return &resource.ApplyResult{Result: resource.Failed}, err
		}
		if exitCode != 0 {
			return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("systemctl %s failed with status %d and err %s", action, exitCode, stderr)
		}
	}

	return &resource.ApplyResult{Result: resource.Changed}, nil
}
