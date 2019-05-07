package template

import (
	"fmt"
	"regexp"
	"strings"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"
	"github.com/spf13/cobra"
)

type RegularFilesSourceOpts struct {
	files     []string
	fileMarks []string
	recursive bool
	output    string
}

func (s *RegularFilesSourceOpts) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&s.files, "file", "f", nil, "File (ie local path, HTTP URL, -) (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&s.fileMarks, "file-mark", nil, "File mark (ie change file path, mark as non-template) (format: file:key=value) (can be specified multiple times)")
	cmd.Flags().BoolVarP(&s.recursive, "recursive", "R", false, "Interpret file as directory")
	cmd.Flags().StringVarP(&s.output, "output", "o", "", "Directory for output")
}

type RegularFilesSource struct {
	opts RegularFilesSourceOpts
	ui   cmdcore.PlainUI
}

func NewRegularFilesSource(opts RegularFilesSourceOpts, ui cmdcore.PlainUI) *RegularFilesSource {
	return &RegularFilesSource{opts, ui}
}

func (s *RegularFilesSource) HasInput() bool  { return len(s.opts.files) > 0 }
func (s *RegularFilesSource) HasOutput() bool { return true }

func (s *RegularFilesSource) Input() (TemplateInput, error) {
	filesToProcess, err := files.NewFiles(s.opts.files, s.opts.recursive)
	if err != nil {
		return TemplateInput{}, err
	}

	filesToProcess, err = s.applyFileMarks(filesToProcess)
	if err != nil {
		return TemplateInput{}, err
	}

	return TemplateInput{Files: filesToProcess}, nil
}

func (s *RegularFilesSource) Output(out TemplateOutput) error {
	if out.Err != nil {
		return out.Err
	}

	if len(s.opts.output) > 0 {
		return files.NewOutputDirectory(s.opts.output, out.Files, s.ui).Write()
	}

	combinedDocBytes, err := out.DocSet.AsBytes()
	if err != nil {
		return fmt.Errorf("Marshaling combined template result: %s", err)
	}

	s.ui.Debugf("### result\n")
	s.ui.Printf("%s", combinedDocBytes) // no newline

	return nil
}

func (s *RegularFilesSource) applyFileMarks(filesToProcess []*files.File) ([]*files.File, error) {
	var exclusiveForOutputFiles []*files.File

	for _, mark := range s.opts.fileMarks {
		pieces := strings.SplitN(mark, ":", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("Expected file mark '%s' to be in format path:key=value", mark)
		}

		path := pieces[0]

		kv := strings.SplitN(pieces[1], "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("Expected file mark '%s' key-value portion to be in format key=value", mark)
		}

		var matched bool

		for i, file := range filesToProcess {
			if s.fileMarkMatches(file, path) {
				matched = true

				switch kv[0] {
				case "path":
					file.MarkRelativePath(kv[1])

				case "exclude":
					switch kv[1] {
					case "true":
						filesToProcess[i] = nil
					default:
						return nil, fmt.Errorf("Unknown value in file mark '%s'", mark)
					}

				case "type":
					switch kv[1] {
					case "yaml-template": // yaml template processing
						file.MarkType(files.TypeYAML)
						file.MarkTemplate(true)
					case "yaml-plain": // no template processing
						file.MarkType(files.TypeYAML)
						file.MarkTemplate(false)
					case "text-template":
						file.MarkType(files.TypeText)
						file.MarkTemplate(true)
					case "text-plain":
						file.MarkType(files.TypeText)
						file.MarkTemplate(false)
					case "starlark":
						file.MarkType(files.TypeStarlark)
						file.MarkTemplate(false)
					case "data":
						file.MarkType(files.TypeUnknown)
						file.MarkTemplate(false)
					default:
						return nil, fmt.Errorf("Unknown value in file mark '%s'", mark)
					}

				case "for-output":
					switch kv[1] {
					case "true":
						file.MarkForOutput(true)
					default:
						return nil, fmt.Errorf("Unknown value in file mark '%s'", mark)
					}

				case "exclusive-for-output":
					switch kv[1] {
					case "true":
						exclusiveForOutputFiles = append(exclusiveForOutputFiles, file)
					default:
						return nil, fmt.Errorf("Unknown value in file mark '%s'", mark)
					}

				default:
					return nil, fmt.Errorf("Unknown key in file mark '%s'", mark)
				}
			}
		}

		if !matched {
			return nil, fmt.Errorf("Expected file mark '%s' to match at least one file by path, but did not", mark)
		}
	}

	// Remove files that were cleared out
	filesToProcess = s.clearNils(filesToProcess)

	// If there is at least filtered output file, mark all others as non-templates
	if len(exclusiveForOutputFiles) > 0 {
		for _, file := range filesToProcess {
			file.MarkForOutput(false)
		}
		for _, file := range exclusiveForOutputFiles {
			file.MarkForOutput(true)
		}
	}

	return filesToProcess, nil
}

var (
	quotedMultiLevel  = regexp.QuoteMeta("**/*")
	quotedSingleLevel = regexp.QuoteMeta("*")
)

func (s *RegularFilesSource) fileMarkMatches(file *files.File, path string) bool {
	path = regexp.QuoteMeta(path)
	path = strings.Replace(path, quotedMultiLevel, ".+", 1)
	path = strings.Replace(path, quotedSingleLevel, "[^/]+", 1)
	return regexp.MustCompile("^" + path + "$").MatchString(file.OriginalRelativePath())
}

func (s *RegularFilesSource) clearNils(input []*files.File) []*files.File {
	var output []*files.File
	for _, file := range input {
		if file != nil {
			output = append(output, file)
		}
	}
	return output
}
