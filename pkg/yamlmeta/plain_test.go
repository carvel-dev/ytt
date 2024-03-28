// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta_test

import (
	"fmt"
	"reflect"
	"testing"

	"carvel.dev/ytt/pkg/orderedmap"
	"carvel.dev/ytt/pkg/yamlmeta"
)

var _ = fmt.Sprintf

func TestPlainUnmarshalInt(t *testing.T) {
	var val interface{} = "abc" // set to some previous value

	err := yamlmeta.PlainUnmarshal([]byte("123"), &val)
	if err != nil {
		t.Fatalf("Expected to succeed: %s", err)
	}
	if !reflect.DeepEqual(val, 123) {
		t.Fatalf("Expected to be nil: val=%#v type=%T", val, val)
	}
}

func TestPlainUnmarshalNil(t *testing.T) {
	var val interface{} = 123 // set to some previous value

	err := yamlmeta.PlainUnmarshal([]byte("null"), &val)
	if err != nil {
		t.Fatalf("Expected to succeed: %s", err)
	}
	if !reflect.DeepEqual(val, nil) {
		t.Fatalf("Expected to be nil: val=%#v type=%T", val, val)
	}
}

func TestPlainUnmarshalMap(t *testing.T) {
	var val interface{} = 123 // set to some previous value

	err := yamlmeta.PlainUnmarshal([]byte(`{"a":123}`), &val)
	if err != nil {
		t.Fatalf("Expected to succeed: %s", err)
	}
	if !reflect.DeepEqual(val, orderedmap.NewMapWithItems([]orderedmap.MapItem{{Key: "a", Value: 123}})) {
		t.Fatalf("Expected to be nil: val=%#v type=%T", val, val)
	}
}

func TestPlainMarshalLineWidth(t *testing.T) {
	val := map[string]interface{}{
		"foo": "very long string very long string very long string very long string very long string very long string very long string very long string very long string very long string very long string",
	}

	bs, err := yamlmeta.PlainMarshal(val)
	if err != nil {
		t.Fatalf("Expected to succeed: %s", err)
	}

	expectedOutput := "foo: very long string very long string very long string very long string very long string very long string very long string very long string very long string very long string very long string\n"

	if string(bs) != expectedOutput {
		t.Fatalf("Expected to be single line but was '%s'", bs)
	}
}
