package yttlibrary

import (
	"fmt"

	"github.com/k14s/ytt/pkg/structmeta"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

const (
	AnnotationDataValues structmeta.AnnotationName = "data/values"
)

type DataValues struct {
	DocSet    *yamlmeta.DocumentSet
	MetasOpts yamltemplate.MetasOpts
}

func (v DataValues) Find() (interface{}, bool, error) {
	doc, found, err := v.contains(v.DocSet)
	if !found || err != nil {
		return nil, found, err
	}

	return doc.AsInterface(yamlmeta.InterfaceConvertOpts{}), true, nil
}

func (v DataValues) contains(val interface{}) (*yamlmeta.Document, bool, error) {
	node, ok := val.(yamlmeta.Node)
	if !ok {
		return nil, false, nil
	}

	for _, meta := range node.GetMetas() {
		structMeta, err := yamltemplate.NewStructMetaFromMeta(meta, v.MetasOpts)
		if err != nil {
			return nil, false, err
		}

		for _, ann := range structMeta.Annotations {
			if ann.Name == AnnotationDataValues {
				// TODO check for ann emptiness
				doc, isDoc := node.(*yamlmeta.Document)
				if !isDoc {
					return nil, false, fmt.Errorf("Expected YAML doc to be annotated with %s but was %T",
						AnnotationDataValues, node)
				}
				return doc, true, nil
			}
		}
	}

	for _, childVal := range node.GetValues() {
		doc, found, err := v.contains(childVal)
		if found || err != nil {
			return doc, found, err
		}
	}

	return nil, false, nil
}
