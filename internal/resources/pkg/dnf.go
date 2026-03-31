package pkg

import (
	"context"
	"io"

	"github.com/MadJlzz/maddock/internal/util"
)

type dnfManager struct {
	cmder util.Commander
}

func newDnfManager(cmder util.Commander) *dnfManager {
	return &dnfManager{
		cmder: cmder,
	}
}

func (dm *dnfManager) IsInstalled(ctx context.Context, pkg string) (bool, string, error) {
	stdout, _, status, err := dm.cmder.Run(ctx, "rpm", []string{"--query", pkg})
	if err != nil {
		return false, "", err
	}
	if status != 0 {
		return false, "", nil
	}
	pkgName, err := io.ReadAll(stdout)
	if err != nil {
		return false, "", err
	}
	return true, string(pkgName), nil
}

func (dm *dnfManager) Install(ctx context.Context, pkg string) error {
	_, _, _, err := dm.cmder.Run(ctx, "dnf", []string{"install", "--yes", pkg})
	return err
}

func (dm *dnfManager) Remove(ctx context.Context, pkg string) error {
	_, _, _, err := dm.cmder.Run(ctx, "dnf", []string{"remove", "--yes", pkg})
	return err
}
