package core

import (
	"fmt"

	"github.com/k14s/ytt/pkg/orderedmap"
	"go.starlark.net/starlark"
)

type GoValueToStarlarkValueConversion interface {
	AsStarlarkValue() starlark.Value
}

type GoValue struct {
	val         interface{}
	mapIsStruct bool
}

func NewGoValue(val interface{}, mapIsStruct bool) GoValue {
	return GoValue{val, mapIsStruct}
}

func (e GoValue) AsStarlarkValue() starlark.Value {
	return e.asStarlarkValue(e.val)
}

func (e GoValue) asStarlarkValue(val interface{}) starlark.Value {
	if obj, ok := val.(starlark.Value); ok {
		return obj
	}
	if obj, ok := val.(GoValueToStarlarkValueConversion); ok {
		return obj.AsStarlarkValue()
	}

	switch typedVal := val.(type) {
	case nil:
		return starlark.None // TODO is it nil or is it None

	case bool:
		return starlark.Bool(typedVal)

	case string:
		return starlark.String(typedVal)

	case int:
		return starlark.MakeInt(typedVal)

	case int64:
		return starlark.MakeInt64(typedVal)

	case uint:
		return starlark.MakeUint(typedVal)

	case uint64:
		return starlark.MakeUint64(typedVal)

	case float64:
		return starlark.Float(typedVal)

	case map[string]interface{}:
		panic("Expected *orderedmap.Map instead of map[string]interface{} for conversion to starlark value")

	case map[interface{}]interface{}:
		panic("Expected *orderedmap.Map instead of map[interface{}]interface{} for conversion to starlark value")

	case *orderedmap.Map:
		return e.dictAsStarlarkValue(typedVal)

	case []interface{}:
		return e.listAsStarlarkValue(typedVal)

	case *starlark.Builtin:
		return typedVal

	default:
		panic(fmt.Sprintf("unknown type %T for conversion to starlark value", val))
	}
}

func (e GoValue) dictAsStarlarkValue(val *orderedmap.Map) starlark.Value {
	if e.mapIsStruct {
		data := orderedmap.NewMap()
		val.Iterate(func(k, v interface{}) {
			if keyStr, ok := k.(string); ok {
				data.Set(keyStr, e.asStarlarkValue(v))
			} else {
				panic(fmt.Sprintf("expected struct key %s to be string", k)) // TODO
			}
		})
		return &StarlarkStruct{data}
	}

	result := &starlark.Dict{}
	val.Iterate(func(k, v interface{}) {
		result.SetKey(e.asStarlarkValue(k), e.asStarlarkValue(v))
	})
	return result
}

func (e GoValue) listAsStarlarkValue(val []interface{}) *starlark.List {
	result := []starlark.Value{}
	for _, v := range val {
		result = append(result, e.asStarlarkValue(v))
	}
	return starlark.NewList(result)
}
