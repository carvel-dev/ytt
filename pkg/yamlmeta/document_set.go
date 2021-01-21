// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"bytes"
	"io"
)

type DocSetOpts struct {
	WithoutMeta bool
	Strict      bool
	// associatedName is typically a file name where data came from
	AssociatedName string
}

func NewDocumentSetFromBytes(data []byte, opts DocSetOpts) (*DocumentSet, error) {
	parserOpts := ParserOpts{WithoutMeta: opts.WithoutMeta, Strict: opts.Strict}

	docSet, err := NewParser(parserOpts).ParseBytes(data, opts.AssociatedName)
	if err != nil {
		return nil, err
	}
	docSet.originalBytes = &data
	return docSet, nil
}

func (ds *DocumentSet) Print(writer io.Writer) {
	NewPrinter(writer).Print(ds)
}

// AsSourceBytes() returns bytes used to make original DocumentSet.
// Any changes made to the DocumentSet are not reflected in any way
func (ds *DocumentSet) AsSourceBytes() ([]byte, bool) {
	if ds.originalBytes != nil {
		return *ds.originalBytes, true
	}
	return nil, false
}

func (ds *DocumentSet) AsBytes() ([]byte, error) {
	return ds.AsBytesWithPrinter(nil)
}

func (ds *DocumentSet) SetRawBytes(docSet *DocumentSet) {
	// TODO: DocumentSet.originalBytes already contain this information
	//       This sets RawData into the documents, but in the end it contains the full document set bits and not
	//       a document specific raw data. Eventually it might be ok but we should consider this as it might be a problem
	//       Some options that I see are:
	//         rename RawData to originalBytesDocumentSet or something.
	//         templating will split these originalBytes we have into per document string.
	//       I think the second option would be the one I would like but I think it might be a little bit of a big change
	//       because we will have to change the yaml parser to do this....
	for _, item := range docSet.Items {
		for _, doc := range ds.Items {
			if doc.Position != nil && doc.Position.AsCompactString() == item.Position.AsCompactString() {
				doc.RawData = item.RawData
				break
			}
		}
	}
}

func (ds *DocumentSet) AsBytesWithPrinter(printerFunc func(io.Writer) DocumentPrinter) ([]byte, error) {
	if printerFunc == nil {
		printerFunc = func(w io.Writer) DocumentPrinter { return NewYAMLPrinter(w) }
	}

	buf := new(bytes.Buffer)
	printer := printerFunc(buf)

	for _, item := range ds.Items {
		if item.injected || item.IsEmpty() {
			continue
		}
		printer.Print(item)
	}

	return buf.Bytes(), nil
}
