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

func (d *DocumentSet) Print(writer io.Writer) {
	NewPrinter(writer).Print(d)
}

// AsSourceBytes() returns bytes used to make original DocumentSet.
// Any changes made to the DocumentSet are not reflected in any way
func (d *DocumentSet) AsSourceBytes() ([]byte, bool) {
	if d.originalBytes != nil {
		return *d.originalBytes, true
	}
	return nil, false
}

func (d *DocumentSet) AsBytes() ([]byte, error) {
	return d.AsBytesWithPrinter(nil)
}

func (d *DocumentSet) AsBytesWithPrinter(printerFunc func(io.Writer) DocumentPrinter) ([]byte, error) {
	if printerFunc == nil {
		printerFunc = func(w io.Writer) DocumentPrinter { return NewYAMLPrinter(w) }
	}

	buf := new(bytes.Buffer)
	printer := printerFunc(buf)

	for _, item := range d.Items {
		if item.injected || item.IsEmpty() {
			continue
		}
		printer.Print(item)
	}

	return buf.Bytes(), nil
}
