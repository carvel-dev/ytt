// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamlfmt

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"carvel.dev/ytt/pkg/yamlmeta"
)

// TODO strict mode
type Printer struct {
	writer   *writer
	opts     PrinterOpts
	locWidth int
}

type PrinterOpts struct{}

func NewPrinter(writer io.Writer) *Printer {
	return &Printer{newWriter(writer), PrinterOpts{}, 0}
}

func NewPrinterWithOpts(writer io.Writer, opts PrinterOpts) *Printer {
	return &Printer{newWriter(writer), opts, 0}
}

func (p *Printer) Print(val interface{}) {
	p.print(val, whitespace{}, p.writer)
}

func (p *Printer) PrintStr(val interface{}) string {
	buf := new(bytes.Buffer)
	p.print(val, whitespace{}, newWriter(buf))
	return buf.String()
}

func (p *Printer) print(val interface{}, ws whitespace, writer *writer) {
	switch typedVal := val.(type) {
	case *yamlmeta.DocumentSet:
		for i, item := range typedVal.Items {
			// TODO deal with first empty doc
			meta := p.printMeta(item, ws, writer, i == 0)
			if i != 0 || len(meta.Suffix) > 0 {
				writer.AddContent(writerChunk{
					Indent:  ws.Indent,
					Content: "---" + meta.Suffix,
				})
			}
			leafVal := p.leafValue(item.Value)
			if !leafVal.IsNil {
				p.print(item.Value, ws, writer)
			}
			if meta.AnyFull {
				writer.AddContent(writerChunk{Spacer: true})
			}
		}

	case *yamlmeta.Document:
		panic(fmt.Sprintf("Unexpected %T in Printer", val))

	case *yamlmeta.Map:
		for i, item := range typedVal.Items {
			meta := p.printMeta(item, ws, writer, i == 0)
			leafVal := p.leafValue(item.Value)
			if !leafVal.IsLeaf || leafVal.IsMultiline() {
				writer.AddContent(writerChunk{
					Indent:         ws.Indent,
					Content:        fmt.Sprintf("%s:%s", item.Key, meta.Suffix),
					AllowsInlining: len(meta.Suffix) == 0 && leafVal.IsLeaf,
					InliningSpacer: " ",
					CanBeInlined:   true,
				})
				p.print(item.Value, ws.NewIndented(), writer)
			} else {
				val := ""
				if !(len(meta.Suffix) > 0 && leafVal.IsNil) {
					val = " " + leafVal.String
				}
				writer.AddContent(writerChunk{
					Indent:       ws.Indent,
					Content:      fmt.Sprintf("%s:%s%s", item.Key, val, meta.Suffix),
					CanBeInlined: true,
				})
			}
			if meta.AnyFull {
				writer.AddContent(writerChunk{Spacer: true})
			}
		}

	case *yamlmeta.MapItem:
		panic(fmt.Sprintf("Unexpected %T in Printer", val))

	case *yamlmeta.Array:
		for i, item := range typedVal.Items {
			meta := p.printMeta(item, ws, writer, i == 0)
			leafVal := p.leafValue(item.Value)
			if !leafVal.IsLeaf || leafVal.IsMultiline() {
				itemWs := ws.NewIndented()

				writer.AddContent(writerChunk{
					Indent:         ws.Indent,
					Content:        fmt.Sprintf("-%s", meta.Suffix),
					AllowsInlining: len(meta.Suffix) == 0,
					InliningSpacer: " ",
					CanBeInlined:   true,
				})
				p.print(item.Value, itemWs, writer)
			} else {
				val := ""
				if !(len(meta.Suffix) > 0 && leafVal.IsNil) {
					val = " " + leafVal.String
				}
				writer.AddContent(writerChunk{
					Indent:       ws.Indent,
					Content:      fmt.Sprintf("-%s%s", val, meta.Suffix),
					CanBeInlined: true,
				})
			}
			if meta.AnyFull {
				writer.AddContent(writerChunk{Spacer: true})
			}
		}

	case *yamlmeta.ArrayItem:
		panic(fmt.Sprintf("Unexpected %T in Printer", val))

	default:
		leafVal := p.leafValue(val)
		if !leafVal.IsLeaf {
			panic(fmt.Sprintf("Expected leaf, but was %T", typedVal))
		}
		// TODO multiline string indent
		writer.AddContent(writerChunk{
			Indent:       ws.Indent,
			Content:      leafVal.String,
			CanBeInlined: true,
		})
	}
}

type printerLeafValue struct {
	String string
	IsLeaf bool
	IsNil  bool
}

func (v printerLeafValue) IsMultiline() bool {
	return strings.Contains(v.String, "\n")
}

func (p *Printer) leafValue(val interface{}) printerLeafValue {
	switch typedVal := val.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Document, *yamlmeta.Map, *yamlmeta.MapItem, *yamlmeta.Array, *yamlmeta.ArrayItem:
		return printerLeafValue{}

	default:
		typedValBs, err := yamlmeta.PlainMarshal(typedVal)
		if err != nil {
			panic(fmt.Sprintf("Failed to serialize %T", typedVal))
		}
		return printerLeafValue{
			String: string(typedValBs[:len(typedValBs)-1]), // strip newline at the end
			IsLeaf: true,
			IsNil:  typedVal == nil,
		}
	}
}

type printerMeta struct {
	Suffix  string
	AnyFull bool
}

func (p *Printer) printMeta(val interface{}, ws whitespace, writer *writer, firstChild bool) printerMeta {
	suffix := ""
	anyFull := false
	spaced := firstChild // do not add space before first child

	if typedVal, ok := val.(yamlmeta.Node); ok {
		for _, meta := range typedVal.GetComments() {
			if typedVal.GetPosition().IsKnown() && meta.Position.LineNum() == typedVal.GetPosition().LineNum() {
				suffix = fmt.Sprintf(" #%s", meta.Data)
			} else {
				if !spaced {
					spaced = true
					writer.AddContent(writerChunk{Spacer: true})
				}
				anyFull = true
				writer.AddContent(writerChunk{
					Indent:  ws.Indent,
					Content: "#" + meta.Data,
				})
			}
		}
	}
	return printerMeta{Suffix: suffix, AnyFull: anyFull}
}

type whitespace struct {
	Indent string
}

func (w whitespace) NewIndented() whitespace {
	const indentLvl = "  "
	return whitespace{Indent: w.Indent + indentLvl}
}
