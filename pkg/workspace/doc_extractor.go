// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

type DocExtractor struct {
	DocSet *yamlmeta.DocumentSet
}

func (v DocExtractor) Extract(annName template.AnnotationName) ([]*yamlmeta.Document,
	[]*yamlmeta.Document, error) {

	err := v.checkNonDocs(v.DocSet, annName)
	if err != nil {
		return nil, nil, err
	}

	matchedDocs, nonMatchedDocs, err := v.extract(v.DocSet, annName)
	if err != nil {
		return nil, nil, err
	}

	return matchedDocs, nonMatchedDocs, nil
}

func (v DocExtractor) extract(docSet *yamlmeta.DocumentSet,
	annName template.AnnotationName) ([]*yamlmeta.Document, []*yamlmeta.Document, error) {

	var matchedDocs []*yamlmeta.Document
	var nonMatchedDocs []*yamlmeta.Document

	for _, doc := range docSet.Items {
		var hasMatchingAnn bool

		for _, comment := range doc.GetComments() {
			// TODO potentially use template.NewAnnotations(doc).Has(yttoverlay.AnnotationMatch)
			// however if doc was not processed by the template, it wont have any annotations set
			ann, err := yamltemplate.NewTemplateAnnotationFromYAMLComment(comment, doc.GetPosition(), yamltemplate.MetasOpts{IgnoreUnknown: true})
			if err != nil {
				return nil, nil, err
			}
			if ann.Name == annName {
				if hasMatchingAnn {
					return nil, nil, fmt.Errorf("%s annotation may only be used once per document", annName)
				}
				hasMatchingAnn = true
			}
		}

		if hasMatchingAnn {
			matchedDocs = append(matchedDocs, doc)
		} else {
			nonMatchedDocs = append(nonMatchedDocs, doc)
		}
	}

	return matchedDocs, nonMatchedDocs, nil
}

func (v DocExtractor) checkNonDocs(val interface{}, annName template.AnnotationName) error {
	node, ok := val.(yamlmeta.Node)
	if !ok {
		return nil
	}

	for _, comment := range node.GetComments() {
		ann, err := yamltemplate.NewTemplateAnnotationFromYAMLComment(comment, node.GetPosition(), yamltemplate.MetasOpts{IgnoreUnknown: true})
		if err != nil {
			return err
		}

		if ann.Name == annName {
			// TODO check for annotation emptiness
			_, isDoc := node.(*yamlmeta.Document)
			if !isDoc {
				errMsg := "Found @%s on %s (%s); only documents (---) can be annotated with @%s"
				return fmt.Errorf(errMsg, annName, yamlmeta.TypeName(node), node.GetPosition().AsCompactString(), annName)
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
