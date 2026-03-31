package pkg

import (
	"context"

	"github.com/MadJlzz/maddock/internal/util"
)

type aptManager struct {
	cmder util.Commander
}

func newAptManager(cmder util.Commander) *aptManager {
	return &aptManager{
		cmder: cmder,
	}
}

func (am *aptManager) IsInstalled(ctx context.Context, pkg string) (bool, string, error) {
	// TODO implement me
	panic("implement me")
}

func (am *aptManager) Install(ctx context.Context, pkg string) error {
	// TODO implement me
	panic("implement me")
}

func (am *aptManager) Remove(ctx context.Context, pkg string) error {
	// TODO implement me
	panic("implement me")
}
