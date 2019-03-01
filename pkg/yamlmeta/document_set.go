package yamlmeta

import (
	"bytes"
	"fmt"
	"io"

	"github.com/get-ytt/ytt/pkg/yamlmeta/internal/yaml.v2"
)

// associatedName is typicall a file name where data came from
func NewDocumentSetFromBytes(data []byte, associatedName string) (*DocumentSet, error) {
	docSet, err := NewParser().ParseBytes(data, associatedName)
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
	buf := new(bytes.Buffer)
	var writtenOnce bool

	for i, item := range d.Items {
		if item.injected || item.IsEmpty() {
			continue
		}

		if writtenOnce {
			buf.Write([]byte("---\n")) // TODO use encoder?
		} else {
			writtenOnce = true
		}

		bs, err := yaml.Marshal(item.AsInterface(InterfaceConvertOpts{OrderedMap: true}))
		if err != nil {
			return nil, fmt.Errorf("marshaling doc %d: %s", i, err)
		}

		buf.Write(bs)
	}

	return buf.Bytes(), nil
}
