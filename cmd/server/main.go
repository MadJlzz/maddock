package main

import (
	"fmt"
	"os"

	"github.com/MadJlzz/maddock/internal/logging"
	_ "github.com/MadJlzz/maddock/internal/resources/command"
	_ "github.com/MadJlzz/maddock/internal/resources/file"
	_ "github.com/MadJlzz/maddock/internal/resources/pkg"
	_ "github.com/MadJlzz/maddock/internal/resources/service"

	"github.com/spf13/cobra"
)

var Version = "dev"

func newRootCmd() *cobra.Command {
	var logLevel string
	cmd := &cobra.Command{
		Use:     "maddock-server",
		Short:   "Maddock server — orchestrates catalog pushes to agents",
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.Setup(logLevel)
		},
	}
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level: debug|info|warn|error")
	cmd.AddCommand(newPushCmd())
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
