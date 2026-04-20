package main

import (
	"fmt"
	"os"

	"github.com/MadJlzz/maddock/internal/logging"
	_ "github.com/MadJlzz/maddock/internal/resources/file"
	_ "github.com/MadJlzz/maddock/internal/resources/pkg"
	_ "github.com/MadJlzz/maddock/internal/resources/service"

	"github.com/spf13/cobra"
)

var Version = "dev"

func newRootCmd() *cobra.Command {
	var logLevel string
	cmd := &cobra.Command{
		Use:     "maddock-agent",
		Short:   "Maddock agent — converge a Linux host to a desired state",
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.Setup(logLevel)
		},
	}
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level: debug|info|warn|error")
	cmd.AddCommand(newApplyCmd())
	cmd.AddCommand(newServeCmd())
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
