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

	return reflect.DeepEqual(oldKeyVal, newKeyVal), nil
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

		result, _ := b.compareSubset(actualObj, expectedObj)
		return starlark.Bool(result), nil
	}

	return starlark.NewBuiltin("overlay.subset_matcher", core.ErrWrapper(matchFunc)), nil
}

func (b *overlayModule) compareSubset(left, right interface{}) (bool, string) {
	switch typedRight := right.(type) {
	case *yamlmeta.DocumentSet:
		panic("Unexpected docset")

	case *yamlmeta.Document:
		typedLeft, isDoc := left.(*yamlmeta.Document)
		if !isDoc {
			return false, fmt.Sprintf("Expected doc, but was %T", left)
		}

		return b.compareSubset(typedLeft.Value, typedRight.Value)

	case *yamlmeta.Map:
		typedLeft, isMap := left.(*yamlmeta.Map)
		if !isMap {
			return false, fmt.Sprintf("Expected map, but was %T", left)
		}

		for _, rightItem := range typedRight.Items {
			matched := false
			for _, leftItem := range typedLeft.Items {
				if reflect.DeepEqual(leftItem.Key, rightItem.Key) {
					result, explain := b.compareSubset(leftItem, rightItem)
					if !result {
						return false, explain
					}
					matched = true
				}
			}
			if !matched {
				return false, "Expected at least one map item to match by key"
			}
		}

		return true, ""

	case *yamlmeta.MapItem:
		typedLeft, isMapItem := left.(*yamlmeta.MapItem)
		if !isMapItem {
			return false, fmt.Sprintf("Expected mapitem, but was %T", left)
		}

		return b.compareSubset(typedLeft.Value, typedRight.Value)

	case *yamlmeta.Array:
		typedLeft, isArray := left.(*yamlmeta.Array)
		if !isArray {
			return false, fmt.Sprintf("Expected array, but was %T", left)
		}

		for i, item := range typedRight.Items {
			if i >= len(typedLeft.Items) {
				return false, "Expected to have matching number of array items"
			}
			result, explain := b.compareSubset(typedLeft.Items[i].Value, item.Value)
			if !result {
				return false, explain
			}
		}

		return true, ""

	case *yamlmeta.ArrayItem:
		typedLeft, isArrayItem := left.(*yamlmeta.ArrayItem)
		if !isArrayItem {
			return false, fmt.Sprintf("Expected arrayitem, but was %T", left)
		}

		return b.compareSubset(typedLeft.Value, typedRight.Value)

	default:
		return b.compareLeafs(left, right)
	}
}

func (b *overlayModule) compareLeafs(left, right interface{}) (bool, string) {
	if reflect.DeepEqual(left, right) {
		return true, ""
	}

	if result, _ := b.compareAsInt64s(left, right); result {
		return true, ""
	}

	return false, fmt.Sprintf("Expected leaf values to match %T %T", left, right)
}

func (b *overlayModule) compareAsInt64s(left, right interface{}) (bool, string) {
	leftVal, ok := b.upcastToInt64(left)
	if !ok {
		return false, "Left obj is upcastable to int64"
	}

	rightVal, ok := b.upcastToInt64(right)
	if !ok {
		return false, "Right obj is upcastable to int64"
	}

	return leftVal == rightVal, "Left and right numbers are not equal"
}

func (b *overlayModule) upcastToInt64(val interface{}) (int64, bool) {
	switch typedVal := val.(type) {
	case int:
		return int64(typedVal), true
	case int16:
		return int64(typedVal), true
	case int32:
		return int64(typedVal), true
	case int64:
		return int64(typedVal), true
	case int8:
		return int64(typedVal), true
	default:
		return 0, false
	}
}
