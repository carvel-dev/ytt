package yttlibrary

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/version"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
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
		return starlark.None, fmt.Errorf("ytt version %s does not meet the minimum required version %s", version.Version, val)
	}

	return starlark.None, nil
}
