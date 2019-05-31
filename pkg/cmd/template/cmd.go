package template

import (
	"fmt"
	"time"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/eval"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/spf13/cobra"
)

type TemplateOptions struct {
	IgnoreUnknownComments bool
	Debug                 bool

	BulkFilesSourceOpts    BulkFilesSourceOpts
	RegularFilesSourceOpts RegularFilesSourceOpts
	DataValuesFlags        DataValuesFlags
	Recursive              bool
}

type TemplateInput struct {
	Files []*files.File
}

type TemplateOutput struct {
	Files  []files.OutputFile
	DocSet *yamlmeta.DocumentSet
	Err    error
	Empty  bool
}

type FileSource interface {
	HasInput() bool
	HasOutput() bool
	Input() (TemplateInput, error)
	Output(TemplateOutput) error
}

var _ []FileSource = []FileSource{&BulkFilesSource{}, &RegularFilesSource{}}

func NewOptions() *TemplateOptions {
	return &TemplateOptions{
		Recursive: true,
	}
}

func NewCmd(o *TemplateOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"t", "tpl"},
		Short:   "Process YAML templates (deprecated; use top-level command -- e.g. `ytt -f-` instead of `ytt template -f-`)",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	cmd.Flags().BoolVar(&o.IgnoreUnknownComments, "ignore-unknown-comments", false,
		"Configure whether unknown comments are considered as errors (comments that do not start with '#@' or '#!')")
	cmd.Flags().BoolVar(&o.Debug, "debug", false, "Enable debug output")
	cmd.Flags().BoolVarP(&o.Recursive, "recursive", "R", true, "Template subdirectories (true by default)")
	o.BulkFilesSourceOpts.Set(cmd)
	o.RegularFilesSourceOpts.Set(cmd)
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
	if out.Empty {
		return nil
	}

	return o.pickSource(srcs, func(s FileSource) bool { return s.HasOutput() }).Output(out)
}

func (o *TemplateOptions) RunWithFiles(in TemplateInput, ui cmdcore.PlainUI) TemplateOutput {
	rootLibrary := workspace.NewRootLibrary(in.Files, o.Recursive)
	rootLibrary.Print(ui.DebugWriter())

	loaderOpts := eval.TemplateLoaderOpts{IgnoreUnknownComments: o.IgnoreUnknownComments}
	loadedLibrary, err := workspace.LoadLibrary(rootLibrary, loaderOpts)
	if err != nil {
		return TemplateOutput{Err: err}
	}

	loadedLibrary.UI = ui

	astFlagValues, err := o.DataValuesFlags.ASTValues()
	if err != nil {
		return TemplateOutput{Err: err}
	}

	filter := eval.NewValuesFilter(astFlagValues)

	loadedLibrary.Values, err = filter(loadedLibrary.Values)
	if err != nil {
		return TemplateOutput{Err: err}
	}

	if o.DataValuesFlags.Inspect {
		return o.inspectValues(loadedLibrary.Values, ui)
	}

	res, err := loadedLibrary.Eval()
	if err != nil {
		return TemplateOutput{Err: err}
	}

	return TemplateOutput{Files: res.OutputFiles, DocSet: res.DocSet}
}

func (o *TemplateOptions) pickSource(srcs []FileSource, pickFunc func(FileSource) bool) FileSource {
	for _, src := range srcs {
		if pickFunc(src) {
			return src
		}
	}
	return srcs[len(srcs)-1]
}

func (o *TemplateOptions) inspectValues(values interface{}, ui cmdcore.PlainUI) TemplateOutput {
	docSet := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{{Value: values}},
	}

	docBytes, err := docSet.AsBytes()
	if err != nil {
		return TemplateOutput{Err: fmt.Errorf("Marshaling data values: %s", err)}
	}

	ui.Printf("%s", docBytes) // no newline

	return TemplateOutput{Empty: true}
}
