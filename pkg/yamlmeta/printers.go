package yamlmeta

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
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

	bs, err := yaml.Marshal(item.AsInterface(InterfaceConvertOpts{OrderedMap: true}))
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
	bs, err := json.Marshal(item.AsInterface(InterfaceConvertOpts{OrderedMap: true, Plain: true}))
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
