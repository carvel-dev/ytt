package yamlmeta

import (
	"fmt"
	"reflect"

	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

func PlainMarshal(val interface{}) ([]byte, error) {
	return yaml.Marshal(val)
}

func PlainUnmarshal(data []byte, val interface{}) error {
	docSet, err := NewParser().ParseBytes(data, "")
	if err != nil {
		return err
	}

	if len(docSet.Items) != 1 {
		return fmt.Errorf("Expected to find exactly one YAML document")
	}

	decodedVal := docSet.Items[0].AsInterface(InterfaceConvertOpts{})

	reflect.Indirect(reflect.ValueOf(val)).Set(reflect.ValueOf(decodedVal))

	return nil
}
