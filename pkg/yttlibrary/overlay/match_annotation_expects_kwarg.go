package overlay

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"go.starlark.net/starlark"
)

const (
	MatchAnnotationKwargBy        string = "by"
	MatchAnnotationKwargExpects   string = "expects"
	MatchAnnotationKwargMissingOK string = "missing_ok"
)

type MatchAnnotationExpectsKwarg struct {
	expects   *starlark.Value
	missingOK *starlark.Value
	thread    *starlark.Thread
}

func (a *MatchAnnotationExpectsKwarg) FillInDefaults(defaults MatchChildDefaultsAnnotation) {
	if a.expects == nil {
		a.expects = defaults.expects.expects
	}
	if a.missingOK == nil {
		a.missingOK = defaults.expects.missingOK
	}
}

func (a MatchAnnotationExpectsKwarg) Check(matches []*filepos.Position) error {
	switch {
	case a.missingOK != nil && a.expects != nil:
		return fmt.Errorf("Expected only one of keyword arguments ('%s', '%s') specified",
			MatchAnnotationKwargMissingOK, MatchAnnotationKwargExpects)

	case a.missingOK != nil:
		if typedResult, ok := (*a.missingOK).(starlark.Bool); ok {
			if typedResult {
				allowedVals := []starlark.Value{starlark.MakeInt(0), starlark.MakeInt(1)}
				return a.checkValue(starlark.NewList(allowedVals), matches)
			}
			return a.checkValue(starlark.MakeInt(1), matches)
		}
		return fmt.Errorf("Expected keyword argument '%s' to be a boolean",
			MatchAnnotationKwargMissingOK)

	case a.expects != nil:
		return a.checkValue(*a.expects, matches)

	default:
		return a.checkValue(starlark.MakeInt(1), matches)
	}
}

func (a MatchAnnotationExpectsKwarg) checkValue(val interface{}, matches []*filepos.Position) error {
	switch typedVal := val.(type) {
	case starlark.Int:
		return a.checkInt(typedVal, matches)

	case starlark.String:
		return a.checkString(typedVal, matches)

	case *starlark.List:
		return a.checkList(typedVal, matches)

	case starlark.Callable:
		result, err := starlark.Call(a.thread, typedVal, starlark.Tuple{starlark.MakeInt(len(matches))}, []starlark.Tuple{})
		if err != nil {
			return err
		}
		if typedResult, ok := result.(starlark.Bool); ok {
			if !bool(typedResult) {
				return fmt.Errorf("Expectation of number of matched nodes failed")
			}
			return nil
		}
		return fmt.Errorf("Expected keyword argument 'expects' to have a function that returns a boolean")

	default:
		return fmt.Errorf("Expected '%s' annotation keyword argument 'expects'"+
			" to be either int, string or function, but was %T", AnnotationMatch, typedVal)
	}
}

func (a MatchAnnotationExpectsKwarg) checkInt(typedVal starlark.Int, matches []*filepos.Position) error {
	i1, ok := typedVal.Int64()
	if ok {
		if i1 != int64(len(matches)) {
			return fmt.Errorf("Expected number of matched nodes to be %d, but was %d%s",
				i1, len(matches), a.formatPositions(matches))
		}
		return nil
	}

	i2, ok := typedVal.Uint64()
	if ok {
		if i2 != uint64(len(matches)) {
			return fmt.Errorf("Expected number of matched nodes to be %d, but was %d%s",
				i2, len(matches), a.formatPositions(matches))
		}
		return nil
	}

	panic("Unsure how to convert starlark.Int to int")
}

func (a MatchAnnotationExpectsKwarg) checkString(typedVal starlark.String, matches []*filepos.Position) error {
	typedValStr := string(typedVal)

	if strings.HasSuffix(typedValStr, "+") {
		typedInt, err := strconv.Atoi(strings.TrimSuffix(typedValStr, "+"))
		if err != nil {
			return fmt.Errorf("Expected '%s' to be in format 'i+' where i is an integer", typedValStr)
		}

		if len(matches) < typedInt {
			return fmt.Errorf("Expected number of matched nodes to be >= %d, but was %d%s",
				typedInt, len(matches), a.formatPositions(matches))
		}

		return nil
	}

	return fmt.Errorf("Expected '%s' to be in format 'i+' where i is an integer", typedValStr)
}

func (a MatchAnnotationExpectsKwarg) checkList(typedVal *starlark.List, matches []*filepos.Position) error {
	var lastErr error
	var val starlark.Value

	iter := typedVal.Iterate()
	defer iter.Done()

	for iter.Next(&val) {
		lastErr = a.checkValue(val, matches)
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func (MatchAnnotationExpectsKwarg) formatPositions(pos []*filepos.Position) string {
	if len(pos) == 0 {
		return ""
	}
	lines := []string{}
	for _, p := range pos {
		lines = append(lines, p.AsCompactString())
	}
	return " (lines: " + strings.Join(lines, ", ") + ")"
}
