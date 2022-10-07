// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
)

var (
	GzipAPI = starlark.StringDict{
		"gzip": &starlarkstruct.Module{
			Name: "gzip",
			Members: starlark.StringDict{
				"compress":   starlark.NewBuiltin("gzip.compress", core.ErrWrapper(gzipModule{}.Compress)),
				"decompress": starlark.NewBuiltin("gzip.decompress", core.ErrWrapper(gzipModule{}.Decompress)),
			},
		},
	}
)

type gzipModule struct{}

func (b gzipModule) Compress(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write([]byte(val)); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return starlark.String(buf.String()), nil
}

func (b gzipModule) Decompress(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	buf := bytes.NewBufferString(val)

	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if _, err := out.ReadFrom(gz); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return starlark.String(out.String()), nil
}
