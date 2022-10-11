// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"
	"github.com/k14s/semver/v4"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/vmware-tanzu/carvel-ytt/pkg/orderedmap"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
	"strings"
)

var (
	SemverAPI = starlark.StringDict{
		"semver": &starlarkstruct.Module{
			Name: "semver",
			Members: starlark.StringDict{
				"version":  starlark.NewBuiltin("semver.version", core.ErrWrapper(semverModule{}.Version)),
				"from_str": starlark.NewBuiltin("semver.version", core.ErrWrapper(semverModule{}.FromStr)),
				"range":    starlark.NewBuiltin("semver.version", core.ErrWrapper(semverModule{}.Range)),
				//"range":    starlark.NewBuiltin("semver.version", core.ErrWrapper(semverModule{}.Range)),
			},
		},
	}
)

type semverModule struct{}

func (s semverModule) Version(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() > 5 || len(kwargs) > 5 || args.Len()+len(kwargs) > 5 {
		return starlark.None, fmt.Errorf("expected at most five arguments in total")
	}

	for _, kwarg := range kwargs {
		key := string(kwarg[0].(starlark.String))
		if key != "major" && key != "minor" && key != "patch" && key != "prerelease" && key != "build" {
			return starlark.None, fmt.Errorf("unexpected kwarg %q. expected kwargs are [major, minor, patch, prerelease, build]", key)
		}
	}

	var (
		major, minor, patch starlark.Int
		prerelease, build   starlark.String
	)

	err := starlark.UnpackArgs("version", args, kwargs, "major?", &major, "minor?", &minor, "patch?", &patch, "prerelease?", &prerelease, "build?", &build)
	if err != nil {
		return nil, err
	}

	majorInt, err := core.NewStarlarkValue(major).AsInt64()
	if err != nil {
		return nil, err
	}
	minorInt, err := core.NewStarlarkValue(minor).AsInt64()
	if err != nil {
		return nil, err
	}
	patchInt, err := core.NewStarlarkValue(patch).AsInt64()
	if err != nil {
		return nil, err
	}

	var versionStr strings.Builder
	versionStr.WriteString(fmt.Sprintf("%d.%d.%d", majorInt, minorInt, patchInt))

	if prerelease.GoString() != "" {
		versionStr.WriteString(fmt.Sprintf("-%s", prerelease.GoString()))
	}
	if build.GoString() != "" {
		versionStr.WriteString(fmt.Sprintf("+%s", build.GoString()))
	}

	v, err := semver.Parse(versionStr.String())
	if err != nil {
		return nil, err
	}

	return (&SemverValue{v, nil}).AsStarlarkValue(), nil
}

const semverVersionTypeName = "semver.version"

type SemverValue struct {
	version semver.Version
	*core.StarlarkStruct
}

func (sv *SemverValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("major", starlark.MakeInt64(int64(sv.version.Major)))
	m.Set("minor", starlark.MakeInt64(int64(sv.version.Minor)))
	m.Set("patch", starlark.MakeInt64(int64(sv.version.Patch)))
	var prVersions []string
	for _, p := range sv.version.Pre {
		prVersions = append(prVersions, p.String())
	}
	m.Set("prerelease", starlark.String(strings.Join(prVersions, ".")))
	var buildVersions []string
	for _, p := range sv.version.Build {
		buildVersions = append(buildVersions, p.String())
	}
	m.Set("build", starlark.String(strings.Join(buildVersions, ".")))
	m.Set("dict", starlark.NewBuiltin(fmt.Sprintf("%s.dict", semverVersionTypeName), core.ErrWrapper(sv.dict)))
	m.Set("string", starlark.NewBuiltin(fmt.Sprintf("%s.string", semverVersionTypeName), core.ErrWrapper(sv.string)))
	m.Set("cmp", starlark.NewBuiltin(fmt.Sprintf("%s.string", semverVersionTypeName), core.ErrWrapper(sv.cmp)))
	m.Set("next_patch", starlark.NewBuiltin(fmt.Sprintf("%s.string", semverVersionTypeName), core.ErrWrapper(sv.nextPatch)))
	m.Set("next_minor", starlark.NewBuiltin(fmt.Sprintf("%s.string", semverVersionTypeName), core.ErrWrapper(sv.nextMinor)))
	m.Set("next_major", starlark.NewBuiltin(fmt.Sprintf("%s.string", semverVersionTypeName), core.ErrWrapper(sv.nextMajor)))
	sv.StarlarkStruct = core.NewStarlarkStruct(m)
	return sv
}

