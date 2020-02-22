package yamlmeta

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

func NewASTFromInterface(val interface{}) interface{} {
	return convertToAST(val)
}

func NewGoFromAST(val interface{}) interface{} {
	return convertToGo(val)
}

func convertToLowYAML(val interface{}) interface{} {
	switch typedVal := val.(type) {
	case map[interface{}]interface{}:
		panic("Expected *orderedmap.Map instead of map[interface{}]interface{} in convertToLowYAML")

	case map[string]interface{}:
		panic("Expected *orderedmap.Map instead of map[string]interface{} in convertToLowYAML")

	case *orderedmap.Map:
		result := yaml.MapSlice{}
		typedVal.Iterate(func(k, v interface{}) {
			result = append(result, yaml.MapItem{
				Key:   k,
				Value: convertToLowYAML(v),
			})
		})
		return result

	case []interface{}:
		result := []interface{}{}
		for _, item := range typedVal {
			result = append(result, convertToLowYAML(item))
		}
		return result

	default:
		return val
	}
}

func convertToGo(val interface{}) interface{} {
	switch typedVal := val.(type) {
	case *DocumentSet:
		panic("Unexpected docset value within document")

	case *Document:
		panic("Unexpected document within document")

	case *Map:
		result := orderedmap.NewMap()
		for _, item := range typedVal.Items {
			// Catch any cases where unique key invariant is violated
			if _, found := result.Get(item.Key); found {
				panic(fmt.Sprintf("Unexpected duplicate key: %s", item.Key))
			}
			result.Set(item.Key, convertToGo(item.Value))
		}
		return result

	case *Array:
		result := []interface{}{}
		for _, item := range typedVal.Items {
			result = append(result, convertToGo(item.Value))
		}
		return result

	case []interface{}:
		result := []interface{}{}
		for _, item := range typedVal {
			result = append(result, convertToGo(item))
		}
		return result

	case map[interface{}]interface{}:
		panic("Expected *orderedmap.Map instead of map[interface{}]interface{} in convertToGo")

	case map[string]interface{}:
		panic("Expected *orderedmap.Map instead of map[string]interface{} in convertToGo")

	case *orderedmap.Map:
		result := orderedmap.NewMap()
		typedVal.Iterate(func(k, v interface{}) {
			result.Set(k, convertToGo(v))
		})
		return result

	default:
		return val
	}
}

func convertToAST(val interface{}) interface{} {
	switch typedVal := val.(type) {
	// necessary for overlay processing
	case []*DocumentSet:
		for i, item := range typedVal {
			typedVal[i] = convertToAST(item).(*DocumentSet)
		}
		return typedVal

	case *DocumentSet:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToAST(item).(*Document)
		}
		return typedVal

	case *Document:
		typedVal.Value = convertToAST(typedVal.Value)
		return typedVal

	case *Map:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToAST(item).(*MapItem)
		}
		return typedVal

	case *MapItem:
		typedVal.Key = convertToAST(typedVal.Key)
		typedVal.Value = convertToAST(typedVal.Value)
		return typedVal

	case *Array:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToAST(item).(*ArrayItem)
		}
		return typedVal

	case *ArrayItem:
		typedVal.Value = convertToAST(typedVal.Value)
		return typedVal

	case []interface{}:
		result := &Array{}
		for _, item := range typedVal {
			result.Items = append(result.Items, &ArrayItem{
				Value:    convertToAST(item),
				Position: filepos.NewUnknownPosition(),
			})
		}
		return result

	case map[interface{}]interface{}:
		panic("Expected *orderedmap.Map instead of map[interface{}]interface{} in convertToAST")

	case map[string]interface{}:
		panic("Expected *orderedmap.Map instead of map[string]interface{} in convertToAST")

	case *orderedmap.Map:
		result := &Map{Position: filepos.NewUnknownPosition()}
		typedVal.Iterate(func(k, v interface{}) {
			result.Items = append(result.Items, &MapItem{
				Key:      k,
				Value:    convertToAST(v),
				Position: filepos.NewUnknownPosition(),
			})
		})
		return result

	default:
		return val
	}
}
