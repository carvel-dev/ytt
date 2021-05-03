// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"time"

	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/spf13/cobra"
)

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

type FileSource interface {
	HasInput() bool
	HasOutput() bool
	Input() (Input, error)
	Output(Output) error
}

var _ []FileSource = []FileSource{&BulkFilesSource{}, &RegularFilesSource{}}

func NewOptions() *Options {
	return &Options{}
}

func NewCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"t", "tpl"},
		Short:   "Process YAML templates (deprecated; use top-level command -- e.g. `ytt -f-` instead of `ytt template -f-`)",
		RunE:    func(c *cobra.Command, args []string) error { return o.Run() },
	}
	cmd.Flags().BoolVar(&o.IgnoreUnknownComments, "ignore-unknown-comments", false,
		"Configure whether unknown comments are considered as errors (comments that do not start with '#@' or '#!')")
	cmd.Flags().BoolVar(&o.ImplicitMapKeyOverrides, "implicit-map-key-overrides", false,
		"Configure whether implicit map keys overrides are allowed")
	cmd.Flags().BoolVarP(&o.StrictYAML, "strict", "s", false, "Configure to use _strict_ YAML subset")
	cmd.Flags().BoolVar(&o.Debug, "debug", false, "Enable debug output")
	cmd.Flags().BoolVar(&o.InspectFiles, "files-inspect", false, "Inspect files")

	o.BulkFilesSourceOpts.Set(cmd)
	o.RegularFilesSourceOpts.Set(cmd)
	o.FileMarksOpts.Set(cmd)
	o.DataValuesFlags.Set(cmd)
	return cmd
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

	libraryExecutionFactory := workspace.NewLibraryExecutionFactory(ui, workspace.TemplateLoaderOpts{
		IgnoreUnknownComments:   o.IgnoreUnknownComments,
		ImplicitMapKeyOverrides: o.ImplicitMapKeyOverrides,
		StrictYAML:              o.StrictYAML,
	})

	libraryCtx := workspace.LibraryExecutionContext{Current: rootLibrary, Root: rootLibrary}
	rootLibraryExecution := libraryExecutionFactory.New(libraryCtx)

	schema, librarySchemas, err := rootLibraryExecution.Schemas(nil)
	if err != nil {
		return Output{Err: err}
	}

	values, libraryValues, err := rootLibraryExecution.Values(valuesOverlays, schema)
	if err != nil {
		return Output{Err: err}
	}

	libraryValues = append(libraryValues, libraryValuesOverlays...)

	if o.DataValuesFlags.Inspect {
		return Output{
			DocSet: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{values.Doc},
			},
		}
	}

	result, err := rootLibraryExecution.Eval(values, libraryValues, librarySchemas)
	if err != nil {
		return Output{Err: err}
	}

	return Output{Files: result.Files, DocSet: result.DocSet}
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
