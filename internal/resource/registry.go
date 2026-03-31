package resource

import "fmt"

type ParseFn func(name string, attrs map[string]any) (Resource, error)

var registry = map[string]ParseFn{}

func Register(kind string, parseFn ParseFn) {
	registry[kind] = parseFn
}

func Parse(kind string, name string, attrs map[string]any) (Resource, error) {
	parseFn, ok := registry[kind]
	if !ok {
		return nil, fmt.Errorf("unknown resource kind: %s", kind)
	}
	return parseFn(name, attrs)
}
