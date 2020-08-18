// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"fmt"
	"log"

	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

// An example showing how to unmarshal embedded
// structs from YAML.

type StructA struct {
	A string `yaml:"a"`
}

type StructB struct {
	// Embedded structs are not treated as embedded in YAML by default. To do that,
	// add the ",inline" annotation below
	StructA `yaml:",inline"`
	B       string `yaml:"b"`
}

var data = `
a: a string from struct A
b: a string from struct B
`

func ExampleUnmarshal_embedded() {
	var b StructB

	err := yaml.Unmarshal([]byte(data), &b)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	fmt.Println(b.A)
	fmt.Println(b.B)
	// Output:
	// a string from struct A
	// a string from struct B
}
