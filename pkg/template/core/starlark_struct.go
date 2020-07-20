package core

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/orderedmap"
)

type StarlarkStruct struct {
	data *orderedmap.Map
}

var _ starlark.Value = &StarlarkStruct{}
var _ starlark.HasAttrs = &StarlarkStruct{}

func (s *StarlarkStruct) String() string        { return "struct(...)" }
func (s *StarlarkStruct) Type() string          { return "struct" }
func (s *StarlarkStruct) Freeze()               {} // TODO
func (s *StarlarkStruct) Truth() starlark.Bool  { return s.data.Len() > 0 }
func (s *StarlarkStruct) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: struct") }

// returns (nil, nil) if attribute not present
func (s *StarlarkStruct) Attr(name string) (starlark.Value, error) {
	val, found := s.data.Get(name)
	if found {
		return val.(starlark.Value), nil
	}
	return nil, nil
}

// callers must not modify the result.
func (s *StarlarkStruct) AttrNames() []string {
	var keys []string
	s.data.Iterate(func(key, _ interface{}) {
		keys = append(keys, key.(string))
	})
	return keys
}
