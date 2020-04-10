package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yttlibrary"
)

type DataValues struct {
	Doc         *yamlmeta.Document
	libPath     []LibPathPiece
	AfterLibMod bool
}

type LibPathPiece struct {
	LibName string
	Tag     string
}

const (
	AnnotationLibraryName = "library/name"
)

func NewDataValues(doc *yamlmeta.Document) (*DataValues, error) {
	_, libPath, afterLibMod, err := parseDVAnnotations(doc)
	if err != nil {
		return nil, err
	}

	return &DataValues{doc, libPath, afterLibMod}, nil
}

func NewEmptyDataValues() *DataValues {
	return &DataValues{&yamlmeta.Document{}, nil, false}
}

func NewDataValuesWithLib(doc *yamlmeta.Document, libStr string) (*DataValues, error) {
	libPath, err := parseLibStr(libStr)
	if err != nil {
		return nil, err
	}
	hasLibAnn, _, afterLibMod, err := parseDVAnnotations(doc)
	if err != nil {
		return nil, err
	} else if hasLibAnn {
		panic(fmt.Sprintf("Library was provided as arg as well as with %s annotation", AnnotationLibraryName))
	}
	return &DataValues{doc, libPath, afterLibMod}, nil
}

func (dvd *DataValues) HasLib() bool { return len(dvd.libPath) > 0 }

func (dvd *DataValues) PopLib() (string, string, bool, *DataValues) {
	if len(dvd.libPath) == 0 {
		panic("DataValue was not used by specified library")
	}

	updatedDV := dvd.deepCopy()
	name := updatedDV.libPath[0].LibName
	tag := updatedDV.libPath[0].Tag
	final := len(updatedDV.libPath) == 1
	updatedDV.libPath = dvd.libPath[1:]
	return name, tag, final, updatedDV
}

func (dvd *DataValues) deepCopy() *DataValues {
	var copiedPieces []LibPathPiece
	copiedPieces = append(copiedPieces, dvd.libPath...)
	return &DataValues{dvd.Doc.DeepCopy(), copiedPieces, dvd.AfterLibMod}
}

func parseDVAnnotations(doc *yamlmeta.Document) (bool, []LibPathPiece, bool, error) {
	var libPath []LibPathPiece
	var hasLibAnn, afterLibMod bool

	anns := template.NewAnnotations(doc)

	if hasLibAnn = anns.Has(AnnotationLibraryName); hasLibAnn {
		libArgs := anns.Args(AnnotationLibraryName)
		if l := libArgs.Len(); l != 1 {
			return false, nil, false, fmt.Errorf("Expected %s annotation to have one arg, got %d", yttlibrary.AnnotationDataValues, l)
		}

		argString, err := core.NewStarlarkValue(libArgs[0]).AsString()
		if err != nil {
			return false, nil, false, err
		}

		libPath, err = parseLibStr(argString)
		if err != nil {
			return false, nil, false, fmt.Errorf("Annotation %s: %s", AnnotationLibraryName, err.Error())
		}
	}

	for _, kwarg := range anns.Kwargs(yttlibrary.AnnotationDataValues) {
		kwargName, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return false, nil, false, err
		}

		switch kwargName {
		case "after_library_module":
			afterLibMod, err = core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return false, nil, false, err
			} else if len(libPath) == 0 {
				return false, nil, false, fmt.Errorf("Annotation %s: Expected kwarg 'after_library_module' to be used with %s annotation",
					yttlibrary.AnnotationDataValues, AnnotationLibraryName)
			}
		default:
			return false, nil, false, fmt.Errorf("Unknown kwarg %s for annotation %s", kwargName, yttlibrary.AnnotationDataValues)
		}
	}
	return hasLibAnn, libPath, afterLibMod, nil
}

func parseLibStr(libStr string) ([]LibPathPiece, error) {
	if libStr == "" {
		return nil, fmt.Errorf("Expected library name to not be empty")
	}

	if !strings.HasPrefix(libStr, "@") {
		return nil, fmt.Errorf("Expected library string to being with '@'")
	}

	var result []LibPathPiece
	for _, libPathPiece := range strings.Split(libStr, "@")[1:] {
		libAndTag := strings.SplitN(libPathPiece, "~", 2)
		piece := LibPathPiece{LibName: libAndTag[0]}
		if len(libAndTag) == 2 {
			if libAndTag[1] == "" {
				return nil, fmt.Errorf("Expected library tag to not be empty")
			}
			piece.Tag = libAndTag[1]
		}
		result = append(result, piece)
	}

	return result, nil
}
