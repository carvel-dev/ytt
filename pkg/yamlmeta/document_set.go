// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"bytes"
	"io"
)

type DocSetOpts struct {
	WithoutComments bool
	Strict          bool
	// associatedName is typically a file name where data came from
	AssociatedName string
}

func NewDocumentSetFromBytes(data []byte, opts DocSetOpts) (*DocumentSet, error) {
	parserOpts := ParserOpts{WithoutComments: opts.WithoutComments, Strict: opts.Strict}

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
