// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// Copyright 2021 The Bazel Authors. All rights reserved.
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the
//    distribution.
//
// 3. Neither the name of the copyright holder nor the names of its
//    contributors may be used to endorse or promote products derived
//    from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package yttlibrary

import (
	"fmt"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
	"math"
)

// MathAPI contains the definition of the @ytt:math module.
// It contains math-related functions and constants.
// The module defines the following functions:
//
//     ceil(x) - Returns the ceiling of x, the smallest integer greater than or equal to x.
//     copysign(x, y) - Returns a value with the magnitude of x and the sign of y.
//     fabs(x) - Returns the absolute value of x as float.
//     floor(x) - Returns the floor of x, the largest integer less than or equal to x.
//     mod(x, y) - Returns the floating-point remainder of x/y. The magnitude of the result is less than y and its sign agrees with that of x.
//     pow(x, y) - Returns x**y, the base-x exponential of y.
//     remainder(x, y) - Returns the IEEE 754 floating-point remainder of x/y.
//     round(x) - Returns the nearest integer, rounding half away from zero.
//
//     exp(x) - Returns e raised to the power x, where e = 2.718281… is the base of natural logarithms.
//     sqrt(x) - Returns the square root of x.
//
//     acos(x) - Returns the arc cosine of x, in radians.
//     asin(x) - Returns the arc sine of x, in radians.
//     atan(x) - Returns the arc tangent of x, in radians.
//     atan2(y, x) - Returns atan(y / x), in radians.
//                   The result is between -pi and pi.
//                   The vector in the plane from the origin to point (x, y) makes this angle with the positive X axis.
//                   The point of atan2() is that the signs of both inputs are known to it, so it can compute the correct
//                   quadrant for the angle.
//                   For example, atan(1) and atan2(1, 1) are both pi/4, but atan2(-1, -1) is -3*pi/4.
//     cos(x) - Returns the cosine of x, in radians.
//     hypot(x, y) - Returns the Euclidean norm, sqrt(x*x + y*y). This is the length of the vector from the origin to point (x, y).
//     sin(x) - Returns the sine of x, in radians.
//     tan(x) - Returns the tangent of x, in radians.
//
//     degrees(x) - Converts angle x from radians to degrees.
//     radians(x) - Converts angle x from degrees to radians.
//
//     acosh(x) - Returns the inverse hyperbolic cosine of x.
//     asinh(x) - Returns the inverse hyperbolic sine of x.
//     atanh(x) - Returns the inverse hyperbolic tangent of x.
//     cosh(x) - Returns the hyperbolic cosine of x.
//     sinh(x) - Returns the hyperbolic sine of x.
//     tanh(x) - Returns the hyperbolic tangent of x.
//
//     log(x, base) - Returns the logarithm of x in the given base, or natural logarithm by default.
//
//     gamma(x) - Returns the Gamma function of x.
//
// All functions accept both int and float values as arguments.
//
// The module also defines approximations of the following constants:
//
//     e - The base of natural logarithms, approximately 2.71828.
//     pi - The ratio of a circle's circumference to its diameter, approximately 3.14159.
//
var MathAPI = starlark.StringDict{
	"math": &starlarkstruct.Module{
		Name: "math",
		Members: starlark.StringDict{
			"ceil":      starlark.NewBuiltin("ceil", core.ErrWrapper(mathModule{}.ceil)),
			"copysign":  mathModule{}.newBinaryBuiltin("copysign", math.Copysign),
			"fabs":      mathModule{}.newUnaryBuiltin("fabs", math.Abs),
			"floor":     starlark.NewBuiltin("floor", core.ErrWrapper(mathModule{}.floor)),
			"mod":       mathModule{}.newBinaryBuiltin("round", math.Mod),
			"pow":       mathModule{}.newBinaryBuiltin("pow", math.Pow),
			"remainder": mathModule{}.newBinaryBuiltin("remainder", math.Remainder),
			"round":     mathModule{}.newUnaryBuiltin("round", math.Round),

			"exp":  mathModule{}.newUnaryBuiltin("exp", math.Exp),
			"sqrt": mathModule{}.newUnaryBuiltin("sqrt", math.Sqrt),

			"acos":  mathModule{}.newUnaryBuiltin("acos", math.Acos),
			"asin":  mathModule{}.newUnaryBuiltin("asin", math.Asin),
			"atan":  mathModule{}.newUnaryBuiltin("atan", math.Atan),
			"atan2": mathModule{}.newBinaryBuiltin("atan2", math.Atan2),
			"cos":   mathModule{}.newUnaryBuiltin("cos", math.Cos),
			"hypot": mathModule{}.newBinaryBuiltin("hypot", math.Hypot),
			"sin":   mathModule{}.newUnaryBuiltin("sin", math.Sin),
			"tan":   mathModule{}.newUnaryBuiltin("tan", math.Tan),

			"degrees": mathModule{}.newUnaryBuiltin("degrees", mathModule{}.degrees),
			"radians": mathModule{}.newUnaryBuiltin("radians", mathModule{}.radians),

			"acosh": mathModule{}.newUnaryBuiltin("acosh", math.Acosh),
			"asinh": mathModule{}.newUnaryBuiltin("asinh", math.Asinh),
			"atanh": mathModule{}.newUnaryBuiltin("atanh", math.Atanh),
			"cosh":  mathModule{}.newUnaryBuiltin("cosh", math.Cosh),
			"sinh":  mathModule{}.newUnaryBuiltin("sinh", math.Sinh),
			"tanh":  mathModule{}.newUnaryBuiltin("tanh", math.Tanh),

			"log": starlark.NewBuiltin("log", mathModule{}.log),

			"gamma": mathModule{}.newUnaryBuiltin("gamma", math.Gamma),

			"e":  starlark.Float(math.E),
			"pi": starlark.Float(math.Pi),
		},
	},
}

