package yamlmeta

import (
	"bytes"
	"fmt"
	"io"

	"github.com/k14s/ytt/pkg/filepos"
)

type Printer struct {
	writer io.Writer
	opts   PrinterOpts
}

type PrinterOpts struct {
	ExcludeRefs bool
}

func NewPrinter(writer io.Writer) Printer {
	return Printer{writer, PrinterOpts{}}
}

func NewPrinterWithOpts(writer io.Writer, opts PrinterOpts) Printer {
	return Printer{writer, opts}
}

func (p Printer) Print(val interface{}) {
	fmt.Fprintf(p.writer, "%s", p.PrintStr(val))
}

func (p Printer) PrintStr(val interface{}) string {
	buf := new(bytes.Buffer)
	p.print(val, "", buf)
	return buf.String()
}

func (p Printer) print(val interface{}, indent string, writer io.Writer) {
	const indentLvl = "    "

	switch typedVal := val.(type) {
	case *DocumentSet:
		fmt.Fprintf(writer, "%s%s: docset%s\n", indent, p.lineStr(typedVal.Position), p.ptrStr(typedVal))
		p.printMeta(typedVal.Metas, indent, writer)

		for _, item := range typedVal.Items {
			p.print(item, indent+indentLvl, writer)
		}

	case *Document:
		fmt.Fprintf(writer, "%s%s: doc%s\n", indent, p.lineStr(typedVal.Position), p.ptrStr(typedVal))
		p.printMeta(typedVal.Metas, indent, writer)
		p.print(typedVal.Value, indent+indentLvl, writer)

	case *Map:
		fmt.Fprintf(writer, "%s%s: map%s\n", indent, p.lineStr(typedVal.Position), p.ptrStr(typedVal))
		p.printMeta(typedVal.Metas, indent, writer)

		for _, item := range typedVal.Items {
			fmt.Fprintf(writer, "%s%s: key=%s%s\n", indent, p.lineStr(item.Position), item.Key, p.ptrStr(item))
			p.printMeta(item.Metas, indent, writer)
			p.print(item.Value, indent+indentLvl, writer)
		}

	case *MapItem:
		fmt.Fprintf(writer, "%s%s: key=%s%s\n", indent, p.lineStr(typedVal.Position), typedVal.Key, p.ptrStr(typedVal))
		p.printMeta(typedVal.Metas, indent, writer)
		p.print(typedVal.Value, indent+indentLvl, writer)

	case *Array:
		fmt.Fprintf(writer, "%s%s: array%s\n", indent, p.lineStr(typedVal.Position), p.ptrStr(typedVal))
		p.printMeta(typedVal.Metas, indent, writer)

		for i, item := range typedVal.Items {
			fmt.Fprintf(writer, "%s%s: idx=%d%s\n", indent, p.lineStr(item.Position), i, p.ptrStr(item))
			p.printMeta(item.Metas, indent, writer)
			p.print(item.Value, indent+indentLvl, writer)
		}

	case *ArrayItem:
		fmt.Fprintf(writer, "%s%s: idx=top%s\n", indent, p.lineStr(typedVal.Position), p.ptrStr(typedVal))
		p.printMeta(typedVal.Metas, indent, writer)
		p.print(typedVal.Value, indent+indentLvl, writer)

	default:
		fmt.Fprintf(writer, "%s: %v\n", indent, typedVal)
	}
}

func (p Printer) lineStr(pos *filepos.Position) string {
	return pos.As4DigitString()
}

func (p Printer) ptrStr(node Node) string {
	if !p.opts.ExcludeRefs {
		return fmt.Sprintf(" (obj=%p)", node)
	}
	return ""
}

func (p Printer) printMeta(metas []*Meta, indent string, writer io.Writer) {
	for _, meta := range metas {
		fmt.Fprintf(writer, "%smeta: %s: '%s'\n", indent, p.lineStr(meta.Position), meta.Data)
	}
}
