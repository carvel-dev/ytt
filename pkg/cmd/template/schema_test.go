package template_test

import (
	"strings"
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
)

func TestDataValuesConformingToSchemaChecksOk(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
db_conn:
  hostname: ""
  port: 0
  username: ""
  password: ""
`
	dataValuesYAML := `#@data/values
---
db_conn:
  hostname: server.example.com
  port: 5432
  username: sa
  password: changeme
`
	templateYAML := `---
rendered: true`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was: %d", len(out.Files))
	}

	if string(out.Files[0].Bytes()) != "rendered: true\n" {
		t.Fatalf("Expected output to only include template YAML, but got: %s", out.Files[0].Bytes())
	}
}

func TestDataValuesNotConformingToSchemaFailsCheck(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
`
	dataValuesYAML := `#@data/values
---
not_in_schema: "this should fail the type check!"
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)

	if out.Err == nil {
		t.Fatalf("Expected an error about a schema check failure, but succeeded.")
	}
	expectedErr := "Typechecking violations found: [Map item 'not_in_schema' at dataValues.yml:3 is not defined in schema]"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about a schema check failure, but got: %s", out.Err.Error())
	}

}

func TestMultiDataValuesOneDataValuesNotConformingToSchemaFailsCheck(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
hostname: ""
port: 0
username: ""
password: ""
`
	dataValuesYAML1 := `#@data/values
---
hostname: server.example.com
port: 5432
secret: super
`

	dataValuesYAML2 := `#@ load("@ytt:overlay", "overlay")
#@data/values
---
#@overlay/remove
secret:
#@overlay/match missing_ok=True
username: sa
#@overlay/match missing_ok=True
password: changeme
`
	templateYAML := `---
rendered: true`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues1.yml", []byte(dataValuesYAML1))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues2.yml", []byte(dataValuesYAML2))),
		files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected an error about a schema check failure, but succeeded.")
	}

	expectedErr := "Typechecking violations found: [Map item 'secret' at dataValues1.yml:5 is not defined in schema]"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about a schema check failure, but got: %s", out.Err.Error())
	}
}
