package main

import (
	"fmt"
	"os"

	_ "github.com/MadJlzz/maddock/internal/resources/file"
	_ "github.com/MadJlzz/maddock/internal/resources/pkg"
	_ "github.com/MadJlzz/maddock/internal/resources/service"

	"github.com/spf13/cobra"
)

var Version = "dev"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "maddock-server",
		Short:   "Maddock server — orchestrates catalog pushes to agents",
		Version: Version,
	}
	cmd.AddCommand(newPushCmd())
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
