package pkg

import (
	"context"

	"github.com/MadJlzz/maddock/internal/resource"
)

type RepositoryResource struct {
	manager Manager
}

func (r *RepositoryResource) Type() string {
	//TODO implement me
	panic("implement me")
}

func (r *RepositoryResource) Name() string {
	//TODO implement me
	panic("implement me")
}

func (r *RepositoryResource) Check(ctx context.Context) (*resource.CheckResult, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RepositoryResource) Apply(ctx context.Context) (*resource.ApplyResult, error) {
	//TODO implement me
	panic("implement me")
}
