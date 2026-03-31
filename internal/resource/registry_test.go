package resource

import (
	"context"
	"testing"
)

type dummyResource struct {
	name string
}

func (dr *dummyResource) Type() string {
	return "dummyType"
}

func (dr *dummyResource) Name() string {
	return dr.name
}

func (dr *dummyResource) Check(ctx context.Context) (*CheckResult, error) {
	return &CheckResult{}, nil
}

func (dr *dummyResource) Apply(ctx context.Context) (*ApplyResult, error) {
	return &ApplyResult{}, nil
}

func cleanRegistry() {
	registry = make(map[string]ParseFn)
}

func TestRegisterAndParseExistingResource(t *testing.T) {
	t.Cleanup(cleanRegistry)
	Register("dummyType", func(name string, attrs map[string]any) (Resource, error) {
		return &dummyResource{name: name}, nil
	})

	dr, err := Parse("dummyType", "dummyName", map[string]any{})
	if err != nil {
		t.Fatalf("failed to parse resource: %s", err)
	}

	if dr.Type() != "dummyType" {
		t.Errorf("failed to pa"+
			"rse resource, expected dummyType but got %s", dr.Type())
	}

	if dr.Name() != "dummyName" {
		t.Errorf("failed to parse resource, expected dummyName but got %s", dr.Name())
	}
}

func TestParseUnknownResource(t *testing.T) {
	t.Cleanup(cleanRegistry)
	_, err := Parse("unknown", "dummyName", map[string]any{})
	if err == nil {
		t.Fatal("an error is expected here")
	}

	if err.Error() != "unknown resource kind: unknown" {
		t.Errorf("wanted 'unknown resource kind: unknown', got %s", err.Error())
	}
}
