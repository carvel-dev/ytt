// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/workspace/datavalues"
	"github.com/vmware-tanzu/carvel-ytt/pkg/workspace/ref"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	yttoverlay "github.com/vmware-tanzu/carvel-ytt/pkg/yttlibrary/overlay"
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

	Inspect       bool
	InspectSchema bool

	EnvironFunc  func() []string
	ReadFileFunc func(string) ([]byte, error)
}

func (s *DataValuesFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&s.EnvFromStrings, "data-values-env", nil, "Extract data values (as strings) from prefixed env vars (format: PREFIX for PREFIX_all__key1=str) (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&s.EnvFromYAML, "data-values-env-yaml", nil, "Extract data values (parsed as YAML) from prefixed env vars (format: PREFIX for PREFIX_all__key1=true) (can be specified multiple times)")

	cmd.Flags().StringArrayVarP(&s.KVsFromStrings, "data-value", "v", nil, "Set specific data value to given value, as string (format: all.key1.subkey=123) (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&s.KVsFromYAML, "data-value-yaml", nil, "Set specific data value to given value, parsed as YAML (format: all.key1.subkey=true) (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&s.KVsFromFiles, "data-value-file", nil, "Set specific data value to given file contents, as string (format: all.key1.subkey=/file/path) (can be specified multiple times)")

	cmd.Flags().StringArrayVar(&s.FromFiles, "data-values-file", nil, "Set multiple data values via a YAML file (format: /file/path.yml) (can be specified multiple times)")

	cmd.Flags().BoolVar(&s.Inspect, "data-values-inspect", false, "Calculate the final data values (applying any overlays) and display that result")
	cmd.Flags().BoolVar(&s.InspectSchema, "data-values-schema-inspect", false, "Determine the complete schema for data values (applying any overlays) and display the result (only OpenAPI v3.0 is supported, see --output)")
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
	libRef, path, err := s.libraryRefAndKey(fullPath)
	if err != nil {
		return nil, err
	}

	contents, err := s.readFile(path)
	if err != nil {
		return nil, fmt.Errorf("Reading file '%s': %s", path, err)
	}

	docSetOpts := yamlmeta.DocSetOpts{
		AssociatedName: path,
		Strict:         strict,
	}

	docSet, err := yamlmeta.NewDocumentSetFromBytes(contents, docSetOpts)
	if err != nil {
		return nil, fmt.Errorf("Unmarshaling YAML data values file '%s': %s", path, err)
	}

	var result []*datavalues.Envelope

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

	libRef, keyPrefix, err := s.libraryRefAndKeyStrict(prefix)
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

	libRef, key, err := s.libraryRefAndKeyStrict(pieces[0])
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

	contents, err := s.readFile(pieces[1])
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
func (DataValuesFlags) libraryRefAndKeyStrict(key string) (string, string, error) {
	libRef, key, err := DataValuesFlags{}.libraryRefAndKey(key)
	if err != nil {
		return "", "", err
	}
	if len(strings.Split(key, libraryKeySep)) > 1 {
		// error on a common syntax mistake
		return "", "", fmt.Errorf("Expected at most one library-key separator '%s' in '%s'", libraryKeySep, key)
	}
	return libRef, key, nil
}

// libraryRefAndKey separates a library reference and a key.
// A library reference starts with ref.LibrarySep, and ends with the first occurrence of libraryKeySep.
func (DataValuesFlags) libraryRefAndKey(key string) (string, string, error) {
	if strings.HasPrefix(key, ref.LibrarySep) {
		keyPieces := strings.SplitN(key, libraryKeySep, 2)
		switch len(keyPieces) {
		case 1:
			return "", key, nil
		case 2:
			if len(keyPieces[0]) == 1 {
				return "", "", fmt.Errorf("Expected library ref to not be empty")
			}
			return keyPieces[0], keyPieces[1], nil
		}
	}
	return "", key, nil
}

func (s *DataValuesFlags) buildOverlay(keyPieces []string, value interface{}, desc string, line string) *yamlmeta.Document {
	resultMap := &yamlmeta.Map{}
	currMap := resultMap
	var lastMapItem *yamlmeta.MapItem

	pos := filepos.NewPosition(1)
	pos.SetFile(desc)
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

func (s *DataValuesFlags) readFile(path string) ([]byte, error) {
	if s.ReadFileFunc != nil {
		return s.ReadFileFunc(path)
	}
	if path == "-" {
		return files.ReadStdin()
	}
	return ioutil.ReadFile(path)
}
