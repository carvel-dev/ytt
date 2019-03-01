package core

import (
	"fmt"

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

func (e GoValue) AsValueWithCheckedMapKeys(forceMapStringKeys bool) interface{} {
	return e.convertMaps(e.val, forceMapStringKeys)
}

func (e GoValue) asStarlarkValue(val interface{}) starlark.Value {
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

	case map[interface{}]interface{}:
		return e.dictAsStarlarkValue(typedVal)

	case []interface{}:
		return e.listAsStarlarkValue(typedVal)

	default:
		panic(fmt.Sprintf("unknown type %T for conversion to starlark value", val))
	}
}

func (e GoValue) dictAsStarlarkValue(val map[interface{}]interface{}) starlark.Value {
	if e.mapIsStruct {
		data := map[string]starlark.Value{}
		for k, v := range val {
			if keyStr, ok := k.(string); ok {
				data[keyStr] = e.asStarlarkValue(v)
			} else {
				panic(fmt.Sprintf("expected struct key %s to be string", k)) // TODO
			}
		}
		return &StarlarkStruct{data}
	}

	result := &starlark.Dict{}
	for k, v := range val {
		result.SetKey(e.asStarlarkValue(k), e.asStarlarkValue(v))
	}
	return result
}

func (e GoValue) listAsStarlarkValue(val []interface{}) *starlark.List {
	result := []starlark.Value{}
	for _, v := range val {
		result = append(result, e.asStarlarkValue(v))
	}
	return starlark.NewList(result)
}

func (e GoValue) convertMaps(val interface{}, forceMapStringKeys bool) interface{} {
	switch typedVal := val.(type) {
	case []interface{}:
		for i, v := range typedVal {
			typedVal[i] = e.convertMaps(v, forceMapStringKeys)
		}
		return typedVal

	case map[string]interface{}:
		if forceMapStringKeys {
			// keep type as is
			for k, v := range typedVal {
				typedVal[k] = e.convertMaps(v, forceMapStringKeys)
			}
			return typedVal
		}

		result := map[interface{}]interface{}{}
		for k, v := range typedVal {
			result[k] = e.convertMaps(v, forceMapStringKeys)
		}
		return result

	case map[interface{}]interface{}:
		if forceMapStringKeys {
			result := map[string]interface{}{}
			for k, v := range typedVal {
				strK, ok := k.(string)
				if !ok {
					panic("expected key to be string")
				}
				result[strK] = e.convertMaps(v, forceMapStringKeys)
			}
			return result
		}

		// keep type as is
		for k, v := range typedVal {
			typedVal[k] = e.convertMaps(v, forceMapStringKeys)
		}
		return typedVal

	default:
		return val
	}
}
