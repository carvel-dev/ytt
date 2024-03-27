// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"
	"regexp"

	"carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/version"
	semver "github.com/hashicorp/go-version"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

const (
	SemverRegex string = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)` +
		`(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))` +
		`?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
)

var (
	VersionAPI = starlark.StringDict{
		"version": &starlarkstruct.Module{
			Name: "version",
			Members: starlark.StringDict{
				"require_at_least": starlark.NewBuiltin("version.require_at_least", core.ErrWrapper(versionModule{}.RequireAtLeast)),
			},
		},
	}
)

type versionModule struct{}

func (b versionModule) RequireAtLeast(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	r := regexp.MustCompile(SemverRegex)
	if !r.MatchString(val) {
		return starlark.None, fmt.Errorf("version string '%s' must be a valid semver", val)
	}

	userConstraint, err := semver.NewConstraint(">=" + val)
	if err != nil {
		return starlark.None, err
	}

	yttVersion, err := semver.NewVersion(version.Version)
	if err != nil {
		return starlark.None, err
	}

	satisfied := userConstraint.Check(yttVersion)
	if !satisfied {
		return starlark.None, fmt.Errorf("ytt version '%s' does not meet the minimum required version '%s'", version.Version, val)
	}

	return starlark.None, nil
}
