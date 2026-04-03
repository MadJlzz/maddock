package engine

import (
	"context"
	"time"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/report"
	"github.com/MadJlzz/maddock/internal/resource"
)

func Run(ctx context.Context, c *catalog.Catalog, dryRun bool) *report.Report {
	r := &report.Report{
		Name: c.Name,
	}

	for _, res := range c.Resources {
		start := time.Now()
		rr := report.ResourceReport{
			Type: res.Type(),
			Name: res.Name(),
		}

		checkResult, err := res.Check(ctx)
		if err != nil {
			rr.State = resource.Failed
			rr.Error = err
			rr.Duration = time.Since(start)
			r.ResourceReports = append(r.ResourceReports, rr)
			continue
		}

		if !checkResult.Changed {
			rr.State = resource.Ok
			rr.Duration = time.Since(start)
			r.ResourceReports = append(r.ResourceReports, rr)
			continue
		}

		rr.Differences = checkResult.Differences

		if dryRun {
			rr.State = resource.Skipped
			rr.Duration = time.Since(start)
			r.ResourceReports = append(r.ResourceReports, rr)
			continue
		}

		applyResult, err := res.Apply(ctx)
		if err != nil {
			rr.State = resource.Failed
			rr.Error = err
			rr.Duration = time.Since(start)
			r.ResourceReports = append(r.ResourceReports, rr)
			continue
		}

		rr.State = applyResult.Result
		rr.Duration = time.Since(start)
		r.ResourceReports = append(r.ResourceReports, rr)
	}

	return r
}
