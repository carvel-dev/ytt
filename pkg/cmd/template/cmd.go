package template

import (
	"fmt"
	"time"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/texttemplate"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
	"github.com/spf13/cobra"
)

type TemplateOptions struct {
	IgnoreUnknownComments bool
	Debug                 bool
	InspectFiles          bool

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
	cmd.Flags().BoolVar(&o.InspectFiles, "files-inspect", false, "Inspect files")
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
	outputDocSets := map[*workspace.FileInLibrary]*yamlmeta.DocumentSet{}

	rootLibrary := workspace.NewRootLibrary(in.Files)
	rootLibrary.Print(ui.DebugWriter())

	if o.InspectFiles {
		return o.inspectFiles(rootLibrary, ui)
	}

	forOutputFiles, valuesFiles, err := o.categorizeFiles(rootLibrary.ListAccessibleFiles())
	if err != nil {
		return TemplateOutput{Err: err}
	}

	loaderOpts := workspace.TemplateLoaderOpts{IgnoreUnknownComments: o.IgnoreUnknownComments}
	noValuesLoader := workspace.NewTemplateLoader(nil, ui, loaderOpts)

	values, err := DataValuesPreProcessing{valuesFiles, o.DataValuesFlags, noValuesLoader, o.IgnoreUnknownComments}.Apply()
	if err != nil {
		return TemplateOutput{Err: err}
	}

	if o.DataValuesFlags.Inspect {
		return o.inspectValues(values, ui)
	}

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
			outputDocSets[fileInLib] = resultDocSet

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

	for _, fileInLib := range o.sortedOutputDocSets(outputDocSets) {
		docSet := outputDocSets[fileInLib]
		combinedDocSet.Items = append(combinedDocSet.Items, docSet.Items...)

		resultDocBytes, err := docSet.AsBytes()
		if err != nil {
			return TemplateOutput{Err: fmt.Errorf("Marshaling template result: %s", err)}
		}

		ui.Debugf("### %s result\n%s", fileInLib.File.RelativePath(), resultDocBytes)
		outputFile := files.NewOutputFile(fileInLib.File.RelativePath(), resultDocBytes)
		outputFiles = append(outputFiles, outputFile)
	}

	return TemplateOutput{Files: outputFiles, DocSet: combinedDocSet}
}

func (o *TemplateOptions) categorizeFiles(allFiles []*workspace.FileInLibrary) ([]*workspace.FileInLibrary, []*workspace.FileInLibrary, error) {
	allFiles, valuesFiles, err := o.separateValuesFiles(allFiles)
	if err != nil {
		return nil, nil, err
	}

	forOutputFiles := []*workspace.FileInLibrary{}

	for _, fileInLib := range allFiles {
		if fileInLib.File.IsForOutput() {
			forOutputFiles = append(forOutputFiles, fileInLib)
		}
	}

	return forOutputFiles, valuesFiles, nil
}

func (o *TemplateOptions) separateValuesFiles(fs []*workspace.FileInLibrary) ([]*workspace.FileInLibrary, []*workspace.FileInLibrary, error) {
	var nonValuesFiles []*workspace.FileInLibrary
	var valuesFiles []*workspace.FileInLibrary

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

			valuesDocs, _, err := yttlibrary.DataValues{docSet, tplOpts}.Extract()
			if err != nil {
				return nil, nil, err
			}

			if len(valuesDocs) > 0 {
				valuesFiles = append(valuesFiles, fileInLib)
				continue
			}
		}

		nonValuesFiles = append(nonValuesFiles, fileInLib)
	}

	return nonValuesFiles, valuesFiles, nil
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

func (o *TemplateOptions) inspectFiles(rootLibrary *workspace.Library, ui cmdcore.PlainUI) TemplateOutput {
	files := rootLibrary.ListAccessibleFiles()
	workspace.SortFilesInLibrary(files)

	for _, fileInLib := range files {
		ui.Printf("%s\n", fileInLib.File.RelativePath())
	}
	return TemplateOutput{Empty: true}
}

func (o *TemplateOptions) sortedOutputDocSets(outputDocSets map[*workspace.FileInLibrary]*yamlmeta.DocumentSet) []*workspace.FileInLibrary {
	var files []*workspace.FileInLibrary
	for file, _ := range outputDocSets {
		files = append(files, file)
	}
	workspace.SortFilesInLibrary(files)
	return files
}
