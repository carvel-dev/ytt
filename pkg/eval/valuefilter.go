package eval

import (
	"fmt"

	"github.com/k14s/ytt/pkg/yamlmeta"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
	"go.starlark.net/starlark"
)

// NewValuesFilter takes AST values and return a filter function
func NewValuesFilter(argsvalues interface{}) ValuesFilter {
	return func(values interface{}) (interface{}, error) {
		op := yttoverlay.OverlayOp{
			Left:   yamlmeta.NewASTFromInterface(values),
			Right:  argsvalues,
			Thread: &starlark.Thread{Name: "data-values-overlay-pre-processing"},
		}

		newLeft, err := op.Apply()
		if err != nil {
			return nil, fmt.Errorf("Overlaying data values from eval() provided values on top of data values (marked as @data/values): %s", err)
		}

		return (&yamlmeta.Document{Value: newLeft}).AsInterface(yamlmeta.InterfaceConvertOpts{}), nil
	}
}
