// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

type FilePositionPrinter struct {
	writer   io.Writer
	opts     FilePositionPrinterOpts
	locWidth int
}

type FilePositionPrinterOpts struct{}

func NewFilePositionPrinter(writer io.Writer) *FilePositionPrinter {
	return &FilePositionPrinter{writer, FilePositionPrinterOpts{}, 0}
}

func NewFilePositionPrinterWithOpts(writer io.Writer, opts FilePositionPrinterOpts) *FilePositionPrinter {
	return &FilePositionPrinter{writer, opts, 0}
}

func (p *FilePositionPrinter) Print(val interface{}) {
	fmt.Fprintf(p.writer, "%s", p.PrintStr(val))
}

func (p *FilePositionPrinter) PrintStr(val interface{}) string {
	buf := new(bytes.Buffer)
	p.print(val, "", buf)
	return buf.String()
}

func (p *FilePositionPrinter) print(val interface{}, indent string, writer io.Writer) {
	const indentLvl = "  "

	switch typedVal := val.(type) {
	case *DocumentSet:
		fmt.Fprintf(writer, "%s%s[docset]\n", p.lineStr(typedVal.Position), indent)

		for _, item := range typedVal.Items {
			p.print(item, indent+indentLvl, writer)
		}

	case *Document:
		fmt.Fprintf(writer, "%s%s[doc]\n", p.lineStr(typedVal.Position), indent)
		p.print(typedVal.Value, indent+indentLvl, writer)

	case *Map:
		// fmt.Fprintf(writer, "%s%smap\n", indent, p.lineStr(typedVal.Position))

		for _, item := range typedVal.Items {
			valStr, isLeaf := p.leafValue(item.Value)
			if !isLeaf || strings.Contains(valStr, "\n") {
				fmt.Fprintf(writer, "%s%s%s:\n", p.lineStr(item.Position), indent, item.Key)
				p.print(item.Value, indent+indentLvl, writer)
			} else {
				fmt.Fprintf(writer, "%s%s%s: %s\n", p.lineStr(item.Position), indent, item.Key, valStr)
			}
		}

	case *MapItem:
		fmt.Fprintf(writer, "%s%s%s:\n", p.lineStr(typedVal.Position), indent, typedVal.Key)
		p.print(typedVal.Value, indent+indentLvl, writer)

	case *Array:
		// fmt.Fprintf(writer, "%s%sarray\n", indent, p.lineStr(typedVal.Position))

		for i, item := range typedVal.Items {
			valStr, isLeaf := p.leafValue(item.Value)
			if !isLeaf || strings.Contains(valStr, "\n") {
				fmt.Fprintf(writer, "%s%s[%d]\n", p.lineStr(item.Position), indent, i)
				p.print(item.Value, indent+indentLvl, writer)
			} else {
				fmt.Fprintf(writer, "%s%s[%d] %s\n", p.lineStr(item.Position), indent, i, valStr)
			}
		}

	case *ArrayItem:
		fmt.Fprintf(writer, "%s%s[?]\n", p.lineStr(typedVal.Position), indent)
		p.print(typedVal.Value, indent+indentLvl, writer)

	default:
		valStr, isLeaf := p.leafValue(val)
		if !isLeaf {
			panic(fmt.Sprintf("Expected leaf, but was %T", typedVal))
		}
		fmt.Fprintf(writer, p.padLine("")+fmt.Sprintf("%s%s\n", indent, valStr))
	}
}

func (p *FilePositionPrinter) leafValue(val interface{}) (string, bool) {
	switch typedVal := val.(type) {
	case *DocumentSet, *Document, *Map, *MapItem, *Array, *ArrayItem:
		return "", false

	default:
		typedValBs, err := yaml.Marshal(typedVal)
		if err != nil {
			panic(fmt.Sprintf("Failed to serialize %T", typedVal))
		}
		return string(typedValBs[:len(typedValBs)-1]), true // strip newline at the end
	}
}

func (p *FilePositionPrinter) lineStr(pos *filepos.Position) string {
	if pos.IsKnown() {
		return p.padLine(pos.AsCompactString())
	}
	return ""
}

func (p *FilePositionPrinter) padLine(str string) string {
	width := len(str)
	if width > p.locWidth {
		p.locWidth = width + 10
	}
	return fmt.Sprintf(fmt.Sprintf("%%%ds | ", p.locWidth), str)
}
