// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"strings"

	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
)

type AnnotationName string
type AnnotationNs string

const (
	AnnotationNameComment AnnotationName = "comment"
)

type Meta struct {
	Annotations []*Annotation
}

type Annotation struct {
	Name     AnnotationName // eg template/code
	Content  string         // eg if True:
	Position *filepos.Position
}

type Annotations struct {
	tagToNodeAnnotations map[NodeTag]map[AnnotationName]Annotation
}

func NewAnnotationsForTemplate() *Annotations {
	return &Annotations{
		tagToNodeAnnotations: map[NodeTag]map[AnnotationName]Annotation{},
	}
}

func (a *Annotations) AddAnnotation(tag NodeTag, ann Annotation) {
	if _, ok := a.tagToNodeAnnotations[tag]; !ok {
		a.tagToNodeAnnotations[tag] = map[AnnotationName]Annotation{}
	}
	a.tagToNodeAnnotations[tag][ann.Name] = ann
}

func (a *Annotations) FindAnnotation(tag NodeTag, annName AnnotationName) (Annotation, bool) {
	ann, ok := a.tagToNodeAnnotations[tag][annName]
	return ann, ok
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

// NewAnnotationFromString constructs an Annotation from a given string.
//
// if opts.IgnoreUnknown is true and the annotation is unknown, it is returned as a comment.
// if opts.IgnoreUnknown is false and the annotation is unknown, returns an error.
func NewAnnotationFromString(data string, opts MetaOpts) (Annotation, error) {
	switch {
	case len(data) > 0 && data[0] == '!':
		return Annotation{
			Name:    AnnotationNameComment,
			Content: data[1:],
		}, nil

	case len(data) > 0 && data[0] == '@':
		nameAndContent := strings.SplitN(data[1:], " ", 2)
		ann := Annotation{
			Name: AnnotationName(nameAndContent[0]),
		}
		if len(nameAndContent) == 2 {
			ann.Content = nameAndContent[1]
		}
		return ann, nil

	default:
		if opts.IgnoreUnknown {
			return Annotation{
				Name:    AnnotationNameComment,
				Content: data,
			}, nil
		} else {
			return Annotation{}, fmt.Errorf("Expected ytt-formatted string (use '#@' for annotations or code, '#!' for comments)")
		}
	}
}
