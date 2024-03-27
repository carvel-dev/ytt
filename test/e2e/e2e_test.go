// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDataValues(t *testing.T) {
	t.Run("can be set through command-line flags", func(t *testing.T) {
		t.Run("--data-value flag", func(t *testing.T) {
			flags := yttFlags{
				{"--data-value": "nothing=null"},
				{"--data-value": "string=str"},
				{"--data-value": "bool=true"},
				{"--data-value": "int=123"},
				{"--data-value": "float=123.123"},
			}
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values/config.yml", "../../examples/data-values/values.yml"}, "", flags, nil)
			expectedOutput := `nothing: "null"
string: str
bool: "true"
int: "123"
float: "123.123"
`

			require.Equal(t, expectedOutput, actualOutput)
		})
		t.Run("--data-value-yaml flag", func(t *testing.T) {
			flags := yttFlags{
				{"--data-value-yaml": "nothing=null"},
				{"--data-value-yaml": "string=str"},
				{"--data-value-yaml": "bool=true"},
				{"--data-value-yaml": "int=123"},
				{"--data-value-yaml": "float=123.123"},
			}
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values/config.yml", "../../examples/data-values/values.yml"}, "", flags, nil)
			expectedOutput := `nothing: null
string: str
bool: true
int: 123
float: 123.123
`

			require.Equal(t, expectedOutput, actualOutput)
		})
		t.Run("--data-values-env flag", func(t *testing.T) {
			flags := yttFlags{
				{"--data-values-env": "STR_VAL"},
				{"--data-values-env-yaml": "YAML_VAL"},
			}
			envs := []string{
				"STR_VAL_nothing=null",
				"YAML_VAL_string=str",
				"YAML_VAL_bool=true",
				"YAML_VAL_int=123",
				"YAML_VAL_float=123.123",
			}
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values/config.yml", "../../examples/data-values/values.yml"}, "", flags, envs)
			expectedOutput := `nothing: "null"
string: str
bool: true
int: 123
float: 123.123
`

			require.Equal(t, expectedOutput, actualOutput)
		})
		t.Run("--data-value-yaml && --data-values-env-yaml flag", func(t *testing.T) {
			flags := yttFlags{
				{"--data-value-yaml": "nothing=[1,2,3]"},
				{"--data-values-env-yaml": "YAML_VAL"},
			}
			envs := []string{
				"YAML_VAL_string=[1,2,4]",
				"YAML_VAL_bool=true",
				"YAML_VAL_int=123",
				"YAML_VAL_float=123.123",
			}
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values/config.yml", "../../examples/data-values/values.yml"}, "", flags, envs)
			expectedOutput := `nothing:
- 1
- 2
- 3
string:
- 1
- 2
- 4
bool: true
int: 123
float: 123.123
`

			require.Equal(t, expectedOutput, actualOutput)
		})
		t.Run("--data-values-file enumerates YAML files from given directory", func(t *testing.T) {
			testDir := "../../examples/data-values-directory"
			flags := yttFlags{
				{"--data-values-file": testDir + "/values"},
			}
			actualOutput := runYtt(t, testInputFiles{testDir + "/config"}, "", flags, nil)

			expectedOutput, err := os.ReadFile(filepath.Join(testDir, "expected.txt"))
			require.NoError(t, err)

			require.Equal(t, string(expectedOutput), actualOutput)
		})
		t.Run("--data-value-file flag", func(t *testing.T) {
			flags := yttFlags{
				{"--data-value-file": "string=../../examples/data-values/file-as-value.txt"},
			}
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values/config.yml", "../../examples/data-values/values.yml"}, "", flags, nil)
			expectedOutput := `nothing: something
string: |-
  Value that comes from a file.

  Typically, useful for files that contain values
  that should just be passed through to system
  configuration, wholesale (e.g. certs).
bool: false
int: 0
float: 0
`

			require.Equal(t, expectedOutput, actualOutput)
		})

	})
	t.Run("can be 'required'", func(t *testing.T) {
		expectedFileOutput, err := os.ReadFile("../../examples/data-values-required/expected.txt")
		require.NoError(t, err)

		dirs := []string{
			"data-values-required/inline",
			"data-values-required/function",
			"data-values-required/bulk",
		}
		for _, dir := range dirs {
			t.Run(dir, func(t *testing.T) {
				dirPath := fmt.Sprintf("../../examples/%s", dir)
				flags := yttFlags{
					{"-v": "version=123"},
				}
				actualOutput := runYtt(t, testInputFiles{dirPath}, "", flags, nil)

				require.Equal(t, string(expectedFileOutput), actualOutput)
			})
		}
	})
	t.Run("example supporting multi-environment scenarios", func(t *testing.T) {
		t.Run("with defaults", func(t *testing.T) {
			flags := yttFlags{
				{"-v": "version=123"},
			}
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-multiple-envs/config/"}, "", flags, nil)

			expectedOutput := `app_config:
  version: "123"
  ports:
  - 8080
`
			require.Equal(t, expectedOutput, actualOutput)
		})
		t.Run("with 'dev' overrides", func(t *testing.T) {
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-multiple-envs/config/", "../../examples/data-values-multiple-envs/envs/dev.yml"}, "", nil, nil)
			expectedOutput := `app_config:
  version: v1alpha1
  ports:
  - 8080
`
			require.Equal(t, expectedOutput, actualOutput)
		})
		t.Run("with 'staging' overrides", func(t *testing.T) {
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-multiple-envs/config/", "../../examples/data-values-multiple-envs/envs/staging.yml"}, "", nil, nil)
			expectedOutput := `app_config:
  version: v1beta1
  ports:
  - 8081
`
			require.Equal(t, expectedOutput, actualOutput)
		})
		t.Run("with 'prod' overrides", func(t *testing.T) {
			actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-multiple-envs/config/", "../../examples/data-values-multiple-envs/envs/prod.yml"}, "", nil, nil)
			expectedOutput := `app_config:
  version: v1
  ports:
  - 80
`
			require.Equal(t, expectedOutput, actualOutput)
		})
	})
}

