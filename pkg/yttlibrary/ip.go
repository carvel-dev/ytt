// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"
	"net"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template/core"
)

var (
	IPAPI = starlark.StringDict{
		"ip": &starlarkstruct.Module{
			Name: "ip",
			Members: starlark.StringDict{
				"parse_addr": starlark.NewBuiltin("ip.parse_addr", core.ErrWrapper(ipModule{}.ParseAddr)),
				"parse_cidr": starlark.NewBuiltin("ip.parse_cidr", core.ErrWrapper(ipModule{}.ParseCIDR)),
			},
		},
	}
)

type ipModule struct{}

func (m ipModule) ParseAddr(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	ipStr, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return starlark.None, fmt.Errorf("invalid IP address: %s", ipStr)
	}
	return (&IPAddrValue{parsedIP, nil}).AsStarlarkValue(), nil
}

// IPAddrValue stores a parsed IP
type IPAddrValue struct {
	addr                 net.IP
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

func (av *IPAddrValue) Type() string { return "@ytt:ip.addr" }

func (av *IPAddrValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("is_ipv4", starlark.NewBuiltin("ip.addr.is_ipv4", core.ErrWrapper(av.IsIPv4)))
	m.Set("is_ipv6", starlark.NewBuiltin("ip.addr.is_ipv6", core.ErrWrapper(av.IsIPv6)))
	m.Set("string", starlark.NewBuiltin("ip.addr.string", core.ErrWrapper(av.string)))
	av.StarlarkStruct = core.NewStarlarkStruct(m)
	return av
}

func (av *IPAddrValue) ConversionHint() string {
	return av.Type() + " does not automatically encode (hint: use .string())"
}

func (av *IPAddrValue) IsIPv4(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV4 := av.addr != nil && av.addr.To4() != nil
	return starlark.Bool(isV4), nil
}

func (av *IPAddrValue) IsIPv6(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV6 := av.addr != nil && av.addr.To4() == nil
	return starlark.Bool(isV6), nil
}

func (av *IPAddrValue) string(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	return starlark.String(av.addr.String()), nil
}

func (m ipModule) ParseCIDR(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	cidrStr, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	parsedIP, parsedNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return starlark.None, err
	}

	return starlark.Tuple{(&IPAddrValue{parsedIP, nil}).AsStarlarkValue(), (&IPNetValue{parsedNet, nil}).AsStarlarkValue()}, nil
}

type IPNetValue struct {
	net                  *net.IPNet
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

func (inv *IPNetValue) Type() string { return "@ytt:ip.net" }

func (inv *IPNetValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("addr", starlark.NewBuiltin("ip.net.addr", core.ErrWrapper(inv.Addr)))
	m.Set("string", starlark.NewBuiltin("ip.net.string", core.ErrWrapper(inv.string)))
	inv.StarlarkStruct = core.NewStarlarkStruct(m)
	return inv
}

func (inv *IPNetValue) ConversionHint() string {
	return inv.Type() + " does not automatically encode (hint: use .string())"
}

func (inv *IPNetValue) Addr(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}

	ip := &IPAddrValue{inv.net.IP, nil}
	return ip.AsStarlarkValue(), nil
}

func (inv *IPNetValue) string(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	return starlark.String(inv.net.String()), nil
}
