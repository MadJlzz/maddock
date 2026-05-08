package kernel

import (
	"context"
	"fmt"
	"strings"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
)

func init() {
	resource.Register("sysctl", func(name string, attrs map[string]any) (resource.Resource, error) {
		val, found := attrs["value"]
		if !found {
			return nil, fmt.Errorf("missing attr 'value'")
		}
		paramValue, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("value attribute is expected to be a string")
		}
		return &SysctlResource{Key: name, DesiredValue: paramValue}, nil
	})
}

type SysctlResource struct {
	Key          string
	DesiredValue string
	cmder        util.Commander
}

func (s *SysctlResource) Type() string {
	return "sysctl"
}

func (s *SysctlResource) Name() string {
	return s.Key
}

func (s *SysctlResource) GetKernelValue(ctx context.Context) (string, error) {
	stdout, _, status, err := s.cmder.Run(ctx, "sysctl", []string{"--values", s.Key})
	if err != nil {
		return "", err
	}
	if status != 0 {
		return "", fmt.Errorf("sysctl returned non-zero status: %d", status)
	}
	return strings.TrimSpace(stdout), nil
}

func (s *SysctlResource) SetKernelValue(ctx context.Context, value string) error {
	_, _, status, err := s.cmder.Run(ctx, "sysctl", []string{"--write", fmt.Sprintf("%s=%s", s.Key, value)})
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("sysctl returned non-zero status: %d", status)
	}
	return nil
}

func (s *SysctlResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	currentValue, err := s.GetKernelValue(ctx)
	if err != nil {
		return nil, err
	}

	shouldBeUpdated := currentValue != s.DesiredValue
	var diffs []resource.Difference
	if shouldBeUpdated {
		diffs = append(diffs, resource.Difference{
			Attribute: "value",
			Current:   currentValue,
			Desired:   s.DesiredValue,
		})
	}

	return &resource.CheckResult{
		Changed:     shouldBeUpdated,
		Differences: diffs,
	}, nil
}

func (s *SysctlResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	currentValue, err := s.GetKernelValue(ctx)
	if err != nil {
		return nil, err
	}

	shouldBeUpdated := currentValue != s.DesiredValue
	if shouldBeUpdated {
		if err = s.SetKernelValue(ctx, s.DesiredValue); err != nil {
			return &resource.ApplyResult{Result: resource.Failed}, err
		}
		return &resource.ApplyResult{Result: resource.Changed}, nil
	}

	return &resource.ApplyResult{Result: resource.Ok}, nil
}