func TestSchema(t *testing.T) {
	dirs := []string{
		"schema",
		"schema-arrays",
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			dirPath := fmt.Sprintf("../../examples/%s", dir)
			actualOutput := runYtt(t, testInputFiles{dirPath}, "", nil, nil)

			expectedOutput, err := os.ReadFile(filepath.Join(dirPath, "expected.txt"))
			require.NoError(t, err)

			require.Equal(t, string(expectedOutput), actualOutput)
		})
	}
}

func TestOverlays(t *testing.T) {
	dirs := []string{
		"overlay",
		"overlay-files",
		"overlay-not-matcher",
		"overlay-regular-files",
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			dirPath := fmt.Sprintf("../../examples/%s", dir)
			actualOutput := runYtt(t, testInputFiles{dirPath}, "", nil, nil)

			expectedOutput, err := os.ReadFile(filepath.Join(dirPath, "expected.txt"))
			require.NoError(t, err)

			require.Equal(t, string(expectedOutput), actualOutput)
		})
	}
}

func TestVersionIsValid(t *testing.T) {
	runYtt(t, testInputFiles{"../../examples/version-constraint"}, "", yttFlags{}, nil)
}

func TestPipes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test as it applies to linux based OS")
	}

	t.Run("test pipe stdin", func(t *testing.T) {
		err := exec.Command("./assets/test_pipe_stdin.sh").Run()
		require.NoError(t, err)
	})

	t.Run("test pipe redirect", func(t *testing.T) {
		err := exec.Command("./assets/test_pipe_redirect.sh").Run()
		require.NoError(t, err)
	})
}

