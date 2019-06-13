package template

import (
	"fmt"
	"sort"
	"time"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/texttemplate"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
	"github.com/spf13/cobra"
	"go.starlark.net/starlark"
)

type TemplateOptions struct {
	IgnoreUnknownComments bool
	Debug                 bool

	BulkFilesSourceOpts    BulkFilesSourceOpts
	RegularFilesSourceOpts RegularFilesSourceOpts
	DataValuesFlags        DataValuesFlags
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
	return &TemplateOptions{}
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
	outputFiles := []files.OutputFile{}
	outputDocSets := map[string]*yamlmeta.DocumentSet{}

	rootLibrary := workspace.NewRootLibrary(in.Files)
	rootLibrary.Print(ui.DebugWriter())

	forOutputFiles, values, err := o.categorizeFiles(rootLibrary.ListAccessibleFiles())
	if err != nil {
		return TemplateOutput{Err: err}
	}

	values, err = o.overlayFlagValues(values)
	if err != nil {
		return TemplateOutput{Err: err}
	}

	if o.DataValuesFlags.Inspect {
		return o.inspectValues(values, ui)
	}

	loaderOpts := workspace.TemplateLoaderOpts{IgnoreUnknownComments: o.IgnoreUnknownComments}
	loader := workspace.NewTemplateLoader(values, ui, loaderOpts)

	for _, fileInLib := range forOutputFiles {
		// TODO find more generic way
		switch fileInLib.File.Type() {
		case files.TypeYAML:
			_, resultVal, err := loader.EvalYAML(fileInLib.Library, fileInLib.File)
			if err != nil {
				return TemplateOutput{Err: err}
			}

			resultDocSet := resultVal.(*yamlmeta.DocumentSet)
			outputDocSets[fileInLib.File.RelativePath()] = resultDocSet

		case files.TypeText:
			_, resultVal, err := loader.EvalText(fileInLib.Library, fileInLib.File)
			if err != nil {
				return TemplateOutput{Err: err}
			}

			resultStr := resultVal.(*texttemplate.NodeRoot).AsString()

			ui.Debugf("### %s result\n%s", fileInLib.File.RelativePath(), resultStr)
			outputFiles = append(outputFiles, files.NewOutputFile(fileInLib.File.RelativePath(), []byte(resultStr)))

		default:
			return TemplateOutput{Err: fmt.Errorf("Unknown file type")}
		}
	}

	outputDocSets, err = OverlayPostProcessing{outputDocSets}.Apply()
	if err != nil {
		return TemplateOutput{Err: err}
	}

	combinedDocSet := &yamlmeta.DocumentSet{}

	for _, relPath := range o.sortedOutputDocSetPaths(outputDocSets) {
		docSet := outputDocSets[relPath]
		combinedDocSet.Items = append(combinedDocSet.Items, docSet.Items...)

		resultDocBytes, err := docSet.AsBytes()
		if err != nil {
			return TemplateOutput{Err: fmt.Errorf("Marshaling template result: %s", err)}
		}

		ui.Debugf("### %s result\n%s", relPath, resultDocBytes)
		outputFiles = append(outputFiles, files.NewOutputFile(relPath, resultDocBytes))
	}

	return TemplateOutput{Files: outputFiles, DocSet: combinedDocSet}
}

func (o *TemplateOptions) categorizeFiles(allFiles []workspace.FileInLibrary) ([]workspace.FileInLibrary, interface{}, error) {
	allFiles, values, err := o.extractValues(allFiles)
	if err != nil {
		return nil, nil, err
	}

	forOutputFiles := []workspace.FileInLibrary{}

	for _, fileInLib := range allFiles {
		if fileInLib.File.IsForOutput() {
			forOutputFiles = append(forOutputFiles, fileInLib)
		}
	}

	return forOutputFiles, values, nil
}

func (o *TemplateOptions) extractValues(fs []workspace.FileInLibrary) ([]workspace.FileInLibrary, interface{}, error) {
	var foundValues interface{}
	var valuesFile *files.File
	var newFs []workspace.FileInLibrary

	for _, fileInLib := range fs {
		if fileInLib.File.Type() == files.TypeYAML && fileInLib.File.IsTemplate() {
			fileBs, err := fileInLib.File.Bytes()
			if err != nil {
				return nil, nil, err
			}

			docSet, err := yamlmeta.NewDocumentSetFromBytes(fileBs, fileInLib.File.RelativePath())
			if err != nil {
				return nil, nil, fmt.Errorf("Unmarshaling YAML template '%s': %s", fileInLib.File.RelativePath(), err)
			}

			tplOpts := yamltemplate.MetasOpts{IgnoreUnknown: o.IgnoreUnknownComments}

			values, found, err := yttlibrary.DataValues{docSet, tplOpts}.Find()
			if err != nil {
				return nil, nil, err
			}

			if found {
				if valuesFile != nil {
					// TODO until overlays are here
					return nil, nil, fmt.Errorf(
						"Template values could only be specified once, but found multiple (%s, %s)",
						valuesFile.RelativePath(), fileInLib.File.RelativePath())
				}
				valuesFile = fileInLib.File
				foundValues = values
				continue
			}
		}

		newFs = append(newFs, fileInLib)
	}

	return newFs, foundValues, nil
}

func (o *TemplateOptions) pickSource(srcs []FileSource, pickFunc func(FileSource) bool) FileSource {
	for _, src := range srcs {
		if pickFunc(src) {
			return src
		}
	}
	return srcs[len(srcs)-1]
}

func (o *TemplateOptions) sortedOutputDocSetPaths(outputDocSets map[string]*yamlmeta.DocumentSet) []string {
	var paths []string
	for relPath, _ := range outputDocSets {
		paths = append(paths, relPath)
	}
	sort.Strings(paths)
	return paths
}

func (o *TemplateOptions) overlayFlagValues(fileValues interface{}) (interface{}, error) {
	if fileValues == nil {
		fileValues = map[interface{}]interface{}{}
	}

	astFlagValues, err := o.DataValuesFlags.ASTValues()
	if err != nil {
		return nil, err
	}

	op := yttoverlay.OverlayOp{
		Left:   yamlmeta.NewASTFromInterface(fileValues),
		Right:  astFlagValues,
		Thread: &starlark.Thread{Name: "data-values-overlay-pre-processing"},
	}

	newLeft, err := op.Apply()
	if err != nil {
		return nil, fmt.Errorf("Overlaying data values from flags (provided via --data-value-*) on top of data values (marked as @data/values): %s", err)
	}

	return (&yamlmeta.Document{Value: newLeft}).AsInterface(yamlmeta.InterfaceConvertOpts{}), nil
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
