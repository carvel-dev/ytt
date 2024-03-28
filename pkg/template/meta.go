// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"strings"

	"carvel.dev/ytt/pkg/filepos"
)

type AnnotationName string
type AnnotationNs string

const (
	// AnnotationComment contains user-facing documentation and should be ignored.
	AnnotationComment AnnotationName = "comment"
	// AnnotationCode contains Starlark code that should be inserted, verbatim, into the compiled template.
	AnnotationCode AnnotationName = "template/code"
	// AnnotationValue contains a Starlark expression, the result of which should be set as the value of the annotated node.
	AnnotationValue AnnotationName = "template/value"
)

type Annotation struct {
	Name     AnnotationName // eg template/code
	Content  string         // eg if True:
	Position *filepos.Position
}

// Supported formats:
//   "! comment"
//   "@comment content"
//   "@ if True:"
//   "@template/code"
//   "@template/code if True:"
//   "@text/trim-left,text/trim-right,template/code if True:"

type MetaOpts struct {
	IgnoreUnknown bool
}

// NewAnnotationFromComment constructs an Annotation from a string and position from a Comment.
//
// if opts.IgnoreUnknown is true and the annotation is unknown, it is returned as a comment.
// if opts.IgnoreUnknown is false and the annotation is unknown, returns an error.
func NewAnnotationFromComment(data string, position *filepos.Position, opts MetaOpts) (Annotation, error) {
	position = position.DeepCopy()
	switch {
	case len(data) > 0 && data[0] == '!':
		return Annotation{
			Name:     AnnotationComment,
			Content:  data[1:],
			Position: position,
		}, nil

	case len(data) > 0 && data[0] == '@':
		nameAndContent := strings.SplitN(data[1:], " ", 2)
		ann := Annotation{
			Name:     AnnotationName(nameAndContent[0]),
			Position: position,
		}
		if len(nameAndContent) == 2 {
			ann.Content = nameAndContent[1]
		}
		return ann, nil

	default:
		if opts.IgnoreUnknown {
			return Annotation{
				Name:     AnnotationComment,
				Content:  data,
				Position: position,
			}, nil
		} else {
			return Annotation{}, fmt.Errorf("Expected ytt-formatted string (use '#@' for annotations or code, '#!' for comments)")
		}
	}
}
