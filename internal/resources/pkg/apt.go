package pkg

import (
	"context"
	"fmt"

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
	stdout, _, status, err := am.cmder.Run(ctx, "dpkg-query", []string{"--status", pkg})
	if err != nil {
		return false, "", err
	}
	if status != 0 {
		return false, "", nil
	}
	return true, stdout, nil
}

func (am *aptManager) Install(ctx context.Context, pkg string) error {
	_, stderr, status, err := am.cmder.Run(ctx, "apt-get", []string{"install", "--yes", pkg})
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("apt-get install failed with status %d and err %s", status, stderr)
	}
	return nil
}

func (am *aptManager) Remove(ctx context.Context, pkg string) error {
	_, stderr, status, err := am.cmder.Run(ctx, "apt-get", []string{"remove", "--yes", pkg})
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("apt-get remove failed with status %d and err %v", status, stderr)
	}
	return nil
}
