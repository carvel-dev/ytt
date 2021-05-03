// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationNs template.AnnotationNs = "overlay"

	AnnotationMerge   template.AnnotationName = "overlay/merge" // default
	AnnotationRemove  template.AnnotationName = "overlay/remove"
	AnnotationReplace template.AnnotationName = "overlay/replace"
	AnnotationInsert  template.AnnotationName = "overlay/insert" // array only
	AnnotationAppend  template.AnnotationName = "overlay/append" // array only
	AnnotationAssert  template.AnnotationName = "overlay/assert"

	AnnotationMatch              template.AnnotationName = "overlay/match"
	AnnotationMatchChildDefaults template.AnnotationName = "overlay/match-child-defaults"
)

var (
	allOps = []template.AnnotationName{
		AnnotationMerge,
		AnnotationRemove,
		AnnotationReplace,
		AnnotationInsert,
		AnnotationAppend,
		AnnotationAssert,
	}
)

func whichOp(node yamlmeta.Node) (template.AnnotationName, error) {
	var foundOp template.AnnotationName

	for _, op := range allOps {
		if template.NewAnnotations(node).Has(op) {
			if len(foundOp) > 0 {
				return "", fmt.Errorf("Expected to find only one overlay operation")
			}
			foundOp = op
		}
	}

	if len(foundOp) == 0 {
		foundOp = AnnotationMerge
	}

	return foundOp, nil
}
