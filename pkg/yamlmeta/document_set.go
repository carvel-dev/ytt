package yamlmeta

import (
	"bytes"
	"fmt"
	"io"

	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

type DocSetOpts struct {
	WithoutMeta bool
	// associatedName is typically a file name where data came from
	AssociatedName string
}

func NewDocumentSetFromBytesWithOpts(data []byte, opts DocSetOpts) (*DocumentSet, error) {
	docSet, err := NewParser(opts.WithoutMeta).ParseBytes(data, opts.AssociatedName)
	if err != nil {
		return nil, err
	}
	docSet.originalBytes = &data
	return docSet, nil
}

func NewDocumentSetFromBytes(data []byte, associatedName string) (*DocumentSet, error) {
	return NewDocumentSetFromBytesWithOpts(data, DocSetOpts{AssociatedName: associatedName})
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
	buf := new(bytes.Buffer)

	for i, item := range d.Items {
		if item.injected || item.IsEmpty() {
			continue
		}

		buf.Write([]byte("---\n"))

		bs, err := yaml.Marshal(item.AsInterface(InterfaceConvertOpts{OrderedMap: true}))
		if err != nil {
			return nil, fmt.Errorf("marshaling doc %d: %s", i, err)
		}

		buf.Write(bs)
	}

	return buf.Bytes(), nil
}
