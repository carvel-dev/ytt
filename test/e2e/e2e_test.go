// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckStdInReading(t *testing.T) {
	actualOutput := runYtt(t, []string{"-", "../../examples/eirini/input.yml"}, "../../examples/eirini/config.yml", nil, nil)

	expectedFileOutput, err := ioutil.ReadFile("../../examples/eirini/config-result.yml")
	require.NoError(t, err)
	require.Equal(t, string(expectedFileOutput), actualOutput)
}

func TestSanityCheckTemplateWithDataValues(t *testing.T) {
	t.Run("template file with data value", func(t *testing.T) {
		actualOutput := runYtt(t, []string{"../../examples/eirini/config.yml", "../../examples/eirini/input.yml"}, "", nil, nil)

		expectedFileOutput, err := ioutil.ReadFile("../../examples/eirini/config-result.yml")
		require.NoError(t, err)
		require.Equal(t, string(expectedFileOutput), actualOutput)
	})

	t.Run("another template file with data value", func(t *testing.T) {
		actualOutput := runYtt(t, []string{"../../examples/eirini/config-alt1.yml", "../../examples/eirini/input.yml"}, "", nil, nil)

		expectedFileOutput, err := ioutil.ReadFile("../../examples/eirini/config-result.yml")
		require.NoError(t, err)
		require.Equal(t, string(expectedFileOutput), actualOutput)
	})
}

func TestCheckDirectoryReading(t *testing.T) {
	tempOutputDir, err := ioutil.TempDir(os.TempDir(), "ytt-check-dir")
	require.NoError(t, err)
	defer os.Remove(tempOutputDir)
	flags := yttFlags{{fmt.Sprintf("--dangerous-emptied-output-directory=%s", tempOutputDir): ""}}
	runYtt(t, []string{"../../examples/eirini/"}, "", flags, nil)

	expectedFileOutput, err := ioutil.ReadFile("../../examples/eirini/config-result.yml")
	require.NoError(t, err)

	actualOutput, err := ioutil.ReadFile(filepath.Join(tempOutputDir, "config-result.yml"))
	require.NoError(t, err)

	require.Equal(t, string(expectedFileOutput), string(actualOutput))
}

func TestPlaygroundExamples(t *testing.T) {

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
}

func TestOverlays(t *testing.T) {
	t.Run(fmt.Sprintf("overlay ../../examples/overlay"), func(t *testing.T) {
		runYtt(t, testInputFiles{"../../examples/overlay"}, "", nil, nil)
	})

	t.Run(fmt.Sprintf("overlay ../../examples/overlay-files"), func(t *testing.T) {
		runYtt(t, testInputFiles{"../../examples/overlay-files"}, "", nil, nil)
	})

	t.Run(fmt.Sprintf("overlay ../../examples/overlay-regular-files"), func(t *testing.T) {
		runYtt(t, testInputFiles{"../../examples/overlay"}, "", yttFlags{{"--file-mark": "file.yml:type=yaml-plain"}}, nil)
	})
}

func TestDifferentUsages(t *testing.T) {
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
	for _, k8sDirToTest := range dirs {
		t.Run(fmt.Sprintf("k8s: %s", k8sDirToTest), func(t *testing.T) {
			dirPath := fmt.Sprintf("../../examples/%s", k8sDirToTest)
			actualOutput := runYtt(t, testInputFiles{dirPath}, "", nil, nil)

			expectedOutput, err := ioutil.ReadFile(filepath.Join(dirPath, "expected.txt"))
			require.NoError(t, err)

			require.Equal(t, string(expectedOutput), actualOutput)
		})
	}

	dirs = []string{"concourse-overlay"}
	for _, concourseDirToTest := range dirs {
		t.Run(fmt.Sprintf("concourse: %s", concourseDirToTest), func(t *testing.T) {
			dirPath := fmt.Sprintf("../../examples/%s", concourseDirToTest)
			actualOutput := runYtt(t, testInputFiles{dirPath}, "", nil, nil)

			expectedOutput, err := ioutil.ReadFile(filepath.Join(dirPath, "expected.txt"))
			require.NoError(t, err)

			require.Equal(t, string(expectedOutput), actualOutput)
		})
	}

	dirs = []string{"overlay-not-matcher"}
	for _, overlayDirToTest := range dirs {
		t.Run(fmt.Sprintf("overlay-not-matcher: %s", overlayDirToTest), func(t *testing.T) {
			dirPath := fmt.Sprintf("../../examples/%s", overlayDirToTest)
			actualOutput := runYtt(t, testInputFiles{dirPath}, "", nil, nil)

			expectedOutput, err := ioutil.ReadFile(filepath.Join(dirPath, "expected.txt"))
			require.NoError(t, err)

			require.Equal(t, string(expectedOutput), actualOutput)
		})
	}

	dirs = []string{
		"schema",
		"schema-arrays",
	}
	for _, schemaDirToTest := range dirs {
		t.Run(fmt.Sprintf("schema: %s", schemaDirToTest), func(t *testing.T) {
			dirPath := fmt.Sprintf("../../examples/%s", schemaDirToTest)
			actualOutput := runYtt(t, testInputFiles{dirPath}, "", yttFlags{{"--enable-experiment-schema": ""}}, nil)

			expectedOutput, err := ioutil.ReadFile(filepath.Join(dirPath, "expected.txt"))
			require.NoError(t, err)

			require.Equal(t, string(expectedOutput), actualOutput)
		})
	}
}

