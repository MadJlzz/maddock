package command

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
)

func init() {
	resource.Register("command", func(name string, attrs map[string]any) (resource.Resource, error) {
		rawCmd, ok := attrs["command"]
		if !ok {
			return nil, fmt.Errorf("missing attr 'command'")
		}
		cmdStr, ok := rawCmd.(string)
		if !ok {
			return nil, fmt.Errorf("expected string attr 'command'")
		}

		cr := &CommandResource{
			name:    name,
			command: cmdStr,
			cmder:   util.RealCommander{},
		}

		for _, key := range []string{"creates", "unless", "onlyif"} {
			if raw, ok := attrs[key]; ok {
				s, ok := raw.(string)
				if !ok {
					return nil, fmt.Errorf("expected string attr %q", key)
				}
				switch key {
				case "creates":
					cr.creates = s
				case "unless":
					cr.unless = s
				case "onlyif":
					cr.onlyif = s
				}
			}
		}

		return cr, nil
	})
}

type CommandResource struct {
	name    string
	command string
	creates string
	unless  string
	onlyif  string
	cmder   util.Commander
}

func (cr *CommandResource) Type() string {
	return "command"
}

func (cr *CommandResource) Name() string {
	return cr.name
}

func (cr *CommandResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	if cr.creates != "" {
		_, err := os.Stat(cr.creates)
		if err == nil {
			return &resource.CheckResult{Changed: false}, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("checking creates %q: %w", cr.creates, err)
		}
	}

	if cr.unless != "" {
		exitCode, err := cr.runGuard(ctx, cr.unless)
		if err != nil {
			return nil, fmt.Errorf("running unless guard: %w", err)
		}
		if exitCode == 0 {
			return &resource.CheckResult{Changed: false}, nil
		}
	}

	if cr.onlyif != "" {
		exitCode, err := cr.runGuard(ctx, cr.onlyif)
		if err != nil {
			return nil, fmt.Errorf("running onlyif guard: %w", err)
		}
		if exitCode != 0 {
			return &resource.CheckResult{Changed: false}, nil
		}
	}

	return &resource.CheckResult{
		Changed: true,
		Differences: []resource.Difference{{
			Attribute: "command",
			Current:   "",
			Desired:   cr.command,
		}},
	}, nil
}

func (cr *CommandResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	_, stderr, exitCode, err := cr.cmder.Run(ctx, "/bin/sh", []string{"-c", cr.command})
	if err != nil {
		return &resource.ApplyResult{Result: resource.Failed}, err
	}
	if exitCode != 0 {
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("command exited %d: %s", exitCode, stderr)
	}
	return &resource.ApplyResult{Result: resource.Changed}, nil
}

func (cr *CommandResource) runGuard(ctx context.Context, cmd string) (int, error) {
	_, _, exitCode, err := cr.cmder.Run(ctx, "/bin/sh", []string{"-c", cmd})
	return exitCode, err
}
