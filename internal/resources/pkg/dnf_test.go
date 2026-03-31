package pkg

import (
	"context"
	"testing"

	"github.com/MadJlzz/maddock/internal/util"
)

func TestDnfManager_IsInstalled(t *testing.T) {
	mockedCommands := map[string]util.MockCommand{
		"rpm---query-installedPkg": {
			Output:   "installedPkg",
			ExitCode: 0,
		},
		"rpm---query-missingPkg": {
			Output:   "",
			ExitCode: 1,
		},
	}
	manager := dnfManager{cmder: util.MockCommander{Commands: mockedCommands}}

	t.Run("package is installed", func(t *testing.T) {
		installed, pkgName, err := manager.IsInstalled(context.Background(), "installedPkg")
		if err != nil {
			t.Fatalf("failed to check if package is installed: %v", err)
		}
		if !installed {
			t.Errorf("installed = false, want true")
		}
		if pkgName != "installedPkg" {
			t.Errorf("version = %q, want %q", pkgName, "installedPkg")
		}
	})

	t.Run("package is not installed", func(t *testing.T) {
		installed, _, err := manager.IsInstalled(context.Background(), "missingPkg")
		if err != nil {
			t.Fatalf("failed to check if package is installed: %v", err)
		}
		if installed {
			t.Errorf("installed = true, want false")
		}
	})

}

func TestDnfManager_Install(t *testing.T) {
	t.FailNow()
}

func TestDnfManager_Remove(t *testing.T) {
	t.FailNow()
}
