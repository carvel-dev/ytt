package overlay

import (
	"fmt"
	"reflect"

	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	API = starlark.StringDict{
		"overlay": &starlarkstruct.Module{
			Name: "overlay",
			Members: starlark.StringDict{
				"apply":   starlark.NewBuiltin("overlay.apply", core.ErrWrapper(overlayModule{}.Apply)),
				"index":   starlark.NewBuiltin("overlay.index", core.ErrWrapper(overlayModule{}.Index)),
				"all":     starlark.NewBuiltin("overlay.all", core.ErrWrapper(overlayModule{}.All)),
				"map_key": starlark.NewBuiltin("overlay.map_key", core.ErrWrapper(overlayModule{}.MapKey)),
				"subset":  starlark.NewBuiltin("overlay.subset", core.ErrWrapper(overlayModule{}.Subset)),
			},
		},
	}
)

type overlayModule struct{}

func (b overlayModule) Apply(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() == 0 {
		return starlark.None, fmt.Errorf("expected exactly at least argument")
	}

	typedVals := core.NewStarlarkValue(args).AsInterface().([]interface{})
	var result interface{} = typedVals[0]

	for _, right := range typedVals[1:] {
		var err error
		result, err = OverlayOp{Left: result, Right: right, Thread: thread}.Apply() // left is modified
		if err != nil {
			return starlark.None, err
		}
	}

	return yamltemplate.NewStarlarkFragment(result), nil
}

func (b overlayModule) Index(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	expectedIdx64, err := core.NewStarlarkValue(args.Index(0)).AsInt64()
	if err != nil {
		return starlark.None, err
	}

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		idx64, err := core.NewStarlarkValue(args.Index(0)).AsInt64()
		if err != nil {
			return starlark.None, err
		}

		if expectedIdx64 == idx64 {
			return starlark.Bool(true), nil
		}

		return starlark.Bool(false), nil
	}

	return starlark.NewBuiltin("overlay.index_matcher", core.ErrWrapper(matchFunc)), nil
}

func (b overlayModule) All(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 3 {
		return starlark.None, fmt.Errorf("expected exactly 3 arguments")
	}

	return starlark.Bool(true), nil
}

func (b overlayModule) MapKey(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	keyName, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		oldVal := core.NewStarlarkValue(args.Index(1)).AsInterface()
		newVal := core.NewStarlarkValue(args.Index(2)).AsInterface()

		result, err := b.compareByMapKey(keyName, oldVal, newVal)
		if err != nil {
			return nil, err
		}

		return starlark.Bool(result), nil
	}

	return starlark.NewBuiltin("overlay.map_key_matcher", core.ErrWrapper(matchFunc)), nil
}

func (b overlayModule) compareByMapKey(keyName string, oldVal, newVal interface{}) (bool, error) {
	oldKeyVal, err := b.pullOutMapValue(keyName, oldVal)
	if err != nil {
		return false, err
	}

	newKeyVal, err := b.pullOutMapValue(keyName, newVal)
	if err != nil {
		return false, err
	}

	result, _ := Comparison{}.CompareLeafs(oldKeyVal, newKeyVal)
	return result, nil
}

func (b overlayModule) pullOutMapValue(keyName string, newVal interface{}) (interface{}, error) {
	typedArrayItem, ok := newVal.(*yamlmeta.ArrayItem)
	if !ok {
		return starlark.None, fmt.Errorf("Expected new item to be arrayitem, but was %T", newVal)
	}

	typedMap, ok := typedArrayItem.Value.(*yamlmeta.Map)
	if !ok {
		return starlark.None, fmt.Errorf("Expected arrayitem to contain map, but was %T", typedArrayItem.Value)
	}

	for _, item := range typedMap.Items {
		if reflect.DeepEqual(item.Key, keyName) {
			return item.Value, nil
		}
	}

	return starlark.None, fmt.Errorf("Expected to find mapitem with key %s, but did not", keyName)
}

func (b overlayModule) Subset(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	expectedArg := args.Index(0)

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		leftVal := core.NewStarlarkValue(args.Index(1)).AsInterface()
		expectedVal := core.NewStarlarkValue(expectedArg).AsInterface()

		actualObj := yamlmeta.NewASTFromInterface(leftVal)
		expectedObj := yamlmeta.NewASTFromInterface(expectedVal)

		if _, ok := actualObj.(*yamlmeta.ArrayItem); ok {
			expectedObj = &yamlmeta.ArrayItem{Value: expectedObj}
		}
		if _, ok := actualObj.(*yamlmeta.Document); ok {
			expectedObj = &yamlmeta.Document{Value: expectedObj}
		}

		result, _ := Comparison{}.Compare(actualObj, expectedObj)
		return starlark.Bool(result), nil
	}

	return starlark.NewBuiltin("overlay.subset_matcher", core.ErrWrapper(matchFunc)), nil
}