func TestReadingFromStandardIn(t *testing.T) {
	t.Run("through --file", func(t *testing.T) {
		actualOutput := runYtt(t, []string{"-", "../../examples/eirini/input.yml"}, "../../examples/eirini/config.yml", nil, nil)

		expectedFileOutput, err := os.ReadFile("../../examples/eirini/config-result.yml")
		require.NoError(t, err)
		require.Equal(t, string(expectedFileOutput), actualOutput)
	})
	t.Run("through --data-values-file", func(t *testing.T) {
		flags := yttFlags{
			{"--data-values-file": "-"},
			{"--data-values-inspect": ""},
		}
		actualOutput := runYtt(t, []string{}, "../../examples/data-values/values-file.yml", flags, nil)
		expectedOutput := `nothing: something
string: str
bool: true
int: 124
new_thing: new
`
		require.Equal(t, expectedOutput, actualOutput)
	})
	t.Run("can only be read once", func(t *testing.T) {
		flags := yttFlags{
			{"--data-values-file": "-"},
			{"-f": "-"},
		}
		actualOutput := runYttExpectingError(t, nil, flags, nil)
		expectedOutput := "ytt: Error: Extracting data value from file:\n  Reading file 'stdin.yml':\n    Standard input has already been read, has the '-' argument been used in more than one flag?\n"
		require.Equal(t, expectedOutput, actualOutput)
	})
}

func TestCheckDirectoryOutput(t *testing.T) {
	tempOutputDir, err := os.MkdirTemp(os.TempDir(), "ytt-check-dir")
	require.NoError(t, err)
	defer os.Remove(tempOutputDir)
	flags := yttFlags{{fmt.Sprintf("--dangerous-emptied-output-directory=%s", tempOutputDir): ""}}
	runYtt(t, []string{"../../examples/eirini/"}, "", flags, nil)

	expectedFileOutput, err := os.ReadFile("../../examples/eirini/config-result.yml")
	require.NoError(t, err)

	actualOutput, err := os.ReadFile(filepath.Join(tempOutputDir, "config-result.yml"))
	require.NoError(t, err)

	require.Equal(t, string(expectedFileOutput), string(actualOutput))
}

func TestJsonOutput(t *testing.T) {
	actualOutput := runYtt(t, testInputFiles{"../../examples/k8s-adjust-rbac-version"}, "", yttFlags{{"-o": "json"}}, nil)
	expectedOutput := `{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"ClusterRole"}{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"ClusterRole"}`

	require.Equal(t, expectedOutput, actualOutput)
}

func TestPlaygroundExamplesExecuteWithoutError(t *testing.T) {
	t.Run("Basics", func(t *testing.T) {
		filepath.WalkDir("../../examples/playground/basics", func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				return filepath.SkipDir
			}

			switch d.Name() {
			case "example-assert", "example-load-custom-library-module", "example-ytt-library-module":
				return filepath.SkipDir
			case "basics":
				return nil
			default:
				t.Run(fmt.Sprintf("playground %s", d.Name()), func(t *testing.T) {
					runYtt(t, testInputFiles{path}, "", nil, nil)
				})
				return nil
			}
		})
	})
	t.Run("Overlays", func(t *testing.T) {
		filepath.WalkDir("../../examples/playground/overlays", func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				return filepath.SkipDir
			}

			if d.Name() == "overlays" {
				return nil
			}
			t.Run(fmt.Sprintf("playground %s", d.Name()), func(t *testing.T) {
				runYtt(t, testInputFiles{path}, "", nil, nil)
			})
			return nil
		})
	})
}

