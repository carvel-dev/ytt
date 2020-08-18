// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

type DataModule struct {
	values starlark.Value
	loader DataLoader
}

type DataLoader interface {
	FilePaths(string) ([]string, error)
	FileData(string) ([]byte, error)
}

func NewDataModule(values *yamlmeta.Document, loader DataLoader) DataModule {
	val := core.NewGoValueWithOpts(values.AsInterface(), core.GoValueOpts{MapIsStruct: true})
	return DataModule{val.AsStarlarkValue(), loader}
}

func (b DataModule) AsModule() starlark.StringDict {
	return starlark.StringDict{
		"data": &starlarkstruct.Module{
			Name: "data",
			Members: starlark.StringDict{
				"list": starlark.NewBuiltin("data.list", core.ErrWrapper(b.List)),
				"read": starlark.NewBuiltin("data.read", core.ErrWrapper(b.Read)),
				// TODO write?
				"values": b.values,
			},
		},
	}
}

func (b DataModule) List(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() > 1 {
		return starlark.None, fmt.Errorf("expected exactly zero or one argument")
	}

	path := ""

	if args.Len() == 1 {
		pathStr, err := core.NewStarlarkValue(args.Index(0)).AsString()
		if err != nil {
			return starlark.None, err
		}
		path = pathStr
	}

	paths, err := b.loader.FilePaths(path)
	if err != nil {
		return starlark.None, err
	}

	result := []starlark.Value{}
	for _, path := range paths {
		result = append(result, starlark.String(path))
	}
	return starlark.NewList(result), nil
}

func (b DataModule) Read(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	path, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	fileBs, err := b.loader.FileData(path)
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(string(fileBs)), nil
}
