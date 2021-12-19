// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yttlibrary/overlay"
)

type DataValuesFile struct {
	doc *yamlmeta.Document
}

func NewDataValuesFile(doc *yamlmeta.Document) DataValuesFile {
	return DataValuesFile{doc.DeepCopy()}
}

func (f DataValuesFile) AsOverlay() (*yamlmeta.Document, error) {
	doc := f.doc.DeepCopy()

	err := overlay.AnnotateForPlainMerge(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
