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

func (n *DocumentSet) Print(writer io.Writer) {
	NewPrinter(writer).Print(n)
}

// AsSourceBytes() returns bytes used to make original DocumentSet.
// Any changes made to the DocumentSet are not reflected in any way
func (n *DocumentSet) AsSourceBytes() ([]byte, bool) {
	if n.originalBytes != nil {
		return *n.originalBytes, true
	}
	return nil, false
}

func (n *DocumentSet) AsBytes() ([]byte, error) {
	return n.AsBytesWithPrinter(nil)
}

func (n *DocumentSet) AsBytesWithPrinter(printerFunc func(io.Writer) DocumentPrinter) ([]byte, error) {
	if printerFunc == nil {
		printerFunc = func(w io.Writer) DocumentPrinter { return NewYAMLPrinter(w) }
	}

	buf := new(bytes.Buffer)
	printer := printerFunc(buf)

	for _, item := range n.Items {
		if item.injected || item.IsEmpty() {
			continue
		}
		printer.Print(item)
	}

	return buf.Bytes(), nil
}
