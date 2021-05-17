// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

//type MapItem struct {
//	Type     Type
//	Metas    []*Meta
//	Key      interface{}
//	Value    interface{}
//	Position *filepos.Position
//
//	annotations interface{}
//}
func TestTypeAny(t *testing.T) {
	// #@schema/type any=True
	// key: val
	item := &yamlmeta.MapItem{Key: "key", Value: "val", Position: filepos.NewPosition(3)}
	nodeAnnotation := template.NodeAnnotation{Args: nil, Kwargs: []starlark.Tuple{{starlark.String("any"), starlark.True}}}
	nodeAnnotations := template.NodeAnnotations{}
	nodeAnnotations[AnnotationType] = nodeAnnotation
	item.SetAnnotations(nodeAnnotations)
	anns, err := ProcessAnnotations(item)
	require.NoError(t, err)
	assert.Contains(t, anns, &TypeAnnotation{any: true})
	assert.Len(t, anns, 1)

}

func TestNullable(t *testing.T) {
	// #@schema/nullable
	// key: val
	item := &yamlmeta.MapItem{Key: "key", Value: "val", Position: filepos.NewPosition(3)}
	nodeAnnotation := template.NodeAnnotation{}
	nodeAnnotations := template.NodeAnnotations{}
	nodeAnnotations[AnnotationNullable] = nodeAnnotation
	item.SetAnnotations(nodeAnnotations)
	anns, err := ProcessAnnotations(item)
	require.NoError(t, err)
	assert.Contains(t, anns, &NullableAnnotation{true, &ScalarType{Value: *new(string), Position: item.Position}})
	assert.Len(t, anns, 1)

}

func TestAnyTrueAndNullable(t *testing.T) {
	// #@schema/type any=True
	// #@schema/nullable
	// key: val
	item := &yamlmeta.MapItem{Key: "key", Value: "val", Position: filepos.NewPosition(3)}
	nodeAnnotation := template.NodeAnnotation{}
	nodeAnnotations := template.NodeAnnotations{}
	nodeAnnotations[AnnotationNullable] = nodeAnnotation
	nodeAnnotation = template.NodeAnnotation{Args: nil, Kwargs: []starlark.Tuple{{starlark.String("any"), starlark.True}}}
	nodeAnnotations[AnnotationType] = nodeAnnotation
	item.SetAnnotations(nodeAnnotations)
	anns, err := ProcessAnnotations(item)
	require.NoError(t, err)
	assert.Contains(t, anns, &TypeAnnotation{any: true})
	assert.Contains(t, anns, &NullableAnnotation{true, &ScalarType{Value: *new(string), Position: item.Position}})
	assert.Len(t, anns, 2)
}

func TestAnyFalseAndNullable(t *testing.T) {
	// #@schema/type any=False
	// #@schema/nullable
	// key: val
	item := &yamlmeta.MapItem{Key: "key", Value: "val", Position: filepos.NewPosition(3)}
	nodeAnnotation := template.NodeAnnotation{}
	nodeAnnotations := template.NodeAnnotations{}
	nodeAnnotations[AnnotationNullable] = nodeAnnotation
	nodeAnnotation = template.NodeAnnotation{Args: nil, Kwargs: []starlark.Tuple{{starlark.String("any"), starlark.False}}}
	nodeAnnotations[AnnotationType] = nodeAnnotation
	item.SetAnnotations(nodeAnnotations)
	anns, err := ProcessAnnotations(item)
	require.NoError(t, err)
	assert.Contains(t, anns, &TypeAnnotation{any: false})
	assert.Contains(t, anns, &NullableAnnotation{true, &ScalarType{Value: *new(string), Position: item.Position}})
	assert.Len(t, anns, 2)
}
