// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"os"
	"strings"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/files"
	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/workspace/datavalues"
	"carvel.dev/ytt/pkg/workspace/ref"
	"carvel.dev/ytt/pkg/yamlmeta"
	yttoverlay "carvel.dev/ytt/pkg/yttlibrary/overlay"
	"github.com/k14s/starlark-go/starlark"
)

const (
	dvsKVSep      = "="
	dvsMapKeySep  = "."
	libraryKeySep = ":"
)

type DataValuesFlags struct {
	EnvFromStrings []string
	EnvFromYAML    []string

	KVsFromStrings []string
	KVsFromYAML    []string
	KVsFromFiles   []string

	FromFiles []string

	Inspect        bool
	InspectSchema  bool
	SkipValidation bool

	EnvironFunc   func() []string
	ReadFilesFunc func(paths string) ([]*files.File, error)

	*files.SymlinkAllowOpts
}

// Set registers data values ingestion flags and wires-up those flags up to this
// DataValuesFlags to be set when the corresponding cobra.Command is executed.
func (s *DataValuesFlags) Set(cmdFlags CmdFlags) {
	cmdFlags.StringArrayVar(&s.EnvFromStrings, "data-values-env", nil, "Extract data values (as strings) from prefixed env vars (format: PREFIX for PREFIX_all__key1=str) (can be specified multiple times)")
	cmdFlags.StringArrayVar(&s.EnvFromYAML, "data-values-env-yaml", nil, "Extract data values (parsed as YAML) from prefixed env vars (format: PREFIX for PREFIX_all__key1=true) (can be specified multiple times)")

	cmdFlags.StringArrayVarP(&s.KVsFromStrings, "data-value", "v", nil, "Set specific data value to given value, as string (format: all.key1.subkey=123) (can be specified multiple times)")
	cmdFlags.StringArrayVar(&s.KVsFromYAML, "data-value-yaml", nil, "Set specific data value to given value, parsed as YAML (format: all.key1.subkey=true) (can be specified multiple times)")
	cmdFlags.StringArrayVar(&s.KVsFromFiles, "data-value-file", nil, "Set specific data value to contents of a file (format: [@lib1:]all.key1.subkey={file path, HTTP URL, or '-' (i.e. stdin)}) (can be specified multiple times)")
	cmdFlags.StringArrayVar(&s.FromFiles, "data-values-file", nil, "Set multiple data values via plain YAML files (format: [@lib1:]{file path, HTTP URL, or '-' (i.e. stdin)}) (can be specified multiple times)")

	cmdFlags.BoolVar(&s.Inspect, "data-values-inspect", false, "Determine the final data values (applying any overlays) and display that result")
	cmdFlags.BoolVar(&s.SkipValidation, "dangerous-data-values-disable-validation", false, "Skip validating data values (not recommended: may result in templates failing or invalid output)")
	cmdFlags.BoolVar(&s.InspectSchema, "data-values-schema-inspect", false, "Determine the complete schema for data values (applying any overlays) and display the result (only OpenAPI v3.0 is supported, see --output)")
}

type dataValuesFlagsSource struct {
	Values        []string
	TransformFunc valueTransformFunc
	Name          string
}

type valueTransformFunc func(string) (interface{}, error)

// AsOverlays generates Data Values overlays, one for each setting in this DataValuesFlags.
//
// Returns a collection of overlays targeted for the root library and a separate collection of overlays "addressed" to
// children libraries.
func (s *DataValuesFlags) AsOverlays(strict bool) ([]*datavalues.Envelope, []*datavalues.Envelope, error) {
	plainValFunc := func(rawVal string) (interface{}, error) { return rawVal, nil }

	yamlValFunc := func(rawVal string) (interface{}, error) {
		val, err := s.parseYAML(rawVal, strict)
		if err != nil {
			return nil, fmt.Errorf("Deserializing YAML value: %s", err)
		}
		return val, nil
	}

	var result []*datavalues.Envelope

	// Files go first
	for _, file := range s.FromFiles {
		vals, err := s.file(file, strict)
		if err != nil {
			return nil, nil, fmt.Errorf("Extracting data value from file: %s", err)
		}
		result = append(result, vals...)
	}

	// Then env vars take precedence over files
	// since env vars are specific to command execution
	for _, src := range []dataValuesFlagsSource{{s.EnvFromStrings, plainValFunc, "data-values-env"}, {s.EnvFromYAML, yamlValFunc, "data-values-env-yaml"}} {
		for _, envPrefix := range src.Values {
			vals, err := s.env(envPrefix, src)
			if err != nil {
				return nil, nil, fmt.Errorf("Extracting data values from env under prefix '%s': %s", envPrefix, err)
			}
			result = append(result, vals...)
		}
	}

	// KVs take precedence over environment variables
	for _, src := range []dataValuesFlagsSource{{s.KVsFromStrings, plainValFunc, "data-value"}, {s.KVsFromYAML, yamlValFunc, "data-value-yaml"}} {
		for _, kv := range src.Values {
			val, err := s.kv(kv, src)
			if err != nil {
				return nil, nil, fmt.Errorf("Extracting data value from KV: %s", err)
			}
			result = append(result, val)
		}
	}

	// Finally KV files take precedence over rest
	// (technically should be same level as KVs, but gotta pick one)
	for _, file := range s.KVsFromFiles {
		val, err := s.kvFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("Extracting data value from file: %s", err)
		}
		result = append(result, val)
	}

	var overlayValues []*datavalues.Envelope
	var libraryOverlays []*datavalues.Envelope
	for _, doc := range result {
		if doc.IntendedForAnotherLibrary() {
			libraryOverlays = append(libraryOverlays, doc)
		} else {
			overlayValues = append(overlayValues, doc)
		}
	}

	return overlayValues, libraryOverlays, nil
}

