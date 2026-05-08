package hostname

import (
	"context"
	"fmt"
	"strings"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
)

func init() {
	resource.Register("hostname", func(name string, attrs map[string]any) (resource.Resource, error) {
		val, ok := attrs["name"]
		if !ok {
			return nil, fmt.Errorf("missing attr 'name'")
		}
		hostname, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("attr 'name' must be a string")
		}
		return &HostnameResource{name: hostname, cmder: util.RealCommander{}}, nil
	})
}

type HostnameResource struct {
	name  string
	cmder util.Commander
}

func (h *HostnameResource) Type() string {
	return "hostname"
}

func (h *HostnameResource) Name() string {
	return h.name
}

func (h *HostnameResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	currentHostname, err := h.getHostname(ctx)
	if err != nil {
		return nil, err
	}

	shouldBeUpdated := h.name != currentHostname
	var diffs []resource.Difference
	if shouldBeUpdated {
		diffs = append(diffs, resource.Difference{
			Attribute: "name",
			Current:   currentHostname,
			Desired:   h.name,
		})
	}

	return &resource.CheckResult{
		Changed:     shouldBeUpdated,
		Differences: diffs,
	}, nil
}

func (h *HostnameResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	currentHostname, err := h.getHostname(ctx)
	if err != nil {
		return nil, err
	}

	shouldBeUpdated := h.name != currentHostname
	if shouldBeUpdated {
		if err = h.setHostname(ctx); err != nil {
			return &resource.ApplyResult{Result: resource.Failed}, err
		}
		return &resource.ApplyResult{Result: resource.Changed}, nil
	}

	return &resource.ApplyResult{Result: resource.Ok}, nil
}

func (h *HostnameResource) getHostname(ctx context.Context) (string, error) {
	hostname, _, _, err := h.cmder.Run(ctx, "hostnamectl", []string{"--static"})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(hostname), nil
}

func (h *HostnameResource) setHostname(ctx context.Context) error {
	_, _, status, err := h.cmder.Run(ctx, "hostnamectl", []string{"hostname", h.name})
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}
