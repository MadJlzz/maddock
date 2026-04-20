package engine

import (
	"context"
	"log/slog"
	"time"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/report"
	"github.com/MadJlzz/maddock/internal/resource"
)

func Run(ctx context.Context, c *catalog.Catalog, dryRun bool) *report.Report {
	slog.Info("applying catalog", "name", c.Name, "resources", len(c.Resources), "dry_run", dryRun)

	r := &report.Report{
		Name: c.Name,
	}

	for _, res := range c.Resources {
		log := slog.With("type", res.Type(), "name", res.Name())
		log.Debug("checking resource")

		start := time.Now()
		rr := report.ResourceReport{
			Type: res.Type(),
			Name: res.Name(),
		}

		checkResult, err := res.Check(ctx)
		if err != nil {
			log.Warn("check failed", "error", err)
			rr.State = resource.Failed
			rr.Error = &resource.ResourceError{
				Type:  res.Type(),
				Name:  res.Name(),
				Phase: resource.PhaseCheck,
				Err:   err,
			}
			rr.Duration = time.Since(start)
			r.ResourceReports = append(r.ResourceReports, rr)
			continue
		}

		if !checkResult.Changed {
			log.Debug("already in desired state")
			rr.State = resource.Ok
			rr.Duration = time.Since(start)
			r.ResourceReports = append(r.ResourceReports, rr)
			continue
		}

		rr.Differences = checkResult.Differences
		log.Debug("changes detected", "differences", len(rr.Differences))

		if dryRun {
			log.Debug("dry-run: skipping apply")
			rr.State = resource.Skipped
			rr.Duration = time.Since(start)
			r.ResourceReports = append(r.ResourceReports, rr)
			continue
		}

		applyResult, err := res.Apply(ctx)
		if err != nil {
			log.Warn("apply failed", "error", err)
			rr.State = resource.Failed
			rr.Error = &resource.ResourceError{
				Type:  res.Type(),
				Name:  res.Name(),
				Phase: resource.PhaseApply,
				Err:   err,
			}
			rr.Duration = time.Since(start)
			r.ResourceReports = append(r.ResourceReports, rr)
			continue
		}

		log.Info("applied", "state", applyResult.Result)
		rr.State = applyResult.Result
		rr.Duration = time.Since(start)
		r.ResourceReports = append(r.ResourceReports, rr)
	}

	return r
}
