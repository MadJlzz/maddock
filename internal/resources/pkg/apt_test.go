package pkg

import (
	"context"
	"testing"

	"github.com/MadJlzz/maddock/internal/util"
)

func TestAptManager_IsInstalled(t *testing.T) {
	mockedCommands := map[string]util.MockCommand{
		"dpkg-query --status installedPkg": {
			Output:   "installedPkg",
			ExitCode: 0,
		},
		"dpkg-query --status missingPkg": {
			Output:   "",
			ExitCode: 1,
		},
	}
	manager := aptManager{cmder: util.MockCommander{Commands: mockedCommands}}

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

func TestAptManager_Install(t *testing.T) {
	mockedCommands := map[string]util.MockCommand{
		"apt-get install --yes pkg": {
			ExitCode: 0,
		},
		"apt-get install --yes errorPkg": {
			ExitCode: 1,
		},
	}
	manager := aptManager{cmder: util.MockCommander{Commands: mockedCommands}}

	t.Run("package is installed", func(t *testing.T) {
		err := manager.Install(context.Background(), "pkg")
		if err != nil {
			t.Fatalf("failed to install new package: %v", err)
		}
	})

	t.Run("package installation errors", func(t *testing.T) {
		err := manager.Install(context.Background(), "errorPkg")
		if err == nil {
			t.Fatalf("install failed, want error")
		}
	})

}

func TestAptManager_Remove(t *testing.T) {
	mockedCommands := map[string]util.MockCommand{
		"apt-get remove --yes pkg": {
			ExitCode: 0,
		},
		"apt-get remove --yes errorPkg": {
			ExitCode: 1,
		},
	}
	manager := aptManager{cmder: util.MockCommander{Commands: mockedCommands}}

	t.Run("package is removed", func(t *testing.T) {
		err := manager.Remove(context.Background(), "pkg")
		if err != nil {
			t.Fatalf("failed to remove package: %v", err)
		}
	})

	t.Run("package remove errors", func(t *testing.T) {
		err := manager.Remove(context.Background(), "errorPkg")
		if err == nil {
			t.Fatalf("remove failed, want error")
		}
	})

}
