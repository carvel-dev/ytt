package core

import (
	"fmt"

	"go.starlark.net/starlark"
)

type StarlarkStruct struct {
	data map[string]starlark.Value
}

var _ starlark.Value = &StarlarkStruct{}
var _ starlark.HasAttrs = &StarlarkStruct{}

func (s *StarlarkStruct) String() string        { return "struct(...)" }
func (s *StarlarkStruct) Type() string          { return "struct" }
func (s *StarlarkStruct) Freeze()               {} // TODO
func (s *StarlarkStruct) Truth() starlark.Bool  { return len(s.data) > 0 }
func (s *StarlarkStruct) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: struct") }

// returns (nil, nil) if attribute not present
func (s *StarlarkStruct) Attr(name string) (starlark.Value, error) {
	val, found := s.data[name]
	if !found {
		return nil, nil
	}
	return val, nil
}

// callers must not modify the result.
func (s *StarlarkStruct) AttrNames() []string {
	var keys []string
	for k, _ := range s.data {
		keys = append(keys, k)
	}
	return keys
}
