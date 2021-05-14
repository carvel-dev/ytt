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
	item := &yamlmeta.MapItem{Key: "key", Value: "val", Position: filepos.NewPosition(3)}
	nodeAnnotation := template.NodeAnnotation{Args: nil, Kwargs: []starlark.Tuple{{starlark.String("any"),starlark.True}}}
	nodeAnnotations := template.NodeAnnotations{}
	nodeAnnotations[AnnotationType] = nodeAnnotation
	item.SetAnnotations(nodeAnnotations)
	typeAnn, err := ProcessAnnotations(item)
	require.NoError(t, err)
	assert.IsType(t, &TypeAnnotation{}, typeAnn)

}
func TestNullable(t *testing.T) {
	item := &yamlmeta.MapItem{Key: "key", Value: "val", Position: filepos.NewPosition(3)}
	nodeAnnotation := template.NodeAnnotation{}
	nodeAnnotations := template.NodeAnnotations{}
	nodeAnnotations[AnnotationNullable] = nodeAnnotation
	item.SetAnnotations(nodeAnnotations)
	nullableAnn, err := ProcessAnnotations(item)
	require.NoError(t, err)
	assert.IsType(t, &NullableAnnotation{}, nullableAnn)

}