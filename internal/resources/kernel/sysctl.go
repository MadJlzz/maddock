package kernel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/util"
)

const (
	defaultSysctlDir      = "/etc/sysctl.d"
	defaultSysctlFilename = "99-maddock.conf"
)

func init() {
	resource.Register("sysctl", func(name string, attrs map[string]any) (resource.Resource, error) {
		raw, found := attrs["values"]
		if !found {
			return nil, fmt.Errorf("missing attr 'values'")
		}
		rawMap, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("attr 'values' must be a map of string to string")
		}
		if len(rawMap) == 0 {
			return nil, fmt.Errorf("attr 'values' must contain at least one entry")
		}
		values := make(map[string]string, len(rawMap))
		for k, v := range rawMap {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("attr 'values[%s]' must be a string", k)
			}
			values[k] = s
		}

		filename := defaultSysctlFilename
		if f, ok := attrs["filename"]; ok {
			fs, ok := f.(string)
			if !ok {
				return nil, fmt.Errorf("attr 'filename' must be a string")
			}
			filename = fs
		}

		return &SysctlResource{
			ResourceName: name,
			Values:       values,
			Filename:     filename,
			dir:          defaultSysctlDir,
			cmder:        util.RealCommander{},
		}, nil
	})
}

type SysctlResource struct {
	ResourceName string
	Values       map[string]string
	Filename     string
	dir          string
	cmder        util.Commander
}

func (s *SysctlResource) Type() string {
	return "sysctl"
}

func (s *SysctlResource) Name() string {
	return s.ResourceName
}

func (s *SysctlResource) path() string {
	return filepath.Join(s.dir, s.Filename)
}

// renderFile returns the deterministic content of the managed sysctl.d file.
// Keys are sorted so that identical Values maps always produce identical bytes.
func (s *SysctlResource) renderFile() string {
	keys := make([]string, 0, len(s.Values))
	for k := range s.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&sb, "%s = %s\n", k, s.Values[k])
	}
	return sb.String()
}

func (s *SysctlResource) getRuntimeValue(ctx context.Context, key string) (string, error) {
	stdout, _, status, err := s.cmder.Run(ctx, "sysctl", []string{"--values", key})
	if err != nil {
		return "", err
	}
	if status != 0 {
		return "", fmt.Errorf("sysctl returned non-zero status: %d", status)
	}
	return strings.TrimSpace(stdout), nil
}

func (s *SysctlResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	keys := make([]string, 0, len(s.Values))
	for k := range s.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var diffs []resource.Difference
	for _, k := range keys {
		current, err := s.getRuntimeValue(ctx, k)
		if err != nil {
			return nil, err
		}
		if current != s.Values[k] {
			diffs = append(diffs, resource.Difference{
				Attribute: k,
				Current:   current,
				Desired:   s.Values[k],
			})
		}
	}

	desired := s.renderFile()
	current, err := os.ReadFile(s.path())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading %s: %w", s.path(), err)
	}
	if string(current) != desired {
		diffs = append(diffs, resource.Difference{
			Attribute: "file",
			Current:   string(current),
			Desired:   desired,
		})
	}

	return &resource.CheckResult{
		Changed:     len(diffs) > 0,
		Differences: diffs,
	}, nil
}

func (s *SysctlResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	check, err := s.Check(ctx)
	if err != nil {
		return nil, err
	}
	if !check.Changed {
		return &resource.ApplyResult{Result: resource.Ok}, nil
	}

	if err := s.writeFileAtomic(); err != nil {
		return &resource.ApplyResult{Result: resource.Failed}, err
	}

	if err := s.reload(ctx); err != nil {
		return &resource.ApplyResult{Result: resource.Failed}, err
	}

	return &resource.ApplyResult{Result: resource.Changed}, nil
}

// writeFileAtomic writes the rendered file content to a temp file in the
// target directory, then renames it into place. Same-filesystem rename is
// atomic, so readers never see a half-written file.
func (s *SysctlResource) writeFileAtomic() error {
	path := s.path()
	tmp, err := os.CreateTemp(s.dir, ".maddock-sysctl-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.WriteString(s.renderFile()); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("writing content: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("chmod: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func (s *SysctlResource) reload(ctx context.Context) error {
	_, _, status, err := s.cmder.Run(ctx, "sysctl", []string{"-p", s.path()})
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("sysctl -p returned non-zero status: %d", status)
	}
	return nil
}