type mathModule struct{}

// newUnaryBuiltin wraps a unary floating-point Go function
// as a Starlark built-in that accepts int or float arguments.
func (b mathModule) newUnaryBuiltin(name string, fn func(float64) float64) *starlark.Builtin {
	return starlark.NewBuiltin(name, core.ErrWrapper(func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if args.Len() != 1 {
			return starlark.None, fmt.Errorf("expected exactly one argument")
		}

		x, err := core.NewStarlarkValue(args.Index(0)).AsFloat64()
		if err != nil {
			return starlark.None, err
		}

		return starlark.Float(fn(x)), nil
	}))
}

// newBinaryBuiltin wraps a binary floating-point Go function
// as a Starlark built-in that accepts int or float arguments.
func (b mathModule) newBinaryBuiltin(name string, fn func(float64, float64) float64) *starlark.Builtin {
	return starlark.NewBuiltin(name, core.ErrWrapper(func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if args.Len() != 2 {
			return starlark.None, fmt.Errorf("expected exactly two arguments")
		}

		x, err := core.NewStarlarkValue(args.Index(0)).AsFloat64()
		if err != nil {
			return starlark.None, err
		}

		y, err := core.NewStarlarkValue(args.Index(1)).AsFloat64()
		if err != nil {
			return starlark.None, err
		}
		return starlark.Float(fn(x, y)), nil
	}))
}

//  log wraps the Log function
// as a Starlark built-in that accepts int or float arguments.
func (b mathModule) log(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		xValue    starlark.Value
		baseValue starlark.Value = starlark.Float(math.E)
	)
	if err := starlark.UnpackPositionalArgs("log", args, kwargs, 1, &xValue, &baseValue); err != nil {
		return nil, err
	}

	x, err := core.NewStarlarkValue(xValue).AsFloat64()
	if err != nil {
		return starlark.None, err
	}

	base, err := core.NewStarlarkValue(baseValue).AsFloat64()
	if err != nil {
		return starlark.None, err
	}

	return starlark.Float(math.Log(x) / math.Log(base)), nil
}

func (b mathModule) ceil(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value

	if err := starlark.UnpackPositionalArgs("ceil", args, kwargs, 1, &x); err != nil {
		return nil, err
	}

	switch t := x.(type) {
	case starlark.Int:
		return t, nil
	case starlark.Float:
		return starlark.NumberToInt(starlark.Float(math.Ceil(float64(t))))
	}

	return nil, fmt.Errorf("expected float value, but was %T", x)
}

func (b mathModule) floor(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value

	if err := starlark.UnpackPositionalArgs("floor", args, kwargs, 1, &x); err != nil {
		return nil, err
	}

	switch t := x.(type) {
	case starlark.Int:
		return t, nil
	case starlark.Float:
		return starlark.NumberToInt(starlark.Float(math.Floor(float64(t))))
	}

	return nil, fmt.Errorf("expected float value, but was %T", x)
}

func (b mathModule) degrees(x float64) float64 {
	return 360 * x / (2 * math.Pi)
}

func (b mathModule) radians(x float64) float64 {
	return 2 * math.Pi * x / 360
}
