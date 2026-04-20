package resource

import (
	"context"
	"fmt"
)

// Phase identifies which stage of a resource's lifecycle produced an error.
type Phase string

const (
	PhaseCheck Phase = "check"
	PhaseApply Phase = "apply"
)

// ResourceError wraps an error produced while processing a single resource,
// carrying the resource's identity and the phase where the error occurred.
// It implements the errors.Unwrap contract so callers can use errors.Is /
// errors.As to inspect the underlying cause.
type ResourceError struct {
	Type  string
	Name  string
	Phase Phase
	Err   error
}

func (e *ResourceError) Error() string {
	return fmt.Sprintf("%s:%s: %s phase: %v", e.Type, e.Name, e.Phase, e.Err)
}

func (e *ResourceError) Unwrap() error {
	return e.Err
}

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
	Attribute string `json:"attribute"`
	Current   string `json:"current"`
	Desired   string `json:"desired"`
}

type CheckResult struct {
	Changed     bool
	Differences []Difference
}

type ApplyResult struct {
	Result State
}
