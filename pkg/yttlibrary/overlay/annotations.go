package overlay

import (
	"fmt"

	"github.com/k14s/ytt/pkg/structmeta"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationNs structmeta.AnnotationNs = "overlay"

	AnnotationMerge   structmeta.AnnotationName = "overlay/merge" // default
	AnnotationRemove  structmeta.AnnotationName = "overlay/remove"
	AnnotationReplace structmeta.AnnotationName = "overlay/replace"
	AnnotationInsert  structmeta.AnnotationName = "overlay/insert" // array only
	AnnotationAppend  structmeta.AnnotationName = "overlay/append" // array only
	AnnotationAssert  structmeta.AnnotationName = "overlay/assert"

	AnnotationMatch              structmeta.AnnotationName = "overlay/match"
	AnnotationMatchChildDefaults structmeta.AnnotationName = "overlay/match-child-defaults"
)

var (
	allOps = []structmeta.AnnotationName{
		AnnotationMerge,
		AnnotationRemove,
		AnnotationReplace,
		AnnotationInsert,
		AnnotationAppend,
		AnnotationAssert,
	}
)

func whichOp(node yamlmeta.Node) (structmeta.AnnotationName, error) {
	var foundOp structmeta.AnnotationName

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
