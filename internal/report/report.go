package report

import (
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