func (s *DataValuesFlags) file(fullPath string, strict bool) ([]*datavalues.Envelope, error) {
	libRef, path, err := s.libraryRefAndRemainder(fullPath)
	if err != nil {
		return nil, err
	}

	dvFiles, err := s.asFiles(path)
	if err != nil {
		return nil, fmt.Errorf("Find files '%s': %s", path, err)
	}

	var result []*datavalues.Envelope
	for _, dvFile := range dvFiles {
		// Users may want to store other files (docs, etc.) within this directory; ignore those.
		if dvFile.IsImplied() && !(dvFile.Type() == files.TypeYAML) {
			continue
		}
		contents, err := dvFile.Bytes()
		if err != nil {
			return nil, fmt.Errorf("Reading file '%s': %s", dvFile.RelativePath(), err)
		}

		docSetOpts := yamlmeta.DocSetOpts{
			AssociatedName: dvFile.RelativePath(),
			Strict:         strict,
		}
		docSet, err := yamlmeta.NewDocumentSetFromBytes(contents, docSetOpts)
		if err != nil {
			return nil, fmt.Errorf("Unmarshaling YAML data values file '%s': %s", dvFile.RelativePath(), err)
		}

		for _, doc := range docSet.Items {
			if doc.Value != nil {
				dvsOverlay, err := NewDataValuesFile(doc).AsOverlay()
				if err != nil {
					return nil, fmt.Errorf("Checking data values file '%s': %s", path, err)
				}
				dvs, err := datavalues.NewEnvelopeWithLibRef(dvsOverlay, libRef)
				if err != nil {
					return nil, err
				}
				result = append(result, dvs)
			}
		}
	}

	return result, nil
}

func (s *DataValuesFlags) env(prefix string, src dataValuesFlagsSource) ([]*datavalues.Envelope, error) {
	const (
		envKeyPrefix = "_"
		envMapKeySep = "__"
	)

	result := []*datavalues.Envelope{}
	envVars := os.Environ()

	if s.EnvironFunc != nil {
		envVars = s.EnvironFunc()
	}

	libRef, keyPrefix, err := s.libraryRefAndKey(prefix)
	if err != nil {
		return nil, err
	}

	for _, envVar := range envVars {
		pieces := strings.SplitN(envVar, dvsKVSep, 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("Expected env variable to be key-value pair (format: key=value)")
		}

		if !strings.HasPrefix(pieces[0], keyPrefix+envKeyPrefix) {
			continue
		}

		val, err := src.TransformFunc(pieces[1])
		if err != nil {
			return nil, fmt.Errorf("Extracting data value from env variable '%s': %s", pieces[0], err)
		}

		// '__' gets translated into a '.' since periods may not be liked by shells
		keyPieces := strings.Split(strings.TrimPrefix(pieces[0], keyPrefix+envKeyPrefix), envMapKeySep)
		desc := fmt.Sprintf("(%s arg) %s", src.Name, keyPrefix)
		overlay := s.buildOverlay(keyPieces, val, desc, envVar)

		dvs, err := datavalues.NewEnvelopeWithLibRef(overlay, libRef)
		if err != nil {
			return nil, err
		}

		result = append(result, dvs)
	}

	return result, nil
}

func (s *DataValuesFlags) kv(kv string, src dataValuesFlagsSource) (*datavalues.Envelope, error) {
	pieces := strings.SplitN(kv, dvsKVSep, 2)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("Expected format key=value")
	}

	val, err := src.TransformFunc(pieces[1])
	if err != nil {
		return nil, fmt.Errorf("Deserializing value for key '%s': %s", pieces[0], err)
	}

	libRef, key, err := s.libraryRefAndKey(pieces[0])
	if err != nil {
		return nil, err
	}
	desc := fmt.Sprintf("(%s arg)", src.Name)
	overlay := s.buildOverlay(strings.Split(key, dvsMapKeySep), val, desc, kv)

	return datavalues.NewEnvelopeWithLibRef(overlay, libRef)
}

