package yaml

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var _ = fmt.Printf

var yamlStyleInt = regexp.MustCompile(`^[-+]?[0-9]+([eE][-+][0-9]+)?$`)

func strictScalarResolve(tag string, in string) (string, interface{}) {
	// fmt.Printf("resolve: '%s' '%s'\n", tag, in)

	if len(tag) > 0 {
		return resolve(tag, in)
	}

	switch in {
	case "":
		return yaml_NULL_TAG, nil
	case "true":
		return yaml_BOOL_TAG, true
	case "false":
		return yaml_BOOL_TAG, false

	default:
		switch {
		case yamlStyleInt.MatchString(in):
			intv, err := strconv.ParseInt(in, 0, 64)
			if err != nil {
				failf("Parsing int '%s': %s", in, err)
				panic("Unreachable")
			}
			return yaml_INT_TAG, intv

		case yamlStyleFloat.MatchString(in):
			floatv, err := strconv.ParseFloat(in, 64)
			if err != nil {
				failf("Parsing float '%s': %s", in, err)
				panic("Unreachable")
			}
			return yaml_FLOAT_TAG, floatv

		case strings.IndexFunc(in, unicode.IsSpace) != -1:
			failf("Strings with whitespace must be explicitly quoted: '%s'", in)
			panic("Unreachable")

		case strings.Contains(in, ":"):
			failf("Strings with colon must be explicitly quoted: '%s'", in)
			panic("Unreachable")

		default:
			return yaml_STR_TAG, in
		}
	}
}
