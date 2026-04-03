package report

import (
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
		fmt.Fprintf(&builder, "[%d/%d] %s %s %s %s\n", i+1, len(r.ResourceReports), label, dots, rr.State, formatDifferences(rr.Differences))
		stateCounter[rr.State]++
		totalDuration += rr.Duration
	}

	fmt.Fprintf(&builder, "\nSummary: %d changed, %d ok, %d failed | %s\n", stateCounter[resource.Changed], stateCounter[resource.Ok], stateCounter[resource.Failed], totalDuration)
	return builder.String()
}
