// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package orderedmap provides a map implementation where the order of keys is
maintained (unlike the native Go map).

This flavor of map is crucial in keeping the output of ytt deterministic and
stable.
*/
package orderedmap
