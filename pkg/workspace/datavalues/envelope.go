// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package datavalues

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/workspace/ref"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	// AnnotationDataValues is the name of the annotation that marks a YAML document as a "Data Values" overlay.
	AnnotationDataValues template.AnnotationName = "data/values"

	// AnnotationDataValuesSchema is the name of the annotation that marks a YAML document as a "Data Values Schema"
	// overlay.
	AnnotationDataValuesSchema template.AnnotationName = "data/values-schema"
)

// Envelope wraps a YAML document containing Data Values along with addressing and usage bookkeeping — for which
// workspace.Library are these Data Values is intended.
type Envelope struct {
	Doc         *yamlmeta.Document
	AfterLibMod bool
	used        bool

	originalLibRef []ref.LibraryRef
	libRef         []ref.LibraryRef
}

// NewEnvelope generates a new Envelope from a YAML document containing Data Values, extracting "addressing" from any
// ref.LibraryRef annotated on the document.
func NewEnvelope(doc *yamlmeta.Document) (*Envelope, error) {
	libRef, afterLibMod, err := parseDVAnnotations(ref.LibraryRefExtractor{}, doc)
	if err != nil {
		return nil, err
	}

	return &Envelope{Doc: doc, AfterLibMod: afterLibMod, libRef: libRef, originalLibRef: libRef}, nil
}

// NewEmptyEnvelope constructs a new Envelope that contains no Data Values.
func NewEmptyEnvelope() *Envelope {
	return &Envelope{Doc: NewEmptyDataValuesDocument()}
}

// NewEmptyDataValuesDocument produces a YAML document that contains an empty map: the expected result when no data
// values are specified.
func NewEmptyDataValuesDocument() *yamlmeta.Document {
	return &yamlmeta.Document{
		Value:    &yamlmeta.Map{},
		Position: filepos.NewUnknownPosition(),
	}
}

// NewEnvelopeWithLibRef constructs a new Envelope using the given library reference for "addressing"
// If libRefStr is empty string, then the Envelope has no "addressing"
func NewEnvelopeWithLibRef(doc *yamlmeta.Document, libRefStr string) (*Envelope, error) {
	if len(libRefStr) > 0 {
		return newEnvelopeWithLibRef(ref.LibraryRefExtractor{}, doc, libRefStr)
	}
	return NewEnvelope(doc)
}

func newEnvelopeWithLibRef(libRefs ExtractLibRefs, doc *yamlmeta.Document, libRefStr string) (*Envelope, error) {
	libRefsFromStr, err := libRefs.FromStr(libRefStr)
	if err != nil {
		return nil, err
	}

	libRefsFromAnnotation, afterLibMod, err := parseDVAnnotations(libRefs, doc)
	if err != nil {
		return nil, err
	} else if len(libRefsFromAnnotation) > 0 {
		panic(fmt.Sprintf("Library was provided as arg as well as with %s annotation", ref.AnnotationLibraryRef))
	}

	return &Envelope{Doc: doc, AfterLibMod: afterLibMod, libRef: libRefsFromStr, originalLibRef: libRefsFromStr}, nil
}

// IsUsed reports whether this Envelope of Data Values has been "delivered"/used.
func (dvd *Envelope) IsUsed() bool { return dvd.used }
func (dvd *Envelope) markUsed()    { dvd.used = true }

// Desc reports the destination library of this Envelope
func (dvd *Envelope) Desc() string {
	// TODO: Update to output file location of annotation. If no annotation use doc position although these will always be used
	var desc []string
	for _, refPiece := range dvd.originalLibRef {
		desc = append(desc, refPiece.AsString())
	}
	return fmt.Sprintf("Data Value belonging to library '%s%s' on %s", ref.LibrarySep,
		strings.Join(desc, ref.LibrarySep), dvd.Doc.Position.AsString())
}

// IntendedForAnotherLibrary indicates whether this Envelope is "addressed" to some other library.
func (dvd *Envelope) IntendedForAnotherLibrary() bool { return len(dvd.libRef) > 0 }

// UsedInLibrary marks this Envelope as "delivered"/used if its destination is included in expectedRefPiece.
//
// If the Envelope should be used exactly in the specified library, returns a copy of this Envelope with no addressing
// (i.e. IntendedForAnotherLibrary() will indicate true).
// If the Envelope should be used in a _child_ of the specified library, returns a copy of this Envelope with
// the address of _that_ library.
// If the Envelope should **not** be used in the specified library or child, returns nil
func (dvd *Envelope) UsedInLibrary(expectedRefPiece ref.LibraryRef) *Envelope {
	if len(dvd.libRef) == 0 {
		dvd.markUsed()
		return dvd.deepCopyUnused()
	}
	if !dvd.libRef[0].Matches(expectedRefPiece) {
		return nil
	}
	dvd.markUsed()
	childDV := dvd.deepCopyUnused()
	childDV.libRef = childDV.libRef[1:]
	return childDV
}

func (dvd *Envelope) deepCopyUnused() *Envelope {
	var copiedPieces []ref.LibraryRef
	copiedPieces = append(copiedPieces, dvd.libRef...)
	return &Envelope{Doc: dvd.Doc.DeepCopy(), AfterLibMod: dvd.AfterLibMod,
		libRef: copiedPieces, originalLibRef: dvd.originalLibRef}
}

func parseDVAnnotations(libRefs ExtractLibRefs, doc *yamlmeta.Document) ([]ref.LibraryRef, bool, error) {
	var afterLibMod bool
	anns := template.NewAnnotations(doc)

	libRef, err := libRefs.FromAnnotation(anns)
	if err != nil {
		return nil, false, err
	}

	for _, kwarg := range anns.Kwargs(AnnotationDataValues) {
		kwargName, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return nil, false, err
		}

		switch kwargName {
		case "after_library_module":
			afterLibMod, err = core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return nil, false, err
			} else if len(libRef) == 0 {
				return nil, false, fmt.Errorf("Annotation %s: Expected kwarg 'after_library_module' to be used with %s annotation",
					AnnotationDataValues, ref.AnnotationLibraryRef)
			}
		default:
			return nil, false, fmt.Errorf("Unknown kwarg %s for annotation %s", kwargName, AnnotationDataValues)
		}
	}
	return libRef, afterLibMod, nil
}
