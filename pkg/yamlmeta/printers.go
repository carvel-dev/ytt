// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/k14s/ytt/pkg/orderedmap"
)

type DocumentPrinter interface {
	Print(*Document) error
}

type YAMLPrinter struct {
	buf         io.Writer
	writtenOnce bool
}

var _ DocumentPrinter = &YAMLPrinter{}

func NewYAMLPrinter(writer io.Writer) *YAMLPrinter {
	return &YAMLPrinter{writer, false}
}

func (p *YAMLPrinter) Print(item *Document) error {
	if p.writtenOnce {
		p.buf.Write([]byte("---\n")) // TODO use encoder?
	} else {
		p.writtenOnce = true
	}

	bs, err := item.AsYAMLBytes()
	if err != nil {
		return fmt.Errorf("marshaling doc: %s", err)
	}
	p.buf.Write(bs)
	return nil
}

type JSONPrinter struct {
	buf io.Writer
}

var _ DocumentPrinter = &JSONPrinter{}

func NewJSONPrinter(writer io.Writer) JSONPrinter {
	return JSONPrinter{writer}
}

func (p JSONPrinter) Print(item *Document) error {
	val := item.AsInterface()

	bs, err := json.Marshal(orderedmap.Conversion{val}.AsUnorderedStringMaps())
	if err != nil {
		return fmt.Errorf("marshaling doc: %s", err)
	}
	p.buf.Write(bs)
	return nil
}

type WrappedFilePositionPrinter struct {
	Printer *FilePositionPrinter
}

var _ DocumentPrinter = WrappedFilePositionPrinter{}

func (p WrappedFilePositionPrinter) Print(item *Document) error {
	p.Printer.Print(item)
	return nil
}
