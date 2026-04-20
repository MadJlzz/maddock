package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/report"
	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/transport"

	"github.com/spf13/cobra"
)

type pushResult struct {
	target Target
	report *report.Report
	err    error
}

func newPushCmd() *cobra.Command {
	var (
		configPath string
		dryRun     bool
		targetName string
		parallel   int
	)
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push catalogs to one or more agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(configPath)
			if err != nil {
				return err
			}

			targets := cfg.Targets
			if targetName != "" {
				targets = slices.DeleteFunc(slices.Clone(cfg.Targets), func(t Target) bool {
					return t.Hostname != targetName
				})
				if len(targets) == 0 {
					return fmt.Errorf("no target named %q in config", targetName)
				}
			}

			if parallel < 1 {
				parallel = 1
			}

			results := pushAll(cmd.Context(), targets, dryRun, parallel)
			printResults(cmd.OutOrStdout(), results)
			os.Exit(exitCode(results))
			return nil
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "server.yaml", "path to server config")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "check-only mode, no changes applied")
	cmd.Flags().StringVar(&targetName, "target", "", "only push to this target hostname")
	cmd.Flags().IntVar(&parallel, "parallel", 4, "max concurrent pushes")
	return cmd
}

// pushAll fans out one goroutine per target, bounded by a semaphore.
// Results are written into a pre-allocated slice at the target's index so
// output order matches the config order regardless of completion order.
func pushAll(ctx context.Context, targets []Target, dryRun bool, parallel int) []pushResult {
	results := make([]pushResult, len(targets))
	sem := make(chan struct{}, parallel)
	var wg sync.WaitGroup

	for i, t := range targets {
		wg.Add(1)
		go func(i int, t Target) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = pushToTarget(ctx, t, dryRun)
		}(i, t)
	}
	wg.Wait()
	return results
}

func pushToTarget(ctx context.Context, t Target, dryRun bool) pushResult {
	result := pushResult{target: t}

	rawManifest, err := os.ReadFile(t.Manifest)
	if err != nil {
		result.err = fmt.Errorf("reading manifest: %w", err)
		return result
	}

	rc, err := catalog.ParseRaw(rawManifest)
	if err != nil {
		result.err = fmt.Errorf("parsing manifest: %w", err)
		return result
	}

	client, err := transport.NewClient(t.Address)
	if err != nil {
		result.err = err
		return result
	}
	defer func() { _ = client.Close() }()

	r, err := client.ApplyCatalog(ctx, rc, dryRun)
	if err != nil {
		result.err = err
		return result
	}
	result.report = r
	return result
}

func printResults(w interface{ Write(p []byte) (int, error) }, results []pushResult) {
	for _, res := range results {
		fmt.Fprintf(w, "=== %s (%s) ===\n", res.target.Hostname, res.target.Address)
		if res.err != nil {
			fmt.Fprintf(w, "ERROR: %v\n\n", res.err)
			continue
		}
		fmt.Fprintln(w, res.report)
	}
}

// exitCode mirrors the agent's exit code scheme but aggregates across hosts:
//   - 2 if any host returned a transport error or any resource failed
//   - 3 if any host had pending changes (dry-run)
//   - 0 otherwise
func exitCode(results []pushResult) int {
	anySkipped := false
	for _, res := range results {
		if res.err != nil {
			return 2
		}
		for _, rr := range res.report.ResourceReports {
			switch rr.State {
			case resource.Failed:
				return 2
			case resource.Skipped:
				anySkipped = true
			}
		}
	}
	if anySkipped {
		return 3
	}
	return 0
}
