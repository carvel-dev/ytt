// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"strings"
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
	Name    AnnotationName // eg template/code
	Content string         // eg if True:
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

func NewMetaFromString(data string, opts MetaOpts) (Meta, error) {
	meta := Meta{}

	// TODO better error messages?
	switch {
	case len(data) > 0 && data[0] == '!':
		meta.Annotations = []*Annotation{{
			Name:    AnnotationNameComment,
			Content: data[1:],
		}}

	case len(data) > 0 && data[0] == '@':
		pieces := strings.SplitN(data[1:], " ", 2)
		for _, name := range strings.Split(pieces[0], ",") {
			meta.Annotations = append(meta.Annotations, &Annotation{
				Name: AnnotationName(name),
			})
		}
		if len(pieces) == 2 {
			meta.Annotations[len(meta.Annotations)-1].Content = pieces[1]
		}

	default:
		if opts.IgnoreUnknown {
			meta.Annotations = []*Annotation{{
				Name:    AnnotationNameComment,
				Content: data,
			}}
		} else {
			return Meta{}, fmt.Errorf("Unrecognized comment type (expected '#@' or '#!')")
		}
	}

	return meta, nil
}
