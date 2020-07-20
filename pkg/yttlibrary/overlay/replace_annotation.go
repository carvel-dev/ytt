package overlay

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

type ReplaceAnnotation struct {
	newNode template.EvaluationNode
	thread  *starlark.Thread
	via     *starlark.Value
}

func NewReplaceAnnotation(newNode template.EvaluationNode, thread *starlark.Thread) (ReplaceAnnotation, error) {
	annotation := ReplaceAnnotation{
		newNode: newNode,
		thread:  thread,
	}
	kwargs := template.NewAnnotations(newNode).Kwargs(AnnotationReplace)

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))
		switch kwargName {
		case "via":
			annotation.via = &kwarg[1]
		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationReplace, kwargName)
		}
	}

	return annotation, nil
}

func (a ReplaceAnnotation) Value(existingNode template.EvaluationNode) (interface{}, error) {
	// Make sure original nodes are not affected in any way
	existingNode = existingNode.DeepCopyAsInterface().(template.EvaluationNode)
	newNode := a.newNode.DeepCopyAsInterface().(template.EvaluationNode)

	// TODO currently assumes that we can always get at least one value
	if a.via == nil {
		return newNode.GetValues()[0], nil
	}

	switch typedVal := (*a.via).(type) {
	case starlark.Callable:
		viaArgs := starlark.Tuple{
			yamltemplate.NewGoValueWithYAML(existingNode.GetValues()[0]).AsStarlarkValue(),
			yamltemplate.NewGoValueWithYAML(newNode.GetValues()[0]).AsStarlarkValue(),
		}

		// TODO check thread correctness
		result, err := starlark.Call(a.thread, *a.via, viaArgs, []starlark.Tuple{})
		if err != nil {
			return nil, err
		}

		return tplcore.NewStarlarkValue(result).AsGoValue(), nil

	default:
		return nil, fmt.Errorf("Expected '%s' annotation keyword argument 'via'"+
			" to be function, but was %T", AnnotationReplace, typedVal)
	}
}
