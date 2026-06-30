package main

import (
	"fmt"
	"os"

	"github.com/MadJlzz/maddock/internal/logging"
	_ "github.com/MadJlzz/maddock/internal/resources/command"
	_ "github.com/MadJlzz/maddock/internal/resources/file"
	_ "github.com/MadJlzz/maddock/internal/resources/hostname"
	_ "github.com/MadJlzz/maddock/internal/resources/kernel"
	_ "github.com/MadJlzz/maddock/internal/resources/pkg"
	_ "github.com/MadJlzz/maddock/internal/resources/service"

	"github.com/spf13/cobra"
)

var Version = "dev"

func newRootCmd() *cobra.Command {
	var logLevel string
	stateDir, _ := defaultStateDir()
	cmd := &cobra.Command{
		Use:           "maddock-controlplane",
		Short:         "Maddock control plane — orchestrates catalog pushes to agents",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if stateDir == "" {
				d, err := defaultStateDir()
				if err != nil {
					return err
				}
				stateDir = d
			}
			return logging.Setup(logLevel)
		},
	}
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level: debug|info|warn|error")
	cmd.PersistentFlags().StringVar(&stateDir, "state-dir", stateDir, "path to control plane state directory")
	cmd.AddCommand(newPushCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newCertCmd())
	cmd.AddCommand(newTokenCmd())
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
