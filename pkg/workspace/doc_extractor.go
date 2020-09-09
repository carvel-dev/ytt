// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"

	"github.com/k14s/ytt/pkg/structmeta"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

const (
	AnnotationDataValues  structmeta.AnnotationName = "data/values"
	AnnotationSchemaMatch structmeta.AnnotationName = "schema/match"
)

type DocExtractor struct {
	DocSet    *yamlmeta.DocumentSet
	MetasOpts yamltemplate.MetasOpts
}

func (v DocExtractor) Extract(annName structmeta.AnnotationName) ([]*yamlmeta.Document,
	[]*yamlmeta.Document, error) {

	err := v.checkNonDocs(v.DocSet, annName)
	if err != nil {
		return nil, nil, err
	}

	valuesDocs, nonValuesDocs, err := v.extract(v.DocSet, annName)
	if err != nil {
		return nil, nil, err
	}

	return valuesDocs, nonValuesDocs, nil
}

func (v DocExtractor) extract(docSet *yamlmeta.DocumentSet,
	annName structmeta.AnnotationName) ([]*yamlmeta.Document, []*yamlmeta.Document, error) {

	var valuesDocs []*yamlmeta.Document
	var nonValuesDocs []*yamlmeta.Document

	for _, doc := range docSet.Items {
		var hasMatchingAnn bool

		for _, meta := range doc.GetMetas() {
			// TODO potentially use template.NewAnnotations(doc).Has(yttoverlay.AnnotationMatch)
			// however if doc was not processed by the template, it wont have any annotations set
			structMeta, err := yamltemplate.NewStructMetaFromMeta(meta, v.MetasOpts)
			if err != nil {
				return nil, nil, err
			}
			for _, ann := range structMeta.Annotations {
				if ann.Name == annName {
					if hasMatchingAnn {
						return nil, nil, fmt.Errorf("%s annotation may only be used once per YAML doc", annName)
					}
					hasMatchingAnn = true
				}
			}
		}

		if hasMatchingAnn {
			valuesDocs = append(valuesDocs, doc)
		} else {
			nonValuesDocs = append(nonValuesDocs, doc)
		}
	}

	return valuesDocs, nonValuesDocs, nil
}

func (v DocExtractor) checkNonDocs(val interface{}, annName structmeta.AnnotationName) error {
	node, ok := val.(yamlmeta.Node)
	if !ok {
		return nil
	}

	for _, meta := range node.GetMetas() {
		structMeta, err := yamltemplate.NewStructMetaFromMeta(meta, v.MetasOpts)
		if err != nil {
			return err
		}

		for _, ann := range structMeta.Annotations {
			if ann.Name == annName {
				// TODO check for annotation emptiness
				_, isDoc := node.(*yamlmeta.Document)
				if !isDoc {
					errMsg := "Expected YAML document to be annotated with %s but was %T"
					return fmt.Errorf(errMsg, annName, node)
				}
			}
		}
	}

	for _, childVal := range node.GetValues() {
		err := v.checkNonDocs(childVal, annName)
		if err != nil {
			return err
		}
	}

	return nil
}
