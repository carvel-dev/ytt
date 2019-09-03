package yaml

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	_                = fmt.Printf // debug
	strictStyleInt   = regexp.MustCompile(`^\-?(0|[1-9][0-9]*)$`)
	strictStyleFloat = regexp.MustCompile(`^\-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][-+]?[0-9]+)?$`)
)

func strictScalarResolve(tag, in string) (string, interface{}) {
	// fmt.Printf("resolve: '%s' '%s'\n", tag, in)

	nativeTag, nativeVal := resolve(tag, in)
	if len(tag) > 0 {
		return nativeTag, nativeVal
	}

	conTag, conVal := strictScalarResolveConservative(in)

	if conTag != nativeTag {
		failf("Strict parsing: Found '%s' ambigious (could be %s or %s)",
			in, shortTag(conTag), shortTag(nativeTag))
		panic("Unreachable")
	}

	if !reflect.DeepEqual(conVal, nativeVal) {
		failf("Strict parsing: Found '%s' ambigious (could be '%s' or '%s')",
			in, conVal, nativeVal)
		panic("Unreachable")
	}

	return conTag, conVal
}

func strictScalarResolveConservative(in string) (string, interface{}) {
	switch in {
	case "":
		return yaml_NULL_TAG, nil
	case "true":
		return yaml_BOOL_TAG, true
	case "false":
		return yaml_BOOL_TAG, false

	default:
		switch {
		case strictStyleInt.MatchString(in):
			intv, err := strconv.ParseInt(in, 0, 64)
			if err != nil {
				uintv, err := strconv.ParseUint(in, 0, 64)
				if err == nil {
					return yaml_INT_TAG, uintv
				}

				failf("Strict parsing: Parsing int '%s': %s", in, err)
				panic("Unreachable")
			}
			if intv == int64(int(intv)) {
				return yaml_INT_TAG, int(intv)
			}
			return yaml_INT_TAG, intv

		case strictStyleFloat.MatchString(in):
			floatv, err := strconv.ParseFloat(in, 64)
			if err != nil {
				failf("Strict parsing: Parsing float '%s': %s", in, err)
				panic("Unreachable")
			}
			return yaml_FLOAT_TAG, floatv

		case strings.IndexFunc(in, unicode.IsSpace) != -1:
			failf("Strict parsing: Strings with whitespace must be explicitly quoted: '%s'", in)
			panic("Unreachable")

		case strings.Contains(in, ":"):
			failf("Strict parsing: Strings with colon must be explicitly quoted: '%s'", in)
			panic("Unreachable")

		// Catch missing new line before document start
		case strings.Contains(in, "---"):
			failf("Strict parsing: Strings with triple-dash must be explicitly quoted: '%s'", in)
			panic("Unreachable")

		default:
			return yaml_STR_TAG, in
		}
	}
}
