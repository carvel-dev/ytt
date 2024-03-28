// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"time"

	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
	"carvel.dev/ytt/pkg/schema"
	"carvel.dev/ytt/pkg/workspace"
	"carvel.dev/ytt/pkg/workspace/datavalues"
	"carvel.dev/ytt/pkg/yamlmeta"
)

// Options both contains the configuration for a "template" command AND the
// top-level implementation of that command.
//
// For most intents and purposes, RunWithFiles() is the entrypoint.
type Options struct {
	IgnoreUnknownComments   bool
	ImplicitMapKeyOverrides bool

	StrictYAML   bool
	Debug        bool
	InspectFiles bool

	BulkFilesSourceOpts    BulkFilesSourceOpts
	RegularFilesSourceOpts RegularFilesSourceOpts
	FileMarksOpts          FileMarksOpts
	DataValuesFlags        DataValuesFlags
}

type Input struct {
	Files []*files.File
}

type Output struct {
	Files  []files.OutputFile
	DocSet *yamlmeta.DocumentSet
	Err    error
}

// FileSource provides both a means of loading from sources (i.e. Input) and rendering into sinks (i.e. Output)
type FileSource interface {
	HasInput() bool
	HasOutput() bool
	Input() (Input, error)
	// Output renders the results (i.e. an instance of Output) to the configured output/sink of this FileSource
	Output(Output) error
}

var _ []FileSource = []FileSource{&BulkFilesSource{}, &RegularFilesSource{}}

// NewOptions initializes a new instance of template.Options.
func NewOptions() *Options {
	var opts files.SymlinkAllowOpts
	return &Options{
		RegularFilesSourceOpts: RegularFilesSourceOpts{SymlinkAllowOpts: &opts},
		DataValuesFlags:        DataValuesFlags{SymlinkAllowOpts: &opts},
	}
}

// BindFlags registers template flags for template command.
func (o *Options) BindFlags(cmdFlags CmdFlags) {
	cmdFlags.BoolVar(&o.IgnoreUnknownComments, "ignore-unknown-comments", false,
		"Configure whether unknown comments are considered as errors (comments that do not start with '#@' or '#!')")
	cmdFlags.BoolVar(&o.ImplicitMapKeyOverrides, "implicit-map-key-overrides", false,
		"Configure whether implicit map keys overrides are allowed")
	cmdFlags.BoolVarP(&o.StrictYAML, "strict", "s", false, "Configure to use _strict_ YAML subset")
	cmdFlags.BoolVar(&o.Debug, "debug", false, "Enable debug output")
	cmdFlags.BoolVar(&o.InspectFiles, "files-inspect", false, "Determine the set of files that would be processed and display that result")

	o.BulkFilesSourceOpts.Set(cmdFlags)
	o.RegularFilesSourceOpts.Set(cmdFlags)
	o.FileMarksOpts.Set(cmdFlags)
	o.DataValuesFlags.Set(cmdFlags)
}

func (o *Options) Run() error {
	ui := ui.NewTTY(o.Debug)
	t1 := time.Now()

	defer func() {
		ui.Debugf("total: %s\n", time.Now().Sub(t1))
	}()

	srcs := []FileSource{
		NewBulkFilesSource(o.BulkFilesSourceOpts, ui),
		NewRegularFilesSource(o.RegularFilesSourceOpts, ui),
	}

	in, err := o.pickSource(srcs, func(s FileSource) bool { return s.HasInput() }).Input()
	if err != nil {
		return err
	}

	out := o.RunWithFiles(in, ui)
	return o.pickSource(srcs, func(s FileSource) bool { return s.HasOutput() }).Output(out)
}

func (o *Options) RunWithFiles(in Input, ui ui.UI) Output {
	var err error

	in.Files, err = o.FileMarksOpts.Apply(in.Files)
	if err != nil {
		return Output{Err: err}
	}

	rootLibrary := workspace.NewRootLibrary(in.Files)
	rootLibrary.Print(ui.DebugWriter())

	if o.InspectFiles {
		return o.inspectFiles(rootLibrary)
	}

	valuesOverlays, libraryValuesOverlays, err := o.DataValuesFlags.AsOverlays(o.StrictYAML)
	if err != nil {
		return Output{Err: err}
	}

	libraryExecutionFactory := workspace.NewLibraryExecutionFactory(
		ui,
		workspace.TemplateLoaderOpts{
			IgnoreUnknownComments:   o.IgnoreUnknownComments,
			ImplicitMapKeyOverrides: o.ImplicitMapKeyOverrides,
			StrictYAML:              o.StrictYAML,
		},
		o.DataValuesFlags.SkipValidation)

	libraryCtx := workspace.LibraryExecutionContext{Current: rootLibrary, Root: rootLibrary}
	rootLibraryExecution := libraryExecutionFactory.New(libraryCtx)

	schema, librarySchemas, err := rootLibraryExecution.Schemas(nil)
	if err != nil {
		return Output{Err: err}
	}

	if o.DataValuesFlags.InspectSchema {
		return o.inspectSchema(schema)
	}

	schemaType, err := o.RegularFilesSourceOpts.OutputType.Schema()
	if err != nil {
		return Output{Err: err}
	}
	if schemaType == RegularFilesOutputTypeOpenAPI {
		return Output{Err: fmt.Errorf("Output type currently only supported for data values schema (i.e. include --data-values-schema-inspect)")}
	}

	values, libraryValues, err := rootLibraryExecution.Values(valuesOverlays, schema)
	if err != nil {
		return Output{Err: err}
	}

	libraryValues = append(libraryValues, libraryValuesOverlays...)

	if o.DataValuesFlags.Inspect {
		return o.inspectDataValues(values)
	}

	result, err := rootLibraryExecution.Eval(values, libraryValues, librarySchemas)
	if err != nil {
		return Output{Err: err}
	}

	return Output{Files: result.Files, DocSet: result.DocSet}
}

func (o *Options) inspectDataValues(values *datavalues.Envelope) Output {
	return Output{
		DocSet: &yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{values.Doc},
		},
	}
}

func (o *Options) inspectSchema(dataValuesSchema *datavalues.Schema) Output {
	format, err := o.RegularFilesSourceOpts.OutputType.Schema()
	if err != nil {
		return Output{Err: err}
	}
	if format == RegularFilesOutputTypeOpenAPI {
		openAPIDoc := schema.NewOpenAPIDocument(dataValuesSchema.GetDocumentType())
		return Output{
			DocSet: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{openAPIDoc.AsDocument()},
			},
		}
	}
	return Output{Err: fmt.Errorf("Data values schema export only supported in OpenAPI v3 format; specify format with --output=%s flag",
		RegularFilesOutputTypeOpenAPI)}
}

func (o *Options) pickSource(srcs []FileSource, pickFunc func(FileSource) bool) FileSource {
	for _, src := range srcs {
		if pickFunc(src) {
			return src
		}
	}
	return srcs[len(srcs)-1]
}

func (o *Options) inspectFiles(rootLibrary *workspace.Library) Output {
	accessibleFiles := rootLibrary.ListAccessibleFiles()
	workspace.SortFilesInLibrary(accessibleFiles)

	paths := &yamlmeta.Array{}

	for _, fileInLib := range accessibleFiles {
		paths.Items = append(paths.Items, &yamlmeta.ArrayItem{
			Value: fileInLib.File.RelativePath(),
		})
	}

	return Output{
		DocSet: &yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{{Value: paths}},
		},
	}
}
