package yamltemplate

import (
	"fmt"

	tplcore "github.com/k14s/ytt/pkg/template/core"
	"go.starlark.net/starlark"
)

type StarlarkFragment struct {
	data interface{}
}

var _ starlark.Value = &StarlarkFragment{}
var _ tplcore.StarlarkValueToGoValueConversion = &StarlarkFragment{}
var _ tplcore.GoValueToStarlarkValueConversion = &StarlarkFragment{}

func NewStarlarkFragment(data interface{}) *StarlarkFragment {
	return &StarlarkFragment{data}
}

func (s *StarlarkFragment) String() string       { return "yamlfragment(...)" }
func (s *StarlarkFragment) Type() string         { return "yamlfragment" }
func (s *StarlarkFragment) Freeze()              {}              // TODO
func (s *StarlarkFragment) Truth() starlark.Bool { return true } // TODO
func (s *StarlarkFragment) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: yamlfragment")
}

func (s *StarlarkFragment) AsGoValue() interface{}          { return s.data }
func (s *StarlarkFragment) AsStarlarkValue() starlark.Value { return s }
