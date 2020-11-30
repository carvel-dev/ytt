package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/structmeta"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationSchemaNullable structmeta.AnnotationName = "schema/nullable"
)

var _ yamlmeta.Type = (*DocumentType)(nil)
var _ yamlmeta.Type = (*MapType)(nil)
var _ yamlmeta.Type = (*MapItemType)(nil)
var _ yamlmeta.Type = (*ArrayType)(nil)
var _ yamlmeta.Type = (*ArrayItemType)(nil)

type DocumentType struct {
	Source    *yamlmeta.Document
	ValueType yamlmeta.Type // typically one of: MapType, ArrayType, ScalarType
}
type MapType struct {
	Items []*MapItemType
}
type MapItemType struct {
	Key          interface{} // usually a string
	ValueType    yamlmeta.Type
	DefaultValue interface{}
	Position     *filepos.Position
	Annotations  TypeAnnotations
}
type ArrayType struct {
	ItemsType yamlmeta.Type
}
type ArrayItemType struct {
	ValueType yamlmeta.Type
}
type ScalarType struct {
	Type interface{}
}

type TypeAnnotations map[structmeta.AnnotationName]interface{}

func (t *DocumentType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (m MapType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (m MapItemType) GetValueType() yamlmeta.Type {
	return m.ValueType
}
func (a ArrayType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (a ArrayItemType) GetValueType() yamlmeta.Type {
	return a.ValueType
}
func (m ScalarType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}

func (t *DocumentType) CheckType(_ yamlmeta.TypeWithValues, _ string) (chk yamlmeta.TypeCheck) {
	return
}
func (m *MapType) CheckType(node yamlmeta.TypeWithValues, prependErrorMessage string) (chk yamlmeta.TypeCheck) {
	violationErrorMessage := prependErrorMessage + " was type %T when %T was expected"

	_, ok := node.(*yamlmeta.Map)
	if !ok {
		chk.Violations = append(chk.Violations, fmt.Sprintf(violationErrorMessage, node.GetValues()[0], &yamlmeta.Map{}))
		return
	}
	return
}
func (m MapItemType) CheckType(node yamlmeta.TypeWithValues, prependErrorMessage string) (chk yamlmeta.TypeCheck) {
	violationErrorMessage := prependErrorMessage + " was type %T when %T was expected"

	mapItem, ok := node.(*yamlmeta.MapItem)
	if !ok {
		chk.Violations = append(chk.Violations, fmt.Sprintf(violationErrorMessage, node.GetValues()[0], node))
		return
	}
	if mapItem.Value == nil && !m.IsNullable() {
		chk.Violations = append(chk.Violations, fmt.Sprintf(violationErrorMessage, node.GetValues()[0], m.ValueType))
	}

	return
}
func (a ArrayType) CheckType(node yamlmeta.TypeWithValues, prependErrorMessage string) (chk yamlmeta.TypeCheck) {
	violationErrorMessage := prependErrorMessage + " was type %T when %T was expected"

	_, ok := node.(*yamlmeta.Array)
	if !ok {
		chk.Violations = append(chk.Violations, fmt.Sprintf(violationErrorMessage, node.GetValues()[0], &yamlmeta.Array{}))
		return
	}
	return
}
func (a ArrayItemType) CheckType(node yamlmeta.TypeWithValues, prependErrorMessage string) (chk yamlmeta.TypeCheck) {
	violationErrorMessage := prependErrorMessage + " was type %T when %T was expected"

	_, ok := node.(*yamlmeta.ArrayItem)
	if !ok {
		chk.Violations = append(chk.Violations, fmt.Sprintf(violationErrorMessage, node.GetValues()[0], node))
		return
	}
	return
}
func (m ScalarType) CheckType(node yamlmeta.TypeWithValues, prependErrorMessage string) (chk yamlmeta.TypeCheck) {
	violationErrorMessage := prependErrorMessage + " was type %T when %T was expected"

	value := node.GetValues()[0]
	switch itemValueType := value.(type) {
	case string:
		if _, ok := m.Type.(string); !ok {
			violation := fmt.Sprintf(violationErrorMessage, itemValueType, m.Type)
			chk.Violations = append(chk.Violations, violation)
		}
	case int:
		if _, ok := m.Type.(int); !ok {
			violation := fmt.Sprintf(violationErrorMessage, itemValueType, m.Type)
			chk.Violations = append(chk.Violations, violation)
		}
	case bool:
		if _, ok := m.Type.(bool); !ok {
			violation := fmt.Sprintf(violationErrorMessage, itemValueType, m.Type)
			chk.Violations = append(chk.Violations, violation)
		}
	default:
		violation := fmt.Sprintf(violationErrorMessage, itemValueType, m.Type)
		chk.Violations = append(chk.Violations, violation)
	}
	return
}

func (t *DocumentType) CheckAllows(item *yamlmeta.MapItem) yamlmeta.TypeCheck {
	panic("Attempt to check if a MapItem is allowed as a value of a Document.")
}
func (m MapItemType) CheckAllows(item *yamlmeta.MapItem) yamlmeta.TypeCheck {
	panic("Attempt to check if a MapItem is allowed as a value of a MapItem.")
}
func (a ArrayType) CheckAllows(item *yamlmeta.MapItem) yamlmeta.TypeCheck {
	panic("Attempt to check if a MapItem is allowed as a value of an Array.")
}
func (a ArrayItemType) CheckAllows(item *yamlmeta.MapItem) yamlmeta.TypeCheck {
	panic("Attempt to check if a MapItem is allowed as a value of an ArrayItemType.")
}
func (m ScalarType) CheckAllows(item *yamlmeta.MapItem) yamlmeta.TypeCheck {
	panic("Attempt to check if a MapItem is allowed as a value of a ScalarType.")
}

func (t *DocumentType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	doc, ok := typeable.(*yamlmeta.Document)
	if !ok {
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &yamlmeta.Document{}, typeable)}
		return
	}
	typeable.SetType(t)
	typeableChild, ok := doc.Value.(yamlmeta.Typeable)
	if ok || doc.Value == nil {
		if t.ValueType != nil {
			tChild := typeableChild
			if doc.Value == nil {
				switch t.ValueType.(type) {
				case *MapType:
					tChild = &yamlmeta.Map{}
				default:
					chk.Violations = append(chk.Violations, fmt.Sprintf("Expected node at %s to be %s, but was a %T", typeableChild.GetPosition().AsCompactString(), "Map", t.ValueType))
				}
				doc.Value = tChild
			}
			childCheck := t.ValueType.AssignTypeTo(tChild)
			chk.Violations = append(chk.Violations, childCheck.Violations...)
		} else {
			chk.Violations = []string{fmt.Sprintf("Expected node at %s to be %s, but was a %T", typeableChild.GetPosition().AsCompactString(), "nil", typeableChild)}
		}
	} else {

	} // else, at a leaf
	return
}
func (t *MapType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	mapNode, ok := typeable.(*yamlmeta.Map)
	if !ok {
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &yamlmeta.Map{}, typeable)}
		return
	}
	var foundKeys []interface{}
	typeable.SetType(t)
	for _, mapItem := range mapNode.Items {
		for _, itemType := range t.Items {
			if mapItem.Key == itemType.Key {
				foundKeys = append(foundKeys, itemType.Key)
				childCheck := itemType.AssignTypeTo(mapItem)
				chk.Violations = append(chk.Violations, childCheck.Violations...)
				break
			}
		}
	}

	t.applySchemaDefaults(foundKeys, chk, mapNode)
	return
}

