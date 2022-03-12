// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"fmt"
	"strings"

	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/orderedmap"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta/internal/yaml.v2"
	yaml3 "gopkg.in/yaml.v3"
)

func NewASTFromInterface(val interface{}) interface{} {
	return convertToAST(val, filepos.NewUnknownPosition())
}

func NewASTFromInterfaceWithPosition(val interface{}, defaultPosition *filepos.Position) interface{} {
	return convertToAST(val, defaultPosition)
}

func NewASTFromInterfaceWithNoPosition(val interface{}) interface{} {
	return convertToASTWithNoPosition(val)
}

func NewGoFromAST(val interface{}) interface{} {
	return convertToGo(val)
}

func convertToYAML3Node(val interface{}) *yaml3.Node {
	switch typedVal := val.(type) {
	case *Document:
		// head, inline := convertYAML3Comments(typedVal)
		return &yaml3.Node{
			Kind:    yaml3.DocumentNode,
			Content: []*yaml3.Node{convertToYAML3Node(typedVal.Value)},
			// HeadComment: head.String(),
			// LineComment: inline.String(),
		}
	case *Map:
		yamlMap := &yaml3.Node{Kind: yaml3.MappingNode}
		for _, item := range typedVal.Items {
			key := convertToYAML3Node(item.Key)
			value := convertToYAML3Node(item.Value)
			yamlMap.Content = append(yamlMap.Content, key, value)
		}
		return yamlMap
	case *orderedmap.Map:
		yamlMap := &yaml3.Node{Kind: yaml3.MappingNode}
		typedVal.Iterate(func(k, v interface{}) {
			key := convertToYAML3Node(k)
			value := convertToYAML3Node(v)
			yamlMap.Content = append(yamlMap.Content, key, value)
		})
		return yamlMap
	case *Array:
		yamlArray := &yaml3.Node{Kind: yaml3.SequenceNode}
		for _, item := range typedVal.Items {
			yamlArray.Content = append(yamlArray.Content, convertToYAML3Node(item.Value))
		}
		return yamlArray
	case []interface{}:
		yamlArray := &yaml3.Node{Kind: yaml3.SequenceNode}
		for _, item := range typedVal {
			yamlArray.Content = append(yamlArray.Content, convertToYAML3Node(item))
		}
		return yamlArray
	case bool:
		return &yaml3.Node{Kind: yaml3.ScalarNode, Value: fmt.Sprintf("%t", val)}
	case int, int32, int64, uint, uint32, uint64:
		return &yaml3.Node{Kind: yaml3.ScalarNode, Value: fmt.Sprintf("%d", val)}
	case float32, float64:
		return &yaml3.Node{Kind: yaml3.ScalarNode, Value: fmt.Sprintf("%g", val)}
	case nil:
		return &yaml3.Node{Kind: yaml3.ScalarNode, Value: "null"}
	case string:
		if val == "" {
			return &yaml3.Node{Kind: yaml3.ScalarNode, Value: "", Style: yaml3.DoubleQuotedStyle}
		} else {
			return &yaml3.Node{Kind: yaml3.ScalarNode, Value: fmt.Sprintf("%s", val), Tag: "!!str"}
		}
	default:
		panic(fmt.Sprintf("Unexpected type %T = %#v", val, val))
	}
}

