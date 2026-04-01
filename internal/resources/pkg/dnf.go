package pkg

import (
	"context"
	"fmt"

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
	return true, stdout, nil
}

func (dm *dnfManager) Install(ctx context.Context, pkg string) error {
	_, stderr, status, err := dm.cmder.Run(ctx, "dnf", []string{"install", "--yes", pkg})
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("dnf install failed with status %d and err %s", status, stderr)
	}
	return nil
}

func (dm *dnfManager) Remove(ctx context.Context, pkg string) error {
	_, stderr, status, err := dm.cmder.Run(ctx, "dnf", []string{"remove", "--yes", pkg})
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("dnf remove failed with status %d and err %v", status, stderr)
	}
	return nil
}
