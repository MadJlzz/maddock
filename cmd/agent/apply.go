package main

import (
	"fmt"
	"os"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/engine"

	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "apply <manifest.yaml>",
		Short: "Apply a manifest to the local host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rawManifest, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			c, err := catalog.Parse(rawManifest)
			if err != nil {
				return err
			}
			r := engine.Run(cmd.Context(), c, dryRun)
			fmt.Println(r)
			os.Exit(r.ExitCode())
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "check-only mode, no changes applied")
	return cmd
}
