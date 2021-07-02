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
	IPKWARGS = map[string]struct{}{
		"return_error": struct{}{},
	}
)

type ipModule struct{}

// AddrValue stores a parsed IP
type AddrValue struct {
	addr                 net.IP
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

func (m ipModule) ParseAddr(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}
	if err := core.CheckArgNames(kwargs, IPKWARGS); err != nil {
		return starlark.None, err
	}

	returnTuple, err := core.BoolArg(kwargs, "return_error")
	if err != nil {
		return valueOrTuple(starlark.None, err, returnTuple)
	}

	ipStr, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return valueOrTuple(starlark.None, err, returnTuple)
	}

	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return valueOrTuple(starlark.None, fmt.Errorf("invalid IP address: %s", ipStr), returnTuple)
	}
	return valueOrTuple((&AddrValue{parsedIP, nil}).AsStarlarkValue(), nil, returnTuple)
}

func (av *AddrValue) Type() string { return "@ytt:ip.addr" }

func (av *AddrValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("is_ipv4", starlark.NewBuiltin("addr.is_ipv4", core.ErrWrapper(av.IsIPv4)))
	m.Set("is_ipv6", starlark.NewBuiltin("addr.is_ipv6", core.ErrWrapper(av.IsIPv6)))
	av.StarlarkStruct = core.NewStarlarkStruct(m)
	return av
}

func (av *AddrValue) AsGoValue() (interface{}, error) {
	return av.addr.String(), nil
}

func (av *AddrValue) IsIPv4(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV4 := av.addr != nil && av.addr.To4() != nil
	return starlark.Bool(isV4), nil
}

func (av *AddrValue) IsIPv6(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV6 := av.addr != nil && av.addr.To4() == nil
	return starlark.Bool(isV6), nil
}

// CIDRValue stores a parsed CIDR
type CIDRValue struct {
	addr                 net.IP
	net                  *net.IPNet
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

func (cv *CIDRValue) Type() string { return "@ytt:ip.cidr" }

func (m ipModule) ParseCIDR(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}
	if err := core.CheckArgNames(kwargs, IPKWARGS); err != nil {
		return starlark.None, err
	}

	returnTuple, err := core.BoolArg(kwargs, "return_error")
	if err != nil {
		return starlark.None, err
	}

	cidrStr, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return valueOrTuple(starlark.None, err, returnTuple)
	}

	parsedIP, parsedNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return valueOrTuple(starlark.None, err, returnTuple)
	}

	return valueOrTuple((&CIDRValue{parsedIP, parsedNet, nil}).AsStarlarkValue(), nil, returnTuple)
}

func (cv *CIDRValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("is_ipv4", starlark.NewBuiltin("cidr.is_ipv4", core.ErrWrapper(cv.isIPv4)))
	m.Set("is_ipv6", starlark.NewBuiltin("cidr.is_ipv6", core.ErrWrapper(cv.isIPv6)))
	m.Set("addr", cv.Addr())
	m.Set("net", cv.Net())
	cv.StarlarkStruct = core.NewStarlarkStruct(m)
	return cv
}

func (cv *CIDRValue) AsGoValue() (interface{}, error) {
	hostIPNet := net.IPNet{
		IP:   cv.addr,
		Mask: cv.net.Mask,
	}
	return hostIPNet.String(), nil
}

func (cv *CIDRValue) Addr() starlark.Value {
	if cv.addr == nil {
		return starlark.None
	}

	av := &AddrValue{cv.addr, nil}
	return av.AsStarlarkValue()
}

func (cv *CIDRValue) Net() starlark.Value {
	if cv.net == nil {
		return starlark.None
	}

	ipNet := &NetValue{cv.net, nil}
	m := orderedmap.NewMap()
	m.Set("addr", ipNet.IP())
	ipNet.StarlarkStruct = core.NewStarlarkStruct(m)
	return ipNet
}

func (cv *CIDRValue) isIPv4(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV4 := cv.net != nil && cv.addr.To4() != nil
	return starlark.Bool(isV4), nil
}

func (cv *CIDRValue) isIPv6(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV6 := cv.addr != nil && cv.addr.To4() == nil
	return starlark.Bool(isV6), nil
}

type NetValue struct {
	net                  *net.IPNet
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

func (inv *NetValue) Type() string { return "@ytt:ip.net" }

func (inv *NetValue) AsGoValue() (interface{}, error) {
	return inv.net.String(), nil
}

func (inv *NetValue) IP() starlark.Value {
	if inv.net.IP == nil {
		return starlark.None
	}

	ip := &AddrValue{inv.net.IP, nil}
	return ip.AsStarlarkValue()
}

func valueOrTuple(value starlark.Value, err error, returnTuple bool) (starlark.Value, error) {
	if returnTuple {
		var errValue starlark.Value = starlark.None
		if err != nil {
			errValue = starlark.String(err.Error())
		}
		return starlark.Tuple{value, errValue}, nil
	}
	return value, err
}
