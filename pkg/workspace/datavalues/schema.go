// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package datavalues

import (
	"fmt"
	"strings"

	"carvel.dev/ytt/pkg/schema"
	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/workspace/ref"
	"carvel.dev/ytt/pkg/yamlmeta"
)

// Schema is a definition of types and default values for Envelope.
type Schema struct {
	Source     *yamlmeta.Document
	defaultDVs *yamlmeta.Document
	DocType    *schema.DocumentType
}

// SchemaEnvelope is addressing and usage bookkeeping for a Schema — for which library this Schema is intended.
type SchemaEnvelope struct {
	Doc *Schema

	used           bool
	originalLibRef []ref.LibraryRef
	libRef         []ref.LibraryRef
}

// NewSchema calculates a Schema from a YAML document containing schema.
func NewSchema(doc *yamlmeta.Document) (*Schema, error) {
	docType, err := schema.InferTypeFromValue(doc, doc.Position)
	if err != nil {
		return nil, err
	}

	schemaDVs := docType.GetDefaultValue()

	return &Schema{
		Source:     doc,
		defaultDVs: schemaDVs.(*yamlmeta.Document),
		DocType:    docType.(*schema.DocumentType),
	}, nil
}

// NewSchemaEnvelope generates a new Schema wrapped in a SchemaEnvelope form a YAML document containing schema.
func NewSchemaEnvelope(doc *yamlmeta.Document) (*SchemaEnvelope, error) {
	libRef, err := getSchemaLibRef(ref.LibraryRefExtractor{}, doc)
	if err != nil {
		return nil, err
	}

	schema, err := NewSchema(doc)
	if err != nil {
		return nil, err
	}

	return &SchemaEnvelope{
		Doc:            schema,
		originalLibRef: libRef,
		libRef:         libRef,
	}, nil
}

// NewNullSchema provides the "Null Object" value of Schema. This is used in the case where no schema was provided.
func NewNullSchema() *Schema {
	return &Schema{
		Source: &yamlmeta.Document{},
		DocType: &schema.DocumentType{
			ValueType: &schema.AnyType{}},
	}
}

// ExtractLibRefs constructs library references (ref.LibraryRef) from various sources.
type ExtractLibRefs interface {
	FromStr(string) ([]ref.LibraryRef, error)
	FromAnnotation(template.NodeAnnotations) ([]ref.LibraryRef, error)
}

func getSchemaLibRef(libRefs ExtractLibRefs, doc *yamlmeta.Document) ([]ref.LibraryRef, error) {
	anns := template.NewAnnotations(doc)
	libRef, err := libRefs.FromAnnotation(anns)
	if err != nil {
		return nil, err
	}
	return libRef, nil
}

// AssignType decorates `doc` with type metadata sourced from this Schema.
// If `doc` does not conform to the AST structure of this Schema, the returned TypeCheck contains the violations.
// No other type check is performed.
func (s *Schema) AssignType(doc *yamlmeta.Document) schema.TypeCheck {
	return s.DocType.AssignTypeTo(doc)
}

// DefaultDataValues returns a copy of the default values declared in this Schema.
func (s *Schema) DefaultDataValues() *yamlmeta.Document {
	if s.defaultDVs == nil {
		return nil
	}
	return s.defaultDVs.DeepCopy()
}

// GetDocumentType returns a reference to the DocumentType that is the root of this Schema.
func (s *Schema) GetDocumentType() *schema.DocumentType {
	return s.DocType
}

// DeepCopy produces a complete copy of this Schema.
func (s *Schema) DeepCopy() *Schema {
	return &Schema{
		Source:     s.Source.DeepCopy(),
		defaultDVs: s.defaultDVs.DeepCopy(),
		DocType:    s.DocType,
	}
}

// Source yields the original YAML document used to construct this SchemaEnvelope.
func (e *SchemaEnvelope) Source() *yamlmeta.Document {
	return e.Doc.Source
}

// Desc reports which library the contained Schema is intended.
func (e *SchemaEnvelope) Desc() string {
	var desc []string
	for _, refPiece := range e.originalLibRef {
		desc = append(desc, refPiece.AsString())
	}
	return fmt.Sprintf("Schema belonging to library '%s%s' on %s", "@",
		strings.Join(desc, "@"), e.Source().Position.AsString())
}

// IsUsed reports whether or not the contained Schema was delivered/consumed.
func (e *SchemaEnvelope) IsUsed() bool { return e.used }

// IntendedForAnotherLibrary indicates whether the contained Schema is addressed to another library.
func (e *SchemaEnvelope) IntendedForAnotherLibrary() bool {
	return len(e.libRef) > 0
}

// UsedInLibrary marks this SchemaEnvelope as "delivered"/used if its destination is included in expectedRefPiece.
//
// If the SchemaEnvelope should be used exactly in the specified library, returns a copy of this SchemaEnvelope with no
// addressing and true.
// If the SchemaEnvelope should be used in a _child_ of the specified library, returns a copy of this SchemaEnvelope
// with the address of _that_ library and true.
// If the SchemaEnvelope should **not** be used in the specified library or child, returns nil and false.
func (e *SchemaEnvelope) UsedInLibrary(expectedRefPiece ref.LibraryRef) (*SchemaEnvelope, bool) {
	if !e.IntendedForAnotherLibrary() {
		e.markUsed()

		return e.deepCopyUnused(), true
	}

	if !e.libRef[0].Matches(expectedRefPiece) {
		return nil, false
	}
	e.markUsed()
	childSchemaProcessing := e.deepCopyUnused()
	childSchemaProcessing.libRef = childSchemaProcessing.libRef[1:]
	return childSchemaProcessing, !childSchemaProcessing.IntendedForAnotherLibrary()
}

func (e *SchemaEnvelope) markUsed() { e.used = true }

func (e *SchemaEnvelope) deepCopyUnused() *SchemaEnvelope {
	var copiedPieces []ref.LibraryRef
	copiedPieces = append(copiedPieces, e.libRef...)
	return &SchemaEnvelope{
		Doc:            e.Doc.DeepCopy(),
		originalLibRef: e.originalLibRef,
		libRef:         copiedPieces,
	}
}