func (sv *SemverValue) Type() string { return fmt.Sprintf("@ytt:%s", semverVersionTypeName) }

//func (sv *SemverValue) ConversionHint() string {
//	return sv.Type() + " does not automatically encode (hint: use .string())"
//}

func (sv *SemverValue) dict(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}

	m := orderedmap.NewMap()

	m.Set("major", starlark.MakeInt64(int64(sv.version.Major)))
	m.Set("minor", starlark.MakeInt64(int64(sv.version.Minor)))
	m.Set("patch", starlark.MakeInt64(int64(sv.version.Patch)))

	var prVersions []string
	for _, p := range sv.version.Pre {
		prVersions = append(prVersions, p.String())
	}
	m.Set("prerelease", starlark.String(strings.Join(prVersions, ".")))

	var buildVersions []string
	for _, p := range sv.version.Build {
		buildVersions = append(buildVersions, p.String())
	}
	m.Set("build", starlark.String(strings.Join(buildVersions, ".")))

	return core.NewStarlarkStruct(m), nil
}

func (sv *SemverValue) string(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}

	return starlark.String(sv.version.String()), nil
}

func (sv *SemverValue) cmp(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	other, ok := (args.Index(0)).(*SemverValue)
	if !ok {
		return starlark.None, fmt.Errorf("expected argument to be a %s", semverVersionTypeName)
	}

	return starlark.MakeInt64(int64(sv.version.Compare(other.version))), nil
}

func (sv *SemverValue) nextPatch(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}

	// Copy to not mutate the original
	next := sv.version
	err := next.IncrementPatch()
	if err != nil {
		return nil, err
	}

	return (&SemverValue{next, nil}).AsStarlarkValue(), nil
}

func (sv *SemverValue) nextMinor(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}

	// Copy to not mutate the original
	next := sv.version
	err := next.IncrementMinor()
	if err != nil {
		return nil, err
	}

	return (&SemverValue{next, nil}).AsStarlarkValue(), nil
}

func (sv *SemverValue) nextMajor(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}

	// Copy to not	 mutate the original
	next := sv.version
	err := next.IncrementMajor()
	if err != nil {
		return nil, err
	}

	return (&SemverValue{next, nil}).AsStarlarkValue(), nil
}

func (s semverModule) FromStr(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	versionStr, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	v, err := semver.Parse(versionStr)
	if err != nil {
		return nil, err
	}

	return (&SemverValue{v, nil}).AsStarlarkValue(), nil
}

func (s semverModule) Range(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	rangeStr, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	v, err := semver.ParseRange(rangeStr)
	if err != nil {
		return nil, err
	}

	return (&RangeValue{v, nil}).AsStarlarkValue(), nil
}

const semverRangeTypeName = "semver.range"

type RangeValue struct {
	versionRange semver.Range
	*core.StarlarkStruct
}

func (rv *RangeValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("contains", starlark.NewBuiltin(fmt.Sprintf("%s.string", semverVersionTypeName), core.ErrWrapper(rv.contains)))
	rv.StarlarkStruct = core.NewStarlarkStruct(m)
	return rv
}

func (rv *RangeValue) Type() string { return fmt.Sprintf("@ytt:%s", semverRangeTypeName) }

func (rv *RangeValue) ConversionHint() string {
	return rv.Type() + " does not automatically encode (hint: use .string())"
}

func (rv *RangeValue) contains(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	version, ok := (args.Index(0)).(*SemverValue)
	if !ok {
		return starlark.None, fmt.Errorf("expected argument to be a %s", semverVersionTypeName)
	}

	return starlark.Bool(rv.versionRange(version.version)), nil
}
