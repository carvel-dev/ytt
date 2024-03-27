// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"

	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/yamlmeta"
	"carvel.dev/ytt/pkg/yamltemplate"
	"github.com/k14s/starlark-go/starlark"
)

// AnnotateForPlainMerge configures `node` to be an overlay doing a "plain merge":
// - allow new keys via `@overlay/match missing_ok=True`
// - allow arrays and scalars to be replaced with given value (regardless of type) via `@overlay/replace or_add=True`
//
// Returns an error when `node` contains templating; `node` must be plain YAML.
func AnnotateForPlainMerge(node yamlmeta.Node) error {
	if yamltemplate.HasTemplating(node) {
		return fmt.Errorf("Expected to be plain YAML, having no annotations (hint: remove comments starting with `#@`)")
	}

	addOverlayReplace(node)
	return nil
}

func addOverlayReplace(node yamlmeta.Node) {
	anns := template.NodeAnnotations{}

	anns[AnnotationMatch] = template.NodeAnnotation{
		Kwargs: []starlark.Tuple{{
			starlark.String(MatchAnnotationKwargMissingOK),
			starlark.Bool(true),
		}},
	}

	replaceAnn := template.NodeAnnotation{
		Kwargs: []starlark.Tuple{{
			starlark.String(ReplaceAnnotationKwargOrAdd),
			starlark.Bool(true),
		}},
	}

	for _, val := range node.GetValues() {
		switch typedVal := val.(type) {
		case *yamlmeta.Array:
			anns[AnnotationReplace] = replaceAnn
		case yamlmeta.Node:
			addOverlayReplace(typedVal)
		default:
			anns[AnnotationReplace] = replaceAnn
		}
	}

	node.SetAnnotations(anns)
}
