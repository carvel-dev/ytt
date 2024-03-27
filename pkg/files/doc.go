// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package files provides primitives for enumerating and loading data from various
file or file-like Source's and for writing output to filesystem files and
directories.

This allows the rest of ytt code to process logically chunked streams of data
without becoming entangled in the details of how to read or write data.

ytt processes files differently depending on their Type. For example,
File instances that are TypeYAML are parsed as YAML.
*/
package files
