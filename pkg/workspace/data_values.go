// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

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
	AnnotationLibraryRef = "library/ref"

	dvsLibrarySep = "@"
)

type DataValues struct {
	Doc         *yamlmeta.Document
	AfterLibMod bool
	used        bool

	originalLibRef []ref.LibraryRef
	libRef         []ref.LibraryRef
}

func NewDataValues(doc *yamlmeta.Document) (*DataValues, error) {
	libRef, afterLibMod, err := parseDVAnnotations(ref.LibraryRefExtractor{}, doc)
	if err != nil {
		return nil, err
	}

	return &DataValues{Doc: doc, AfterLibMod: afterLibMod, libRef: libRef, originalLibRef: libRef}, nil
}

func NewEmptyDataValues() *DataValues {
	return &DataValues{Doc: newEmptyDataValuesDocument()}
}

func newEmptyDataValuesDocument() *yamlmeta.Document {
	return &yamlmeta.Document{
		Value:    &yamlmeta.Map{},
		Position: filepos.NewUnknownPosition(),
	}
}

type ExtractLibRefs interface {
	FromStr(string) ([]ref.LibraryRef, error)
	FromAnnotation(template.NodeAnnotations) ([]ref.LibraryRef, error)
}

func NewDataValuesWithLib(libRefs ExtractLibRefs, doc *yamlmeta.Document, libRefStr string) (*DataValues, error) {
	libRefsFromStr, err := libRefs.FromStr(libRefStr)
	if err != nil {
		return nil, err
	}

	libRefsFromAnnotation, afterLibMod, err := parseDVAnnotations(libRefs, doc)
	if err != nil {
		return nil, err
	} else if len(libRefsFromAnnotation) > 0 {
		panic(fmt.Sprintf("Library was provided as arg as well as with %s annotation", AnnotationLibraryRef))
	}

	return &DataValues{Doc: doc, AfterLibMod: afterLibMod, libRef: libRefsFromStr, originalLibRef: libRefsFromStr}, nil
}

func NewDataValuesWithOptionalLib(doc *yamlmeta.Document, libRefStr string) (*DataValues, error) {
	if len(libRefStr) > 0 {
		return NewDataValuesWithLib(ref.LibraryRefExtractor{}, doc, libRefStr)
	}
	return NewDataValues(doc)
}

func (dvd *DataValues) IsUsed() bool { return dvd.used }
func (dvd *DataValues) markUsed()    { dvd.used = true }

func (dvd *DataValues) Desc() string {
	// TODO: Update to output file location of annotation. If no annotation use doc position although these will always be used
	var desc []string
	for _, refPiece := range dvd.originalLibRef {
		desc = append(desc, refPiece.AsString())
	}
	return fmt.Sprintf("Data Value belonging to library '%s%s' on %s", dvsLibrarySep,
		strings.Join(desc, dvsLibrarySep), dvd.Doc.Position.AsString())
}

func (dvd *DataValues) IntendedForAnotherLibrary() bool { return len(dvd.libRef) > 0 }

func (dvd *DataValues) UsedInLibrary(expectedRefPiece ref.LibraryRef) *DataValues {
	if len(dvd.libRef) == 0 {
		dvd.markUsed()
		return dvd.deepCopy()
	}
	if !dvd.libRef[0].Matches(expectedRefPiece) {
		return nil
	}
	dvd.markUsed()
	childDV := dvd.deepCopy()
	childDV.libRef = childDV.libRef[1:]
	return childDV
}

func (dvd *DataValues) deepCopy() *DataValues {
	var copiedPieces []ref.LibraryRef
	copiedPieces = append(copiedPieces, dvd.libRef...)
	return &DataValues{Doc: dvd.Doc.DeepCopy(), AfterLibMod: dvd.AfterLibMod,
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
					AnnotationDataValues, AnnotationLibraryRef)
			}
		default:
			return nil, false, fmt.Errorf("Unknown kwarg %s for annotation %s", kwargName, AnnotationDataValues)
		}
	}
	return libRef, afterLibMod, nil
}
