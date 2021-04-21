// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package ref

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
)

const (
	AnnotationLibraryRef = "library/ref"

	librarySep            = "@"
	libraryAliasIndicator = "~"
)

type LibraryRef struct {
	Path  string
	Alias string
}

func (p LibraryRef) Matches(lpp LibraryRef) bool {
	pathMatch := p.Path == lpp.Path
	if p.Alias == "" {
		return pathMatch
	}

	aliasMatch := p.Alias == lpp.Alias
	if p.Path == "" {
		return aliasMatch
	}

	return aliasMatch && pathMatch
}

func (p LibraryRef) AsString() string {
	if p.Alias == "" {
		return p.Path
	}
	return p.Path + libraryAliasIndicator + p.Alias
}

type LibraryRefExtractor struct {
}

func (n LibraryRefExtractor) FromAnnotation(anns template.NodeAnnotations) ([]LibraryRef, error) {
	var libRef []LibraryRef

	if hasLibAnn := anns.Has(AnnotationLibraryRef); hasLibAnn {
		libArgs := anns.Args(AnnotationLibraryRef)
		if l := libArgs.Len(); l != 1 {
			return nil, fmt.Errorf("Expected %s annotation to have one arg, got %d", AnnotationLibraryRef, l)
		}

		argString, err := core.NewStarlarkValue(libArgs[0]).AsString()
		if err != nil {
			return nil, err
		}

		libRef, err = n.FromStr(argString)
		if err != nil {
			return nil, fmt.Errorf("Annotation %s: %s", AnnotationLibraryRef, err.Error())
		}
	}
	return libRef, nil
}

func (n LibraryRefExtractor) FromStr(libRefStr string) ([]LibraryRef, error) {
	if libRefStr == "" {
		return nil, fmt.Errorf("Expected library ref to not be empty")
	}

	if !strings.HasPrefix(libRefStr, librarySep) {
		return nil, fmt.Errorf("Expected library ref to start with '%s'", librarySep)
	}

	var result []LibraryRef
	for _, refPiece := range strings.Split(libRefStr, librarySep)[1:] {
		pathAndAlias := strings.Split(refPiece, libraryAliasIndicator)
		switch l := len(pathAndAlias); {
		case l == 1:
			result = append(result, LibraryRef{Path: pathAndAlias[0]})

		case l == 2:
			if pathAndAlias[1] == "" {
				return nil, fmt.Errorf("Expected library alias to not be empty")
			}

			result = append(result, LibraryRef{Path: pathAndAlias[0], Alias: pathAndAlias[1]})

		default:
			return nil, fmt.Errorf("Expected library ref to have form: '@path', '@~alias', or '@path~alias', got: '%s'", libRefStr)
		}
	}
	return result, nil
}
