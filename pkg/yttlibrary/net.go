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
	NetAPI = starlark.StringDict{
		"net": &starlarkstruct.Module{
			Name: "net",
			Members: starlark.StringDict{
				"parse_ip":   starlark.NewBuiltin("net.parse_ip", core.ErrWrapper(netModule{}.ParseIP)),
				"parse_cidr": starlark.NewBuiltin("net.parse_cidr", core.ErrWrapper(netModule{}.ParseCIDR)),
			},
		},
	}
	NetKWARGS = map[string]struct{}{
		"return_error": struct{}{},
	}
)

type netModule struct{}

// IPValue stores a parsed IP
type IPValue struct {
	ip                   net.IP
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

func (b netModule) ParseIP(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}
	if err := core.CheckArgNames(kwargs, NetKWARGS); err != nil {
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
	return valueOrTuple((&IPValue{parsedIP, nil}).AsStarlarkValue(), nil, returnTuple)
}

func (iv *IPValue) Type() string { return "@ytt:net.ip" }

func (iv *IPValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("is_ipv4", starlark.NewBuiltin("ip.is_ipv4", core.ErrWrapper(iv.IsIPv4)))
	m.Set("is_ipv6", starlark.NewBuiltin("ip.is_ipv6", core.ErrWrapper(iv.IsIPv6)))
	iv.StarlarkStruct = core.NewStarlarkStruct(m)
	return iv
}

func (iv *IPValue) AsGoValue() (interface{}, error) {
	return iv.ip.String(), nil
}

func (iv *IPValue) IsIPv4(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV4 := iv.ip != nil && iv.ip.To4() != nil
	return starlark.Bool(isV4), nil
}

func (iv *IPValue) IsIPv6(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV6 := iv.ip != nil && iv.ip.To4() == nil
	return starlark.Bool(isV6), nil
}

// CIDRValue stores a parsed CIDR
type CIDRValue struct {
	ip                   net.IP
	ipNet                *net.IPNet
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

func (cv *CIDRValue) Type() string { return "@ytt:net.cidr" }

func (b netModule) ParseCIDR(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}
	if err := core.CheckArgNames(kwargs, NetKWARGS); err != nil {
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
	m.Set("ip", cv.IP())
	m.Set("net", cv.IPNet())
	cv.StarlarkStruct = core.NewStarlarkStruct(m)
	return cv
}

func (cv *CIDRValue) AsGoValue() (interface{}, error) {
	hostIPNet := net.IPNet{
		IP:   cv.ip,
		Mask: cv.ipNet.Mask,
	}
	return hostIPNet.String(), nil
}

func (cv *CIDRValue) IP() starlark.Value {
	if cv.ip == nil {
		return starlark.None
	}

	ip := &IPValue{cv.ip, nil}
	return ip.AsStarlarkValue()
}

func (cv *CIDRValue) IPNet() starlark.Value {
	if cv.ipNet == nil {
		return starlark.None
	}

	ipNet := &IPNetValue{cv.ipNet, nil}
	m := orderedmap.NewMap()
	m.Set("ip", ipNet.IP())
	ipNet.StarlarkStruct = core.NewStarlarkStruct(m)
	return ipNet
}

func (cv *CIDRValue) isIPv4(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV4 := cv.ipNet != nil && cv.ip.To4() != nil
	return starlark.Bool(isV4), nil
}

func (cv *CIDRValue) isIPv6(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	isV6 := cv.ip != nil && cv.ip.To4() == nil
	return starlark.Bool(isV6), nil
}

type IPNetValue struct {
	ipNet                *net.IPNet
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

func (inv *IPNetValue) Type() string { return "@ytt:net.net" }

func (inv *IPNetValue) AsGoValue() (interface{}, error) {
	return inv.ipNet.String(), nil
}

func (inv *IPNetValue) IP() starlark.Value {
	if inv.ipNet.IP == nil {
		return starlark.None
	}

	ip := &IPValue{inv.ipNet.IP, nil}
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
