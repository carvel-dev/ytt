package cobrautil

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type FlagHelpSection struct {
	Title string

	PrefixMatch string
	ExactMatch  []string
	NoneMatch   bool
}

func (s FlagHelpSection) Matches(name string) bool {
	if len(s.PrefixMatch) > 0 && strings.HasPrefix(name, s.PrefixMatch+"-") {
		return true
	}
	for _, em := range s.ExactMatch {
		if name == em {
			return true
		}
	}
	return false
}

type flagLine struct {
	Line        string
	SectionIdxs []int
}

func (l flagLine) InSectionIdx(idx int) bool {
	for _, i := range l.SectionIdxs {
		if i == idx {
			return true
		}
	}
	return false
}

var (
	flagHelpFuncCount = 0
	flagNameRegexp    = regexp.MustCompile("^\\s+(\\-[a-z], )?\\-\\-([a-z\\-]+)\\s+")
)

func FlagHelpSectionsUsageTemplate(sections []FlagHelpSection) string {
	flagHelpFuncCount += 1
	flagHelpFuncName := fmt.Sprintf("flagsWithSections%d", flagHelpFuncCount)

	cobra.AddTemplateFunc(flagHelpFuncName, func(str string) string {
		lines := strings.Split(str, "\n")
		flags := map[string]flagLine{}
		flagNames := []string{}

		for _, line := range lines {
			match := flagNameRegexp.FindStringSubmatch(line)
			flagName := match[2]

			if _, found := flags[flagName]; found {
				panic("Expected to not find multiple flags with same name")
			}

			fline := flagLine{Line: line}
			noneMatchIdx := -1

			for i, section := range sections {
				if section.Matches(flagName) {
					fline.SectionIdxs = append(fline.SectionIdxs, i)
				}
				if section.NoneMatch {
					noneMatchIdx = i
				}
			}
			if len(fline.SectionIdxs) == 0 {
				fline.SectionIdxs = []int{noneMatchIdx}
			}

			flags[flagName] = fline
			flagNames = append(flagNames, flagName)
		}

		sort.Strings(flagNames)

		sectionsResult := []string{}

		for i, section := range sections {
			result := section.Title + "\n"
			for _, name := range flagNames {
				fline := flags[name]
				if fline.InSectionIdx(i) {
					result += fline.Line + "\n"
				}
			}
			sectionsResult = append(sectionsResult, result)
		}

		return strings.TrimSpace(strings.Join(sectionsResult, "\n"))
	})

	unmodifiedCmd := &cobra.Command{}
	usageTemplate := unmodifiedCmd.UsageTemplate()

	const defaultTpl = `{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}`

	if !strings.Contains(usageTemplate, defaultTpl) {
		panic("Expected to find available flags section in spf13/cobra default usage template")
	}

	newTpl := fmt.Sprintf(`{{if .HasAvailableLocalFlags}}

{{.LocalFlags.FlagUsages | trimTrailingWhitespaces | %s}}{{end}}`, flagHelpFuncName)

	return strings.Replace(usageTemplate, defaultTpl, newTpl, 1)
}
