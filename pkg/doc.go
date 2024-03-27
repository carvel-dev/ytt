// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package pkg is the collection of packages that make up the implementation of ytt.

This codebase is intentionally organized into well-defined layers. A concerted
effort has been sustained to keep the responsibility of each package concise and
complete. Packages have been designed to be dependent on each other only to the
degree absolutely required.

In the inventory, below, individual packages are named alongside their coupling
with the other packages in the codebase.

	(# of dependents) => <package name> => (# of dependencies)

Where "# of dependents" is the count of packages that import the named package
and "# of dependencies" is the count of packages that this named package
imports.

This is output from the tool https://github.com/jtigger/go-orient.

From top-down (http://www.catb.org/~esr/writings/taoup/html/ch04s03.html), ytt
code is layered in this way:

# Entry Point

ytt is built into two executable formats:

	./cmd/ytt                  // a command-line tool
	./cmd/ytt-lambda-website   // an AWS Lambda function

ytt's contains a mini website that is the "Playground".

	(1) => pkg/website => (0)

# Commands

There are half-a-dozen commands ytt implements. The most commonly used and most
complex is "template".

	(1) => pkg/cmd => (8)
	(2) => pkg/cmd/template => (10)

# The Workspace

ytt processing can be thought of as a sequence of steps: schema "pre-processing",
data values "pre-processing", template evaluation, and overlay "post-processing".
All of these steps are executed over a collection of files we call a
workspace.Library.

	(1) => pkg/workspace => (14)
	(2) => pkg/workspace/datavalues => (6)
	(3) => pkg/workspace/ref => (2)
	(4) => pkg/files => (1)
	(10) => pkg/filepos => (0)

# Templating

The heart of ytt's action is structural templating. There are two flavors of
templates: templating on a YAML structure and templating on text. Each source
template file is "compiled" into a Starlark program whose job is to build-up
the output.

	(9) => pkg/template => (2)
	(9) => pkg/template/core => (1)
	(3) => pkg/yamltemplate => (6)
	(2) => pkg/texttemplate => (2)

# Standard Library

ytt injects a collection of "standard" modules into Starlark executions.
These modules support serialization, hashing, parsing of special kinds
of values, and programmatic access to some of ytt's core capabilities. This
collection is known as a library (not to be confused with the workspace.Library
concept).

	(2) => pkg/yttlibrary => (5)

The implementation for ytt's other fundamental feature -- Overlays -- lives
within this standard library. This allows for both template authors and ytt
itself to use this powerful editing feature.

	(3) => pkg/yttlibrary/overlay => (5)

"Standard Library" modules that have Go dependencies are included as
"extensions" to minimize the set of transitive dependencies that come from the
carvel.dev/ytt module.

	(1) => pkg/yttlibraryext => (1)
	(1) => pkg/yttlibraryext/toml => (4)

# YAML Enhancements

ytt can enrich a YAML structure with more fine-grained data typing and
user-defined constraints on values.

	(3) => pkg/schema => (5)
	(1) => pkg/assertions => (5)

# YAML Structures

At the heart of ytt is the ability to parse annotated YAML.

ytt delegates parsing YAML to the de facto standard YAML library
(https://github.com/go-yaml/yaml/tree/v2). However, ytt needs to store
additional information about nodes (e.g. templating, type info, processing
hints, etc.). It does this by converting the output from the standard YAML
parser into a composite tree of its own yamlmeta.Node structure that can
hold that metadata.

	(11) => pkg/yamlmeta => (3)
	(1) => pkg/yamlmeta/internal/yaml.v2 => (0)

# Utilities

Finally, there is a collection of supporting features.

One of which provides the implementation of the "fmt" command:

	(1) => pkg/yamlfmt => (1)

The remainder are domain-agnostic utilities that provide either an
application-level capability or a specialized piece of logic.

	(5) => pkg/cmd/ui => (0)
	(5) => pkg/orderedmap => (0)
	(2) => pkg/version => (0)
	(2) => pkg/feature => (0)
	(1) => pkg/spell => (0)

# Dependencies

Each package's dependencies on other packages within this module are as follows
(if a package is not listed, it has no dependencies on other packages within
this module):

	pkg/cmd/template:
	- pkg/workspace
	- pkg/workspace/datavalues
	- pkg/schema
	- pkg/workspace/ref
	- pkg/yttlibrary/overlay
	- pkg/files
	- pkg/cmd/ui
	- pkg/template
	- pkg/filepos
	- pkg/yamlmeta
	pkg/workspace:
	- pkg/assertions
	- pkg/texttemplate
	- pkg/workspace/datavalues
	- pkg/yttlibrary
	- pkg/schema
	- pkg/workspace/ref
	- pkg/yamltemplate
	- pkg/yttlibrary/overlay
	- pkg/files
	- pkg/cmd/ui
	- pkg/template
	- pkg/template/core
	- pkg/filepos
	- pkg/yamlmeta
	pkg/assertions:
	- pkg/feature
	- pkg/yamltemplate
	- pkg/template
	- pkg/filepos
	- pkg/yamlmeta
	pkg/yamltemplate:
	- pkg/texttemplate
	- pkg/orderedmap
	- pkg/template
	- pkg/template/core
	- pkg/filepos
	- pkg/yamlmeta
	pkg/texttemplate:
	- pkg/template
	- pkg/filepos
	pkg/template:
	- pkg/template/core
	- pkg/filepos
	pkg/template/core:
	- pkg/orderedmap
	pkg/yamlmeta:
	- pkg/yamlmeta/internal/yaml.v2
	- pkg/orderedmap
	- pkg/filepos
	pkg/workspace/datavalues:
	- pkg/schema
	- pkg/workspace/ref
	- pkg/template
	- pkg/template/core
	- pkg/filepos
	- pkg/yamlmeta
	pkg/schema:
	- pkg/spell
	- pkg/template
	- pkg/template/core
	- pkg/filepos
	- pkg/yamlmeta
	pkg/workspace/ref:
	- pkg/template
	- pkg/template/core
	pkg/yttlibrary:
	- pkg/version
	- pkg/yttlibrary/overlay
	- pkg/orderedmap
	- pkg/template/core
	- pkg/yamlmeta
	pkg/yttlibrary/overlay:
	- pkg/yamltemplate
	- pkg/template
	- pkg/template/core
	- pkg/filepos
	- pkg/yamlmeta
	pkg/files:
	- pkg/cmd/ui
	pkg/cmd:
	- pkg/website
	- pkg/yamlfmt
	- pkg/yttlibraryext
	- pkg/cmd/template
	- pkg/version
	- pkg/files
	- pkg/cmd/ui
	- pkg/yamlmeta
	pkg/yamlfmt:
	- pkg/yamlmeta
	pkg/yttlibraryext:
	- pkg/yttlibraryext/toml
	pkg/yttlibraryext/toml:
	- pkg/yttlibrary
	- pkg/orderedmap
	- pkg/template/core
	- pkg/yamlmeta
*/
package pkg
