package file

import (
	"context"

	"github.com/MadJlzz/maddock/internal/resource"
)

//- file:
///etc/nginx/nginx.conf:
//source: templates/nginx.conf.tmpl
//owner: root
//group: root
//mode: "0644"
//vars:
//worker_connections: 1024

func init() {
	resource.Register("file", func(name string, attrs map[string]any) (resource.Resource, error) {
		return &FileResource{}, nil
	})
}

type FileResource struct {
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
	panic("implement me")
}

func (f *FileResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	// Check file existence
	// Check owner and group
	// Check permissions/mode
	// Check that content hash is equal to written file hash
	// If any of this is different, there is a change.
	// Probably we will have a ContentManager that will either, get the content from file/inline
	// or template a file from a template file using gotemplate.
	panic("implement me")
}

func (f *FileResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	panic("implement me")
}