func TestJsonOutput(t *testing.T) {
	actualOutput := runYtt(t, testInputFiles{"../../examples/k8s-adjust-rbac-version"}, "", yttFlags{{"-o": "json"}}, nil)
	expectedOutput := `{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"ClusterRole"}{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"ClusterRole"}`

	require.Equal(t, expectedOutput, actualOutput)
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

func TestDataValues(t *testing.T) {
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
}

func TestSchema(t *testing.T) {
	t.Run("--data-value flag", func(t *testing.T) {
		flags := yttFlags{
			{"--data-value": "nothing=a new string"},
			{"--data-value": "string=str"},
			{"--enable-experiment-schema": ""},
		}

		actualOutput := runYtt(t, testInputFiles{"../../examples/schema/config.yml", "../../examples/schema/schema.yml"}, "", flags, nil)
		expectedOutput := `nothing: a new string
string: str
bool: false
int: 0
float: 0.1
any: anything
`

		require.Equal(t, expectedOutput, actualOutput)
	})

	t.Run("--data-value-yaml flag", func(t *testing.T) {
		flags := yttFlags{
			{"--data-value-yaml": "nothing=a new string"},
			{"--data-value-yaml": "string=str"},
			{"--data-value-yaml": "bool=true"},
			{"--data-value-yaml": "int=123"},
			{"--data-value-yaml": "float=123.123"},
			{"--data-value-yaml": "any=[1,2,4]"},
			{"--enable-experiment-schema": ""},
		}
		actualOutput := runYtt(t, testInputFiles{"../../examples/schema/config.yml", "../../examples/schema/schema.yml"}, "", flags, nil)
		expectedOutput := `nothing: a new string
string: str
bool: true
int: 123
float: 123.123
any:
- 1
- 2
- 4
`

		require.Equal(t, expectedOutput, actualOutput)
	})

	t.Run("--data-values-env && --data-values-env-yaml flag", func(t *testing.T) {
		flags := yttFlags{
			{"--data-values-env": "STR_VAL"},
			{"--data-values-env-yaml": "YAML_VAL"},
			{"--enable-experiment-schema": ""},
		}
		envs := []string{
			"STR_VAL_nothing=a new string",
			"YAML_VAL_string=str",
			"YAML_VAL_bool=true",
			"YAML_VAL_int=123",
			"YAML_VAL_float=123.123",
			"YAML_VAL_any=[1,2,4]",
		}
		actualOutput := runYtt(t, testInputFiles{"../../examples/schema/config.yml", "../../examples/schema/schema.yml"}, "", flags, envs)
		expectedOutput := `nothing: a new string
string: str
bool: true
int: 123
float: 123.123
any:
- 1
- 2
- 4
`

		require.Equal(t, expectedOutput, actualOutput)
	})
}

func TestDataValuesRequired(t *testing.T) {
	expectedFileOutput, err := ioutil.ReadFile("../../examples/data-values-required/expected.txt")
	require.NoError(t, err)

	t.Run("inline", func(t *testing.T) {
		flags := yttFlags{
			{"-v": "version=123"},
		}
		actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-required/inline"}, "", flags, nil)

		require.Equal(t, string(expectedFileOutput), actualOutput)
	})

	t.Run("function", func(t *testing.T) {
		flags := yttFlags{
			{"-v": "version=123"},
		}
		actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-required/function"}, "", flags, nil)

		require.Equal(t, string(expectedFileOutput), actualOutput)
	})

	t.Run("bulk", func(t *testing.T) {
		flags := yttFlags{
			{"-v": "version=123"},
		}
		actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-required/bulk"}, "", flags, nil)

		require.Equal(t, string(expectedFileOutput), actualOutput)
	})
}

func TestDataValuesUsages(t *testing.T) {
	t.Run("multiple envs", func(t *testing.T) {
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

	t.Run("when using data values in dev", func(t *testing.T) {
		actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-multiple-envs/config/", "../../examples/data-values-multiple-envs/envs/dev.yml"}, "", nil, nil)
		expectedOutput := `app_config:
  version: v1alpha1
  ports:
  - 8080
`
		require.Equal(t, expectedOutput, actualOutput)
	})

	t.Run("when using data values in staging", func(t *testing.T) {
		actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-multiple-envs/config/", "../../examples/data-values-multiple-envs/envs/staging.yml"}, "", nil, nil)
		expectedOutput := `app_config:
  version: v1beta1
  ports:
  - 8081
`
		require.Equal(t, expectedOutput, actualOutput)
	})

	t.Run("when using data values in prod", func(t *testing.T) {
		actualOutput := runYtt(t, testInputFiles{"../../examples/data-values-multiple-envs/config/", "../../examples/data-values-multiple-envs/envs/prod.yml"}, "", nil, nil)
		expectedOutput := `app_config:
  version: v1
  ports:
  - 80
`
		require.Equal(t, expectedOutput, actualOutput)
	})
}

func TestVersionIsValid(t *testing.T) {
	runYtt(t, testInputFiles{"../../examples/version-constraint"}, "", yttFlags{}, nil)
}

type testInputFiles []string

type yttFlags []map[string]string

func runYtt(t *testing.T, files testInputFiles, stdinFileName string, flags yttFlags, envs []string) string {
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

	if stdinFileName != "" {
		fileToUseInStdIn, err := os.OpenFile(stdinFileName, os.O_RDONLY, os.ModeAppend)
		require.NoError(t, err)
		command.Stdin = fileToUseInStdIn
	}
	output, err := command.Output()
	require.NoError(t, err, stdError.String())

	return string(output)
}
