package yamlmeta

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
)

// A Rule represents an argument to an @assert/validate annotation;
// it contains a string description of what constitutes a valid value,
// and a function that asserts the rule against an actual value.
// One @assert/validate annotation can have multiple Rules.
type Rule struct {
	Msg       string
	Assertion starlark.Callable
	Position  *filepos.Position
}

// Validate runs the assertion from the Rule with the node's value as arguments.
//
// Returns an error if the assertion returns False (not-None), or assert.fail()s.
// Otherwise, returns nil.
func (r Rule) Validate(node Node, thread *starlark.Thread) error {
	var key string
	var nodeValue starlark.Value
	switch typedNode := node.(type) {
	case *DocumentSet, *Array, *Map:
		// TODO: is this the correct place for as specific of an error
		panic(fmt.Sprintf("@%s annotation at %s - not supported on %s at %s", "assert/validate", r.Position.AsCompactString(), TypeName(node), node.GetPosition().AsCompactString()))
	case *MapItem:
		key = fmt.Sprintf("%q", typedNode.Key)
		nodeValue = NewGoValueFromAST(typedNode.Value).AsStarlarkValue()
	case *ArrayItem:
		key = TypeName(typedNode)
		nodeValue = NewGoValueFromAST(typedNode.Value).AsStarlarkValue()
	case *Document:
		key = TypeName(typedNode)
		nodeValue = NewGoValueFromAST(typedNode.Value).AsStarlarkValue()
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