func TestApplicationSpecificScenarios(t *testing.T) {
	t.Run("Kubernetes", func(t *testing.T) {
		dirs := []string{
			"k8s-add-global-label",
			"k8s-adjust-rbac-version",
			"k8s-docker-secret",
			"k8s-relative-rolling-update",
			"k8s-config-map-files",
			"k8s-update-env-var",
			"k8s-overlay-all-containers",
			"k8s-overlay-remove-resources",
			"k8s-overlay-in-config-map",
		}
		for _, dir := range dirs {
			t.Run(dir, func(t *testing.T) {
				dirPath := fmt.Sprintf("../../examples/%s", dir)
				actualOutput := runYtt(t, testInputFiles{dirPath}, "", nil, nil)

				expectedOutput, err := os.ReadFile(filepath.Join(dirPath, "expected.txt"))
				require.NoError(t, err)

				require.Equal(t, string(expectedOutput), actualOutput)
			})
		}
	})
	t.Run("Concourse", func(t *testing.T) {
		dirs := []string{"concourse-overlay"}
		for _, dir := range dirs {
			t.Run(dir, func(t *testing.T) {
				dirPath := fmt.Sprintf("../../examples/%s", dir)
				actualOutput := runYtt(t, testInputFiles{dirPath}, "", nil, nil)

				expectedOutput, err := os.ReadFile(filepath.Join(dirPath, "expected.txt"))
				require.NoError(t, err)

				require.Equal(t, string(expectedOutput), actualOutput)
			})
		}
	})
	t.Run("Eirini", func(t *testing.T) {
		t.Run("config.yml", func(t *testing.T) {
			actualOutput := runYtt(t, []string{"../../examples/eirini/config.yml", "../../examples/eirini/input.yml"}, "", nil, nil)

			expectedFileOutput, err := os.ReadFile("../../examples/eirini/config-result.yml")
			require.NoError(t, err)
			require.Equal(t, string(expectedFileOutput), actualOutput)
		})
		t.Run("config-alt1.yml", func(t *testing.T) {
			actualOutput := runYtt(t, []string{"../../examples/eirini/config-alt1.yml", "../../examples/eirini/input.yml"}, "", nil, nil)

			expectedFileOutput, err := os.ReadFile("../../examples/eirini/config-result.yml")
			require.NoError(t, err)
			require.Equal(t, string(expectedFileOutput), actualOutput)
		})
		t.Run("config-alt2.yml and friends", func(t *testing.T) {
			actualOutput := runYtt(t, []string{"../../examples/eirini/config-alt2.yml",
				"../../examples/eirini/config.lib.yaml",
				"../../examples/eirini/config.star",
				"../../examples/eirini/data.txt",
				"../../examples/eirini/data2.lib.txt",
				"../../examples/eirini/input.yml"},
				"", nil, nil)

			expectedFileOutput, err := os.ReadFile("../../examples/eirini/config-result.yml")
			require.NoError(t, err)
			require.Equal(t, string(expectedFileOutput), actualOutput)
		})
	})
}

func TestRemainingExamples(t *testing.T) {
	dirs := []string{
		// test that @ytt:toml module works in default ytt binary
		// as it is loaded in a different way than other @ytt:* modules.
		"toml-serialize",
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			dirPath := fmt.Sprintf("../../examples/%s", dir)
			actualOutput := runYtt(t, testInputFiles{dirPath}, "", nil, nil)

			expectedOutput, err := os.ReadFile(filepath.Join(dirPath, "expected.txt"))
			require.NoError(t, err)

			require.Equal(t, string(expectedOutput), actualOutput)
		})
	}
}

type testInputFiles []string

type yttFlags []map[string]string

func runYtt(t *testing.T, files testInputFiles, stdinFileName string, flags yttFlags, envs []string) string {
	command, stdError := buildCommand(files, flags, envs)

	if stdinFileName != "" {
		fileToUseInStdIn, err := os.OpenFile(stdinFileName, os.O_RDONLY, os.ModeAppend)
		require.NoError(t, err)
		command.Stdin = fileToUseInStdIn
	}
	output, err := command.Output()

	require.NoError(t, err, stdError.String())

	return string(output)
}

func runYttExpectingError(t *testing.T, files testInputFiles, flags yttFlags, envs []string) string {
	command, stdError := buildCommand(files, flags, envs)

	_, err := command.Output()
	require.Error(t, err, stdError.String())

	return stdError.String()

}

func buildCommand(files testInputFiles, flags yttFlags, envs []string) (*exec.Cmd, *bytes.Buffer) {
	var fileFlags []string
	for _, file := range files {
		fileFlags = append(fileFlags, "-f", file)
	}

	var yttFlags []string
	for _, flagElement := range flags {
		for flagName, flagVal := range flagElement {
			if flagVal != "" {
				yttFlags = append(yttFlags, flagName, flagVal)
			} else {
				yttFlags = append(yttFlags, flagName)
			}
		}
	}

	command := exec.Command("../../ytt", append(fileFlags, yttFlags...)...)
	stdError := bytes.NewBufferString("")
	command.Stderr = stdError
	command.Env = append(command.Env, envs...)
	return command, stdError
}
