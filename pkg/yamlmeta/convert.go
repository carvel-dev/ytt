package yamlmeta

import (
	"fmt"
	"sort"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

type InterfaceConvertOpts struct {
	OrderedMap bool
}

func NewASTFromInterface(val interface{}) interface{} {
	return convertToAST(val)
}

func convertToLowYAML(val interface{}, opts InterfaceConvertOpts) interface{} {
	switch typedVal := val.(type) {
	case *DocumentSet:
		panic("Unexpected docset value within document")

	case *Document:
		panic("Unexpected document within document")

	case *Map:
		if opts.OrderedMap {
			result := yaml.MapSlice{}
			itemIndex := map[interface{}]int{}
			for idx, item := range typedVal.Items {
				// TODO do not blindly override previously set items
				if prevIdx, ok := itemIndex[item.Key]; ok {
					result[prevIdx].Value = convertToLowYAML(item.Value, opts)
				} else {
					itemIndex[item.Key] = idx
					result = append(result, yaml.MapItem{
						Key:   item.Key,
						Value: convertToLowYAML(item.Value, opts),
					})
				}
			}
			return result
		}

		result := map[interface{}]interface{}{}
		for _, item := range typedVal.Items {
			result[item.Key] = convertToLowYAML(item.Value, opts)
		}
		return result

	case *Array:
		result := []interface{}{}
		for _, item := range typedVal.Items {
			result = append(result, convertToLowYAML(item.Value, opts))
		}
		return result

	case []interface{}:
		result := []interface{}{}
		for _, item := range typedVal {
			result = append(result, convertToLowYAML(item, opts))
		}
		return result

	case map[interface{}]interface{}:
		// TODO ideally we wouldnt have to be sorting maps here
		// as it implies that upstream code didnt preserve order
		if opts.OrderedMap {
			result := yaml.MapSlice{}
			for _, key := range sortedMapKeys(typedVal) {
				result = append(result, yaml.MapItem{
					Key:   key,
					Value: convertToLowYAML(typedVal[key], opts),
				})
			}
			return result
		}
		return val

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
		result := &Map{Position: filepos.NewUnknownPosition()}
		for _, k := range sortedMapKeys(typedVal) {
			result.Items = append(result.Items, &MapItem{
				Key:      k,
				Value:    convertToAST(typedVal[k]),
				Position: filepos.NewUnknownPosition(),
			})
		}
		return result

	default:
		return val
	}
}

func sortedMapKeys(m map[interface{}]interface{}) []interface{} {
	var keys []interface{}
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		iStr := fmt.Sprintf("%s", keys[i])
		jStr := fmt.Sprintf("%s", keys[j])
		return iStr < jStr
	})
	return keys
}
