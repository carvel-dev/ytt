// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"strings"

	"carvel.dev/ytt/pkg/filepos"
	"github.com/k14s/starlark-go/starlark"
)

type NodeAnnotations map[AnnotationName]NodeAnnotation

type NodeAnnotation struct {
	Args     starlark.Tuple
	Kwargs   []starlark.Tuple
	Position *filepos.Position
}

func NewAnnotations(node EvaluationNode) NodeAnnotations {
	result, ok := node.GetAnnotations().(NodeAnnotations)
	if !ok {
		result = NodeAnnotations{}
	}
	return result
}

func (as NodeAnnotations) DeepCopyAsInterface() interface{} {
	return as.DeepCopy()
}

func (as NodeAnnotations) DeepCopy() NodeAnnotations {
	result := NodeAnnotations{}
	for k, v := range as {
		result[k] = v // Dont need to copy v
	}
	return result
}

func (as NodeAnnotations) Has(name AnnotationName) bool {
	_, found := as[name]
	return found
}

func (as NodeAnnotations) Args(name AnnotationName) starlark.Tuple {
	na, found := as[name]
	if !found {
		return starlark.Tuple{}
	}
	return na.Args
}

func (as NodeAnnotations) Kwargs(name AnnotationName) []starlark.Tuple {
	na, found := as[name]
	if !found {
		return []starlark.Tuple{}
	}
	return na.Kwargs
}

func (as NodeAnnotations) DeleteNs(ns AnnotationNs) {
	prefix := string(ns) + "/"
	for k := range as {
		if strings.HasPrefix(string(k), prefix) {
			delete(as, k)
		}
	}
}
