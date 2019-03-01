package cobrautil

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	cobra.AddTemplateFunc("commandsWithAnnotation", func(cmd *cobra.Command, key, value string) []*cobra.Command {
		var result []*cobra.Command
		for _, c := range cmd.Commands() {
			anns := map[string]string{}
			if c.Annotations != nil {
				anns = c.Annotations
			}
			if anns[key] == value {
				result = append(result, c)
			}
		}
		return result
	})
}

type HelpSection struct {
	Key   string
	Value string
	Title string
}

func HelpSectionsUsageTemplate(sections []HelpSection) string {
	unmodifiedCmd := &cobra.Command{}
	usageTemplate := unmodifiedCmd.UsageTemplate()

	const defaultTpl = `{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}`

	if !strings.Contains(usageTemplate, defaultTpl) {
		panic("Expected to find available commands section in spf13/cobra default usage template")
	}

	newTpl := "{{if .HasAvailableSubCommands}}"

	for _, section := range sections {
		newTpl += fmt.Sprintf(`{{$cmds := (commandsWithAnnotation . "%s" "%s")}}{{if $cmds}}

%s{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}`, section.Key, section.Value, section.Title)
	}

	newTpl += "{{end}}"

	return strings.Replace(usageTemplate, defaultTpl, newTpl, 1)
}
