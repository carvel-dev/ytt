// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamltemplate"
)

// CommentPostProcessing is an instance of the process of filling in templated comments into the configured document
// set.
type CommentPostProcessing struct {
	docSets map[*FileInLibrary]*yamlmeta.DocumentSet
}

// Visit converts annotated comments into YAML comments
func (o CommentPostProcessing) Visit(node yamlmeta.Node) error {
	comments := []*yamlmeta.Comment{}
	anns := template.NewAnnotations(node)
	if anns.Has(yamltemplate.AnnotationHeadComment) {
		comments = append(comments, &yamlmeta.Comment{
			Data:     anns.Args(yamltemplate.AnnotationHeadComment)[0].String(),
			Position: filepos.NewUnknownPosition(),
		})
	}
	if anns.Has(yamltemplate.AnnotationLineComment) {
		comments = append(comments, &yamlmeta.Comment{
			Data:     anns.Args(yamltemplate.AnnotationLineComment)[0].String(),
			Position: node.GetPosition(),
		})
	}
	node.SetComments(comments)
	return nil
}

// Apply performs comment post-processing on the configured document set.
func (o CommentPostProcessing) Apply() (map[*FileInLibrary]*yamlmeta.DocumentSet, error) {
	for _, docSet := range o.docSets {
		err := yamlmeta.Walk(docSet, o)
		if err != nil {
			return nil, err
		}
	}
	return o.docSets, nil
}
