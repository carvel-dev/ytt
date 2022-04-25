// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package assertions

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamltemplate"
)

// A Rule represents a validation attached to a Node via an annotation;
// it contains a string description of what constitutes a valid value,
// and a function that asserts the rule against an actual value.
type Rule struct {
	Msg       string
	Assertion starlark.Callable
	Position  *filepos.Position
}

// Validate runs the assertion from the Rule with the node's value as arguments.
//
// Returns an error if the assertion returns False (not-None), or assert.fail()s.
// Otherwise, returns nil.
func (r Rule) Validate(node yamlmeta.Node, thread *starlark.Thread) error {
	var key string
	var nodeValue starlark.Value
	switch typedNode := node.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
		panic(fmt.Sprintf("validation at %s - not supported on %s at %s", r.Position.AsCompactString(), yamlmeta.TypeName(node), node.GetPosition().AsCompactString()))
	case *yamlmeta.MapItem:
		key = fmt.Sprintf("%q", typedNode.Key)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	case *yamlmeta.ArrayItem:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	case *yamlmeta.Document:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	}

	result, err := starlark.Call(thread, r.Assertion, starlark.Tuple{nodeValue}, []starlark.Tuple{})
	if err != nil {
		return fmt.Errorf("%s (%s) requires %q; %s (by %s)", key, node.GetPosition().AsCompactString(), r.Msg, err.Error(), r.Position.AsCompactString())
	}

	// in order to pass, the assertion must return True or None
	if _, ok := result.(starlark.NoneType); !ok {
		if !result.Truth() {
			return fmt.Errorf("%s (%s) requires %q (by %s)", key, node.GetPosition().AsCompactString(), r.Msg, r.Position.AsCompactString())
		}
	}

	return nil
}
