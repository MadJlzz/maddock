package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MadJlzz/maddock/internal/resource"
)

type ResourceReport struct {
	Type        string
	Name        string
	State       resource.State
	Differences []resource.Difference
	Error       error
	Duration    time.Duration
}

type Report struct {
	Name            string
	ResourceReports []ResourceReport
}

func formatDifferences(diffs []resource.Difference) string {
	if len(diffs) == 0 {
		return ""
	}
	attrs := make([]string, len(diffs))
	for i, d := range diffs {
		attrs[i] = d.Attribute
	}
	return "(" + strings.Join(attrs, ", ") + ")"
}

func (r *Report) String() string {
	const padWidth = 40
	builder := strings.Builder{}

	fmt.Fprintf(&builder, "Maddock Agent — applying: %s\n", r.Name)
	fmt.Fprintf(&builder, "════════════════════════════════════\n")

	stateCounter := make(map[resource.State]int)
	var totalDuration time.Duration
	for i, rr := range r.ResourceReports {
		label := fmt.Sprintf("%s:%s", rr.Type, rr.Name)
		dots := strings.Repeat(".", max(1, padWidth-len(label)))
		detail := formatDifferences(rr.Differences)
		if rr.Error != nil {
			detail = "(" + rr.Error.Error() + ")"
		}
		fmt.Fprintf(&builder, "[%d/%d] %s %s %s %s\n", i+1, len(r.ResourceReports), label, dots, rr.State, detail)
		stateCounter[rr.State]++
		totalDuration += rr.Duration
	}

	fmt.Fprintf(&builder, "\nSummary: %d changed, %d ok, %d failed | %s\n", stateCounter[resource.Changed], stateCounter[resource.Ok], stateCounter[resource.Failed], totalDuration)
	return builder.String()
}

// --- JSON serialization ---

type errorJSON struct {
	Type    string         `json:"type,omitempty"`
	Name    string         `json:"name,omitempty"`
	Phase   resource.Phase `json:"phase,omitempty"`
	Message string         `json:"message"`
}

type resourceReportJSON struct {
	Type        string                `json:"type"`
	Name        string                `json:"name"`
	State       resource.State        `json:"state"`
	Differences []resource.Difference `json:"differences,omitempty"`
	DurationMs  int64                 `json:"duration_ms"`
	Error       *errorJSON            `json:"error,omitempty"`
}

func (rr ResourceReport) MarshalJSON() ([]byte, error) {
	out := resourceReportJSON{
		Type:        rr.Type,
		Name:        rr.Name,
		State:       rr.State,
		Differences: rr.Differences,
		DurationMs:  rr.Duration.Milliseconds(),
	}
	if rr.Error != nil {
		ej := &errorJSON{Message: rr.Error.Error()}
		// If the error is a ResourceError, pull structured fields out.
		var rerr *resource.ResourceError
		if errors.As(rr.Error, &rerr) {
			ej.Type = rerr.Type
			ej.Name = rerr.Name
			ej.Phase = rerr.Phase
			ej.Message = rerr.Err.Error()
		}
		out.Error = ej
	}
	return json.Marshal(out)
}

type summaryJSON struct {
	Ok         int   `json:"ok"`
	Changed    int   `json:"changed"`
	Failed     int   `json:"failed"`
	Skipped    int   `json:"skipped"`
	DurationMs int64 `json:"duration_ms"`
}

type reportJSON struct {
	Name      string           `json:"name"`
	Resources []ResourceReport `json:"resources"`
	Summary   summaryJSON      `json:"summary"`
}

func (r *Report) MarshalJSON() ([]byte, error) {
	counts := make(map[resource.State]int)
	var total time.Duration
	for _, rr := range r.ResourceReports {
		counts[rr.State]++
		total += rr.Duration
	}
	return json.Marshal(reportJSON{
		Name:      r.Name,
		Resources: r.ResourceReports,
		Summary: summaryJSON{
			Ok:         counts[resource.Ok],
			Changed:    counts[resource.Changed],
			Failed:     counts[resource.Failed],
			Skipped:    counts[resource.Skipped],
			DurationMs: total.Milliseconds(),
		},
	})
}

// ExitCode returns the appropriate process exit code:
//   - 0: converged (all OK or CHANGED)
//   - 2: one or more resources failed
//   - 3: dry-run found changes (SKIPPED)
func (r *Report) ExitCode() int {
	for _, rr := range r.ResourceReports {
		if rr.State == resource.Failed {
			return 2
		}
	}
	for _, rr := range r.ResourceReports {
		if rr.State == resource.Skipped {
			return 3
		}
	}
	return 0
}