func convertYAML3Comments(typedVal Node) (strings.Builder, strings.Builder) {
	head := strings.Builder{}
	inline := strings.Builder{}
	for _, comment := range typedVal.GetComments() {
		if comment.Position.IsNextTo(typedVal.GetPosition()) {
			inline.WriteString(comment.Data)
		} else {
			head.WriteString(comment.Data)
		}
	}
	return head, inline
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
	case *DocumentSet, *Document:
		panic(fmt.Sprintf("Unexpected %T value within %T", val, &Document{}))

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

func convertToAST(val interface{}, defaultPosition *filepos.Position) interface{} {
	switch typedVal := val.(type) {
	// necessary for overlay processing
	case []*DocumentSet:
		for i, item := range typedVal {
			typedVal[i] = convertToAST(item, defaultPosition).(*DocumentSet)
		}
		return typedVal

	case *DocumentSet:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToAST(item, defaultPosition).(*Document)
		}
		return typedVal

	case *Document:
		typedVal.Value = convertToAST(typedVal.Value, defaultPosition)
		return typedVal

	case *Map:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToAST(item, defaultPosition).(*MapItem)
		}
		return typedVal

	case *MapItem:
		typedVal.Key = convertToAST(typedVal.Key, defaultPosition)
		typedVal.Value = convertToAST(typedVal.Value, defaultPosition)
		return typedVal

	case *Array:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToAST(item, defaultPosition).(*ArrayItem)
		}
		return typedVal

	case *ArrayItem:
		typedVal.Value = convertToAST(typedVal.Value, defaultPosition)
		return typedVal

	case []interface{}:
		result := &Array{Position: defaultPosition}
		for _, item := range typedVal {
			result.Items = append(result.Items, &ArrayItem{
				Value:    convertToAST(item, defaultPosition),
				Position: defaultPosition,
			})
		}
		return result

	case map[interface{}]interface{}:
		panic("Expected *orderedmap.Map instead of map[interface{}]interface{} in convertToAST")

	case map[string]interface{}:
		panic("Expected *orderedmap.Map instead of map[string]interface{} in convertToAST")

	case *orderedmap.Map:
		result := &Map{Position: defaultPosition}
		typedVal.Iterate(func(k, v interface{}) {
			result.Items = append(result.Items, &MapItem{
				Key:      k,
				Value:    convertToAST(v, defaultPosition),
				Position: defaultPosition,
			})
		})
		return result

	default:
		return val
	}
}

func convertToASTWithNoPosition(val interface{}) interface{} {
	switch typedVal := val.(type) {
	// necessary for overlay processing
	case []*DocumentSet:
		for i, item := range typedVal {
			typedVal[i] = convertToASTWithNoPosition(item).(*DocumentSet)
		}
		return typedVal

	case *DocumentSet:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToASTWithNoPosition(item).(*Document)
		}
		return typedVal

	case *Document:
		typedVal.Value = convertToASTWithNoPosition(typedVal.Value)
		return typedVal

	case *Map:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToASTWithNoPosition(item).(*MapItem)
		}
		return typedVal

	case *MapItem:
		typedVal.Key = convertToASTWithNoPosition(typedVal.Key)
		typedVal.Value = convertToASTWithNoPosition(typedVal.Value)
		return typedVal

	case *Array:
		for i, item := range typedVal.Items {
			typedVal.Items[i] = convertToASTWithNoPosition(item).(*ArrayItem)
		}
		return typedVal

	case *ArrayItem:
		typedVal.Value = convertToASTWithNoPosition(typedVal.Value)
		return typedVal

	case []interface{}:
		result := &Array{}
		for _, item := range typedVal {
			convertedValue := convertToASTWithNoPosition(item)
			result.Items = append(result.Items, &ArrayItem{
				Value:    convertedValue,
				Position: filepos.NewUnknownPositionWithKeyVal("-", convertedValue, ""),
			})
		}
		return result

	case map[interface{}]interface{}:
		panic("Expected *orderedmap.Map instead of map[interface{}]interface{} in convertToAST")

	case map[string]interface{}:
		panic("Expected *orderedmap.Map instead of map[string]interface{} in convertToAST")

	case *orderedmap.Map:
		result := &Map{}
		typedVal.Iterate(func(k, v interface{}) {
			convertedValue := convertToASTWithNoPosition(v)
			result.Items = append(result.Items, &MapItem{
				Key:      k,
				Value:    convertedValue,
				Position: filepos.NewUnknownPositionWithKeyVal(k, convertedValue, ":"),
			})
		})
		return result

	default:
		return val
	}
}
