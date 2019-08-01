package yamlmeta

import (
	"fmt"
	"reflect"
)

func PlainUnmarshal(data []byte, out interface{}) error {
	docSet, err := NewParser(ParserOpts{WithoutMeta: true}).ParseBytes(data, "")
	if err != nil {
		return err
	}

	if len(docSet.Items) != 1 {
		return fmt.Errorf("Expected to find exactly one YAML document")
	}

	newVal := docSet.Items[0].AsInterface(InterfaceConvertOpts{})

	outVal := reflect.ValueOf(out)
	if newVal == nil {
		outVal.Elem().Set(reflect.Zero(outVal.Elem().Type()))
	} else {
		outVal.Elem().Set(reflect.ValueOf(newVal))
	}

	return nil
}
