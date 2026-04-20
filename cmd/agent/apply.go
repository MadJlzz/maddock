package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/engine"
	"github.com/MadJlzz/maddock/internal/report"

	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	var (
		dryRun bool
		output string
	)
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
			if err := writeReport(cmd.OutOrStdout(), r, output); err != nil {
				return err
			}
			os.Exit(r.ExitCode())
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "check-only mode, no changes applied")
	cmd.Flags().StringVar(&output, "output", "text", "output format: text|json")
	return cmd
}

func writeReport(w interface{ Write(p []byte) (int, error) }, r *report.Report, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	case "text", "":
		_, err := fmt.Fprintln(w, r)
		return err
	default:
		return fmt.Errorf("unknown output format %q (expected text|json)", format)
	}
}
