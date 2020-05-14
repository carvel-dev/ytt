package template

import (
	"time"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/spf13/cobra"
)

type TemplateOptions struct {
	IgnoreUnknownComments bool
	StrictYAML            bool
	Debug                 bool
	InspectFiles          bool

	BulkFilesSourceOpts    BulkFilesSourceOpts
	RegularFilesSourceOpts RegularFilesSourceOpts
	FileMarksOpts          FileMarksOpts
	DataValuesFlags        DataValuesFlags
}

type TemplateInput struct {
	Files []*files.File
}

type TemplateOutput struct {
	Files  []files.OutputFile
	DocSet *yamlmeta.DocumentSet
	Err    error
}

type FileSource interface {
	HasInput() bool
	HasOutput() bool
	Input() (TemplateInput, error)
	Output(TemplateOutput) error
}

var _ []FileSource = []FileSource{&BulkFilesSource{}, &RegularFilesSource{}}

func NewOptions() *TemplateOptions {
	return &TemplateOptions{}
}

func NewCmd(o *TemplateOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"t", "tpl"},
		Short:   "Process YAML templates (deprecated; use top-level command -- e.g. `ytt -f-` instead of `ytt template -f-`)",
		RunE:    func(c *cobra.Command, args []string) error { return o.Run() },
	}
	cmd.Flags().BoolVar(&o.IgnoreUnknownComments, "ignore-unknown-comments", false,
		"Configure whether unknown comments are considered as errors (comments that do not start with '#@' or '#!')")
	cmd.Flags().BoolVarP(&o.StrictYAML, "strict", "s", false, "Configure to use _strict_ YAML subset")
	cmd.Flags().BoolVar(&o.Debug, "debug", false, "Enable debug output")
	cmd.Flags().BoolVar(&o.InspectFiles, "files-inspect", false, "Inspect files")
	o.BulkFilesSourceOpts.Set(cmd)
	o.RegularFilesSourceOpts.Set(cmd)
	o.FileMarksOpts.Set(cmd)
	o.DataValuesFlags.Set(cmd)
	return cmd
}

func (o *TemplateOptions) Run() error {
	ui := cmdcore.NewPlainUI(o.Debug)
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

func (o *TemplateOptions) RunWithFiles(in TemplateInput, ui cmdcore.PlainUI) TemplateOutput {
	var err error

	in.Files, err = o.FileMarksOpts.Apply(in.Files)
	if err != nil {
		return TemplateOutput{Err: err}
	}

	rootLibrary := workspace.NewRootLibrary(in.Files)
	rootLibrary.Print(ui.DebugWriter())

	if o.InspectFiles {
		return o.inspectFiles(rootLibrary, ui)
	}

	valuesOverlays, libraryValuesOverlays, err := o.DataValuesFlags.AsOverlays(o.StrictYAML)
	if err != nil {
		return TemplateOutput{Err: err}
	}

	libraryExecutionFactory := workspace.NewLibraryExecutionFactory(ui, workspace.TemplateLoaderOpts{
		IgnoreUnknownComments: o.IgnoreUnknownComments,
		StrictYAML:            o.StrictYAML,
	})

	libraryCtx := workspace.LibraryExecutionContext{Current: rootLibrary, Root: rootLibrary}
	libraryLoader := libraryExecutionFactory.New(libraryCtx)

	values, libraryValues, err := libraryLoader.Values(valuesOverlays)
	if err != nil {
		return TemplateOutput{Err: err}
	}
	libraryValues = append(libraryValues, libraryValuesOverlays...)

	if o.DataValuesFlags.Inspect {
		return TemplateOutput{
			DocSet: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{values.Doc},
			},
		}
	}

	result, err := libraryLoader.Eval(values, libraryValues)
	if err != nil {
		return TemplateOutput{Err: err}
	}

	return TemplateOutput{Files: result.Files, DocSet: result.DocSet}
}

func (o *TemplateOptions) pickSource(srcs []FileSource, pickFunc func(FileSource) bool) FileSource {
	for _, src := range srcs {
		if pickFunc(src) {
			return src
		}
	}
	return srcs[len(srcs)-1]
}

func (o *TemplateOptions) inspectFiles(rootLibrary *workspace.Library, ui cmdcore.PlainUI) TemplateOutput {
	files := rootLibrary.ListAccessibleFiles()
	workspace.SortFilesInLibrary(files)

	paths := &yamlmeta.Array{}

	for _, fileInLib := range files {
		paths.Items = append(paths.Items, &yamlmeta.ArrayItem{
			Value: fileInLib.File.RelativePath(),
		})
	}

	return TemplateOutput{
		DocSet: &yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{{Value: paths}},
		},
	}
}
