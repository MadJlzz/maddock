package file

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_InlineContent(t *testing.T) {
	r, err := resource.Parse("file", "/tmp/test.txt", map[string]any{
		"content": "hello world",
		"owner":   "root",
		"group":   "root",
		"mode":    "0644",
	})
	require.NoError(t, err)

	f := r.(*FileResource)
	assert.Equal(t, "hello world", f.content)
	assert.Equal(t, "/tmp/test.txt", f.path)
}

func TestParse_TemplateSource(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "test.tmpl")
	require.NoError(t, os.WriteFile(tmplPath, []byte("connections={{ .conns }}"), 0644))

	r, err := resource.Parse("file", "/tmp/test.conf", map[string]any{
		"source": tmplPath,
		"vars":   map[string]any{"conns": 1024},
		"owner":  "root",
		"group":  "root",
		"mode":   "0644",
	})
	require.NoError(t, err)

	f := r.(*FileResource)
	assert.Equal(t, "connections=1024", f.content)
}

func TestParse_ContentAndSourceMutuallyExclusive(t *testing.T) {
	_, err := resource.Parse("file", "/tmp/test.txt", map[string]any{
		"content": "hello",
		"source":  "some/template.tmpl",
		"owner":   "root",
		"group":   "root",
		"mode":    "0644",
	})
	assert.Error(t, err)
}

func TestCheckContent_Matching(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0644))

	f := &FileResource{path: path, content: "hello"}
	diff, err := f.checkContent()
	require.NoError(t, err)
	assert.Nil(t, diff)
}

func TestCheckContent_Different(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("old content"), 0644))

	f := &FileResource{path: path, content: "new content"}
	diff, err := f.checkContent()
	require.NoError(t, err)
	require.NotNil(t, diff)
	assert.Equal(t, "content", diff.Attribute)
	assert.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte("old content"))), diff.Current)
	assert.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte("new content"))), diff.Desired)
}

func TestCheckMode_Matching(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0755))

	fi, err := os.Stat(path)
	require.NoError(t, err)

	f := &FileResource{mode: "0755"}
	diff, err := f.checkMode(fi)
	require.NoError(t, err)
	assert.Nil(t, diff)
}

func TestCheckMode_Different(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0644))

	fi, err := os.Stat(path)
	require.NoError(t, err)

	f := &FileResource{mode: "0755"}
	diff, err := f.checkMode(fi)
	require.NoError(t, err)
	require.NotNil(t, diff)
	assert.Equal(t, "mode", diff.Attribute)
	assert.Equal(t, "0644", diff.Current)
	assert.Equal(t, "0755", diff.Desired)
}

func TestCheck_FileDoesNotExist(t *testing.T) {
	f := &FileResource{
		path:    "/tmp/nonexistent-maddock-test-file",
		content: "hello",
		mode:    "0644",
	}
	result, err := f.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Changed)
	require.Len(t, result.Differences, 1)
	assert.Equal(t, "path", result.Differences[0].Attribute)
}

func TestCheck_AllMatching(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0644))

	f := &FileResource{
		path:    path,
		content: "hello",
		mode:    "0644",
	}
	setCurrentOwnership(t, path, f)

	result, err := f.Check(context.Background())
	require.NoError(t, err)
	assert.False(t, result.Changed, "expected no changes, got diffs: %+v", result.Differences)
}

func TestCheck_ContentDiffers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0644))

	f := &FileResource{
		path:    path,
		content: "new",
		mode:    "0644",
	}
	setCurrentOwnership(t, path, f)

	result, err := f.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Changed)

	found := false
	for _, d := range result.Differences {
		if d.Attribute == "content" {
			found = true
		}
	}
	assert.True(t, found, "expected content diff, got %+v", result.Differences)
}

func TestApply_WritesContentAndMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	f := &FileResource{
		path:    path,
		content: "applied content",
		mode:    "0755",
	}
	setCurrentOwnership(t, dir, f)

	result, err := f.Apply(context.Background())
	require.NoError(t, err)
	assert.Equal(t, resource.Changed, result.Result)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "applied content", string(data))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestApply_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	f := &FileResource{
		path:    path,
		content: "idempotent",
		mode:    "0644",
	}
	setCurrentOwnership(t, dir, f)

	_, err := f.Apply(context.Background())
	require.NoError(t, err, "first apply")

	_, err = f.Apply(context.Background())
	require.NoError(t, err, "second apply")

	result, err := f.Check(context.Background())
	require.NoError(t, err)
	assert.False(t, result.Changed, "expected no changes after double apply, got %+v", result.Differences)
}

// setCurrentOwnership populates the FileResource owner/group fields
// with the actual owner of the given path, so ownership checks pass
// without requiring root.
func setCurrentOwnership(t *testing.T, path string, f *FileResource) {
	t.Helper()
	fi, err := os.Stat(path)
	require.NoError(t, err)
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Skip("cannot read file ownership on this platform")
	}
	u, err := user.LookupId(strconv.Itoa(int(stat.Uid)))
	require.NoError(t, err)
	g, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))
	require.NoError(t, err)
	f.owner = u.Username
	f.group = g.Name
}
