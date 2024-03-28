// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package filepos provides the concept of Position: a source name (usually a file)
and line number within that source.

File positions are crucial when reporting errors to the user. It is often
even more useful to share the actual source line as well. For this reason
Position also contains a cached copy of the source line at the Position.

Not all Position point within a file (e.g. code that is generated). The
zero-value of Position (can be created using NewUnknownPosition()) represents
this case.
*/
package filepos