func (t *MapType) applySchemaDefaults(foundKeys []interface{}, chk yamlmeta.TypeCheck, mapNode *yamlmeta.Map) {
	for _, item := range t.Items {
		if contains(foundKeys, item.Key) {
			continue
		}

		val := &yamlmeta.MapItem{
			Key:      item.Key,
			Value:    item.DefaultValue,
			Position: item.Position,
		}
		childCheck := item.AssignTypeTo(val)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
		err := mapNode.AddValue(val)
		if err != nil {
			panic(fmt.Sprintf("Internal inconsistency: adding map item: %s", err))
		}
	}
}

func contains(haystack []interface{}, needle interface{}) bool {
	for _, key := range haystack {
		if key == needle {
			return true
		}
	}
	return false
}

func (t *MapItemType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	mapItem, ok := typeable.(*yamlmeta.MapItem)
	if !ok {
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &yamlmeta.MapItem{}, typeable)}
		return
	}
	typeable.SetType(t)
	typeableValue, ok := mapItem.Value.(yamlmeta.Typeable)
	if ok {
		childCheck := t.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, at a leaf
	return
}
func (t *ArrayType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	arrayNode, ok := typeable.(*yamlmeta.Array)
	if !ok {
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &yamlmeta.Array{}, typeable)}
		return
	}
	typeable.SetType(t)
	for _, arrayItem := range arrayNode.Items {
		childCheck := t.ItemsType.AssignTypeTo(arrayItem)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	}
	return
}
func (t ArrayItemType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	arrayItem, ok := typeable.(*yamlmeta.ArrayItem)
	if !ok {
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &yamlmeta.ArrayItem{}, typeable)}
		return
	}
	typeable.SetType(t)
	typeableValue, ok := arrayItem.Value.(yamlmeta.Typeable)
	if ok {
		childCheck := t.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, at a leaf
	return
}

func (t *ScalarType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	switch t.Type.(type) {
	case int:
		typeable.SetType(t)
	case string:
		typeable.SetType(t)
	default:
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &ScalarType{}, typeable)}
	}
	return
}

func (t *MapType) AllowsKey(key interface{}) bool {
	for _, item := range t.Items {
		if item.Key == key {
			return true
		}
	}
	return false
}

func (t *MapType) CheckAllows(item *yamlmeta.MapItem) (chk yamlmeta.TypeCheck) {
	if !t.AllowsKey(item.Key) {
		chk.Violations = append(chk.Violations, fmt.Sprintf("Map item '%s' at %s is not defined in schema", item.Key, item.Position.AsCompactString()))
	}
	return chk
}

func (t MapItemType) IsNullable() bool {
	_, found := t.Annotations[AnnotationSchemaNullable]
	return found
}
