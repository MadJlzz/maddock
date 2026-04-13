package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"text/template"

	"github.com/MadJlzz/maddock/internal/resource"
)

func init() {
	resource.Register("file", func(name string, attrs map[string]any) (resource.Resource, error) {
		f := &FileResource{path: name}

		if val, ok := attrs["owner"]; ok {
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("expected string attr 'owner'")
			}
			f.owner = s
		}
		if val, ok := attrs["group"]; ok {
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("expected string attr 'group'")
			}
			f.group = s
		}
		if val, ok := attrs["mode"]; ok {
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("expected string attr 'mode'")
			}
			f.mode = s
		}

		// Content: either inline "content" or "source" template, not both.
		_, hasContent := attrs["content"]
		_, hasSource := attrs["source"]
		if hasContent && hasSource {
			return nil, fmt.Errorf("file %s: cannot specify both 'content' and 'source'", name)
		}

		if hasContent {
			s, ok := attrs["content"].(string)
			if !ok {
				return nil, fmt.Errorf("expected string attr 'content'")
			}
			f.content = s
		} else if hasSource {
			s, ok := attrs["source"].(string)
			if !ok {
				return nil, fmt.Errorf("expected string attr 'source'")
			}
			f.source = s

			// Parse vars for template rendering.
			vars := make(map[string]string)
			if rawVars, ok := attrs["vars"]; ok {
				m, ok := rawVars.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("expected map attr 'vars'")
				}
				for k, v := range m {
					vars[k] = fmt.Sprintf("%v", v)
				}
			}
			f.vars = vars

			// Render template at parse time.
			tmpl, err := template.ParseFiles(s)
			if err != nil {
				return nil, fmt.Errorf("file %s: failed to parse template %s: %w", name, s, err)
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, vars); err != nil {
				return nil, fmt.Errorf("file %s: failed to render template %s: %w", name, s, err)
			}
			f.content = buf.String()
		}

		return f, nil
	})
}

type FileResource struct {
	path    string
	source  string
	content string
	owner   string
	group   string
	mode    string
	vars    map[string]string
}

func (f *FileResource) Type() string {
	return "file"
}

func (f *FileResource) Name() string {
	return f.path
}

func (f *FileResource) checkOwnership(fi os.FileInfo) ([]resource.Difference, error) {
	sysStat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok || sysStat == nil {
		return nil, fmt.Errorf("cannot fetch or empty stat_t struct from file %s", fi.Name())
	}
	diffs := make([]resource.Difference, 0)

	userId := strconv.Itoa(int(sysStat.Uid))
	u, err := user.LookupId(userId)
	if err != nil {
		return nil, err
	}
	if u.Username != f.owner {
		diffs = append(diffs, resource.Difference{
			Attribute: "owner",
			Current:   u.Username,
			Desired:   f.owner,
		})
	}

	groupId := strconv.Itoa(int(sysStat.Gid))
	g, err := user.LookupGroupId(groupId)
	if err != nil {
		return nil, err
	}
	if g.Name != f.group {
		diffs = append(diffs, resource.Difference{
			Attribute: "group",
			Current:   g.Name,
			Desired:   f.group,
		})
	}
	return diffs, nil
}

func (f *FileResource) checkMode(fi os.FileInfo) (*resource.Difference, error) {
	desired, err := strconv.ParseUint(f.mode, 8, 32)
	if err != nil {
		return nil, err
	}
	current := fi.Mode().Perm()
	if os.FileMode(desired) == current {
		return nil, nil
	}
	return &resource.Difference{
		Attribute: "mode",
		Current:   fmt.Sprintf("%04o", current),
		Desired:   f.mode,
	}, nil
}

func (f *FileResource) checkContent() (*resource.Difference, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return nil, err
	}
	currentSum := sha256.Sum256(data)
	desiredSum := sha256.Sum256([]byte(f.content))
	if currentSum == desiredSum {
		return nil, nil
	}

	return &resource.Difference{
		Attribute: "content",
		Current:   fmt.Sprintf("%x", currentSum),
		Desired:   fmt.Sprintf("%x", desiredSum),
	}, nil
}

func (f *FileResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	fi, err := os.Stat(f.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &resource.CheckResult{Changed: true, Differences: []resource.Difference{
				{
					Attribute: "path",
					Current:   "",
					Desired:   f.path,
				},
			}}, nil
		}
		return nil, err
	}
	var diffs []resource.Difference
	var changed bool

	ownership, err := f.checkOwnership(fi)
	if err != nil {
		return nil, err
	}
	if len(ownership) > 0 {
		changed = true
		diffs = append(diffs, ownership...)
	}

	mode, err := f.checkMode(fi)
	if err != nil {
		return nil, err
	}
	if mode != nil {
		changed = true
		diffs = append(diffs, *mode)
	}

	content, err := f.checkContent()
	if err != nil {
		return nil, err
	}
	if content != nil {
		changed = true
		diffs = append(diffs, *content)
	}

	return &resource.CheckResult{
		Changed:     changed,
		Differences: diffs,
	}, nil
}

func (f *FileResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	// Write desired content to a temp file in the same directory
	// so os.Rename is an atomic same-filesystem move.
	dir := filepath.Dir(f.path)
	tmp, err := os.CreateTemp(dir, ".maddock-*")
	if err != nil {
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.WriteString(f.content); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("writing content: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("closing temp file: %w", err)
	}

	// Set mode.
	mode, err := strconv.ParseUint(f.mode, 8, 32)
	if err != nil {
		_ = os.Remove(tmpPath)
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("parsing mode: %w", err)
	}
	if err := os.Chmod(tmpPath, os.FileMode(mode)); err != nil {
		_ = os.Remove(tmpPath)
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("chmod: %w", err)
	}

	// Set ownership.
	u, err := user.Lookup(f.owner)
	if err != nil {
		_ = os.Remove(tmpPath)
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("looking up owner %s: %w", f.owner, err)
	}
	g, err := user.LookupGroup(f.group)
	if err != nil {
		_ = os.Remove(tmpPath)
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("looking up group %s: %w", f.group, err)
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(g.Gid)
	if err := os.Chown(tmpPath, uid, gid); err != nil {
		_ = os.Remove(tmpPath)
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("chown: %w", err)
	}

	// Atomic rename into place.
	if err := os.Rename(tmpPath, f.path); err != nil {
		_ = os.Remove(tmpPath)
		return &resource.ApplyResult{Result: resource.Failed}, fmt.Errorf("rename: %w", err)
	}

	return &resource.ApplyResult{Result: resource.Changed}, nil
}
