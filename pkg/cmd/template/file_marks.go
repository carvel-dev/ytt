// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"regexp"
	"strings"

	"carvel.dev/ytt/pkg/files"
)

type FileMarksOpts struct {
	FileMarks []string
}

// Set registers file mark flags and wires-up those flags up to this
// FileMarksOpts to be set when the corresponding cobra.Command is executed.
func (s *FileMarksOpts) Set(cmdFlags CmdFlags) {
	cmdFlags.StringArrayVar(&s.FileMarks, "file-mark", nil, "File mark (ie change file path, mark as non-template) (format: file:key=value) (can be specified multiple times)")
}

func (s *FileMarksOpts) Apply(filesToProcess []*files.File) ([]*files.File, error) {
	var exclusiveForOutputFiles []*files.File

	for _, mark := range s.FileMarks {
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

		// Remove files that were cleared out
		filesToProcess = s.clearNils(filesToProcess)
	}

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

func (s *FileMarksOpts) fileMarkMatches(file *files.File, path string) bool {
	path = regexp.QuoteMeta(path)
	path = strings.Replace(path, quotedMultiLevel, ".+", 1)
	path = strings.Replace(path, quotedSingleLevel, "[^/]+", 1)
	return regexp.MustCompile("^" + path + "$").MatchString(file.OriginalRelativePath())
}

func (s *FileMarksOpts) clearNils(input []*files.File) []*files.File {
	var output []*files.File
	for _, file := range input {
		if file != nil {
			output = append(output, file)
		}
	}
	return output
}
