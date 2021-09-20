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
	// IPAPI describes the contents of "@ytt:ip" module of the ytt standard library.
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

const ipAddrTypeName = "ip.addr"

// Type reports the name of this type as seen from a Starlark program (i.e. via the `type()` built-in)
func (av *IPAddrValue) Type() string { return "@ytt:" + ipAddrTypeName }

// AsStarlarkValue converts this instance into a value suitable for use in a Starlark program.
func (av *IPAddrValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("is_ipv4", starlark.NewBuiltin(ipAddrTypeName+".is_ipv4", core.ErrWrapper(av.IsIPv4)))
	m.Set("is_ipv6", starlark.NewBuiltin(ipAddrTypeName+".is_ipv6", core.ErrWrapper(av.IsIPv6)))
	m.Set("string", starlark.NewBuiltin(ipAddrTypeName+".string", core.ErrWrapper(av.string)))
	av.StarlarkStruct = core.NewStarlarkStruct(m)
	return av
}

// ConversionHint provides a hint on how the user can explicitly convert this value to a type that can be automatically encoded.
func (av *IPAddrValue) ConversionHint() string {
	return av.Type() + " does not automatically encode (hint: use .string())"
}

// IsIPv4 is a core.StarlarkFunc that reveals whether this value is an IPv4 address.
func (av *IPAddrValue) IsIPv4(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV4 := av.addr != nil && av.addr.To4() != nil
	return starlark.Bool(isV4), nil
}

// IsIPv6 is a core.StarlarkFunc that reveals whether this value is an IPv6 address.
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

// ParseCIDR is a core.StarlarkFunc that extracts the IP address and IP network value from a CIDR expression
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

// IPNetValue holds the data for an instance of an IP Network value
type IPNetValue struct {
	net                  *net.IPNet
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

const ipNetTypeName = "ip.net"

// Type reports the name of this type as seen from a Starlark program (i.e. via the `type()` built-in)
func (inv *IPNetValue) Type() string { return "@ytt:" + ipNetTypeName }

// AsStarlarkValue converts this instance into a value suitable for use in a Starlark program.
func (inv *IPNetValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("addr", starlark.NewBuiltin(ipNetTypeName+".addr", core.ErrWrapper(inv.Addr)))
	m.Set("string", starlark.NewBuiltin(ipNetTypeName+".string", core.ErrWrapper(inv.string)))
	inv.StarlarkStruct = core.NewStarlarkStruct(m)
	return inv
}

// ConversionHint provides a hint on how the user can explicitly convert this value to a type that can be automatically encoded.
func (inv *IPNetValue) ConversionHint() string {
	return inv.Type() + " does not automatically encode (hint: use .string())"
}

// Addr is a core.StarlarkFunc that returns the masked address portion of the network value
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
