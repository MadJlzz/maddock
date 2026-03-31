package resource

import "context"

type Resource interface {
	Type() string
	Name() string
	Check(ctx context.Context) (*CheckResult, error)
	Apply(ctx context.Context) (*ApplyResult, error)
}

type State string

const (
	Ok      State = "OK"
	Changed State = "CHANGED"
	Failed  State = "FAILED"
	Skipped State = "SKIPPED"
)

type Difference struct {
	Attribute string
	Current   string
	Desired   string
}

type CheckResult struct {
	Changed     bool
	Differences []Difference
}

type ApplyResult struct {
	Result State
}