func (s *DataValuesFlags) parseYAML(data string, strict bool) (interface{}, error) {
	docSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{Strict: strict}).ParseBytes([]byte(data), "")
	if err != nil {
		return nil, err
	}
	return docSet.Items[0].Value, nil
}

func (s *DataValuesFlags) kvFile(kv string) (*datavalues.Envelope, error) {
	pieces := strings.SplitN(kv, dvsKVSep, 2)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("Expected format key=/file/path")
	}

	dvFile, err := s.asFiles(pieces[1])
	if err != nil {
		return nil, fmt.Errorf("Finding file '%s': %s", pieces[1], err)
	}
	if len(dvFile) > 1 {
		return nil, fmt.Errorf("Expected '%s' to be a file, but is a directory", pieces[1])
	}

	contents, err := dvFile[0].Bytes()
	if err != nil {
		return nil, fmt.Errorf("Reading file '%s'", pieces[1])
	}

	libRef, key, err := s.libraryRefAndKey(pieces[0])
	if err != nil {
		return nil, err
	}
	desc := fmt.Sprintf("(data-value-file arg) %s=%s", key, pieces[1])
	overlay := s.buildOverlay(strings.Split(key, dvsMapKeySep), string(contents), desc, string(contents))

	return datavalues.NewEnvelopeWithLibRef(overlay, libRef)
}

// libraryRefAndKey separates a library reference and a key and validates that no libraryKeySep exist in the key.
// libraryKeySep is disallowed in data value flag keys.
func (DataValuesFlags) libraryRefAndKey(arg string) (string, string, error) {
	libRef, key, err := DataValuesFlags{}.libraryRefAndRemainder(arg)
	if err != nil {
		return "", "", err
	}
	if len(strings.Split(key, libraryKeySep)) > 1 {
		// error on a common syntax mistake
		return "", "", fmt.Errorf("Expected at most one library-key separator '%s' in '%s'", libraryKeySep, key)
	}
	return libRef, key, nil
}

// libraryRefAndRemainder separates a library reference prefix from the remainder of the string.
// A library reference starts with ref.LibrarySep, and ends with the first occurrence of libraryKeySep.
func (DataValuesFlags) libraryRefAndRemainder(arg string) (string, string, error) {
	if strings.HasPrefix(arg, ref.LibrarySep) {
		strPieces := strings.SplitN(arg, libraryKeySep, 2)
		switch len(strPieces) {
		case 1:
			return "", arg, nil
		case 2:
			if len(strPieces[0]) == 1 {
				return "", "", fmt.Errorf("Expected library ref to not be empty")
			}
			return strPieces[0], strPieces[1], nil
		}
	}
	return "", arg, nil
}

func (s *DataValuesFlags) buildOverlay(keyPieces []string, value interface{}, desc string, line string) *yamlmeta.Document {
	resultMap := &yamlmeta.Map{}
	currMap := resultMap
	var lastMapItem *yamlmeta.MapItem

	pos := filepos.NewPositionInFile(1, desc)
	pos.SetLine(line)

	for _, piece := range keyPieces {
		newMap := &yamlmeta.Map{}
		lastMapItem = &yamlmeta.MapItem{Key: piece, Value: newMap, Position: pos}

		// Data values schemas should be enough to provide key checking/validations.
		lastMapItem.SetAnnotations(template.NodeAnnotations{
			yttoverlay.AnnotationMatch: template.NodeAnnotation{
				Kwargs: []starlark.Tuple{{
					starlark.String(yttoverlay.MatchAnnotationKwargMissingOK),
					starlark.Bool(true),
				}},
			},
		})

		currMap.Items = append(currMap.Items, lastMapItem)
		currMap = newMap
	}

	lastMapItem.Value = yamlmeta.NewASTFromInterface(value)

	// Explicitly replace entire value at given key
	// (this allows to specify non-scalar data values)
	existingAnns := template.NewAnnotations(lastMapItem)
	existingAnns[yttoverlay.AnnotationReplace] = template.NodeAnnotation{
		Kwargs: []starlark.Tuple{{
			starlark.String(yttoverlay.ReplaceAnnotationKwargOrAdd),
			starlark.Bool(true),
		}},
	}
	lastMapItem.SetAnnotations(existingAnns)

	return &yamlmeta.Document{Value: resultMap, Position: pos}
}

// asFiles enumerates the files that are found at "path"
//
// If a DataValuesFlags.ReadFilesFunc has been injected, that service is used.
// Otherwise, uses files.NewSortedFilesFromPaths() is used.
func (s *DataValuesFlags) asFiles(path string) ([]*files.File, error) {
	if s.ReadFilesFunc != nil {
		return s.ReadFilesFunc(path)
	}
	return files.NewSortedFilesFromPaths([]string{path}, *s.SymlinkAllowOpts)
}
