// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/filepos"
)

const (
	MatchAnnotationKwargBy        string = "by"
	MatchAnnotationKwargExpects   string = "expects"
	MatchAnnotationKwargMissingOK string = "missing_ok"
	MatchAnnotationKwargWhen      string = "when"
)

type MatchAnnotationExpectsKwarg struct {
	expects   *starlark.Value
	missingOK *starlark.Value
	when      *starlark.Value
	thread    *starlark.Thread
}

type MatchAnnotationNumMatchError struct {
	message  string
	fromWhen bool
}

func (e MatchAnnotationNumMatchError) Error() string {
	return e.message
}

func (e MatchAnnotationNumMatchError) isConditional() bool {
	return e.fromWhen
}

func (a *MatchAnnotationExpectsKwarg) FillInDefaults(defaults MatchChildDefaultsAnnotation) {
	if a.expects == nil {
		a.expects = defaults.expects.expects
	}
	if a.missingOK == nil {
		a.missingOK = defaults.expects.missingOK
	}
	if a.when == nil {
		a.when = defaults.expects.when
	}
}

func (a MatchAnnotationExpectsKwarg) Check(matches []*filepos.Position) error {
	switch {
	case a.missingOK != nil && a.expects != nil:
		return fmt.Errorf("Expected only one of keyword arguments ('%s', '%s') specified",
			MatchAnnotationKwargMissingOK, MatchAnnotationKwargExpects)

	case a.missingOK != nil && a.when != nil:
		return fmt.Errorf("Expected only one of keyword arguments ('%s', '%s') specified",
			MatchAnnotationKwargMissingOK, MatchAnnotationKwargWhen)

	case a.when != nil && a.expects != nil:
		return fmt.Errorf("Expected only one of keyword arguments ('%s', '%s') specified",
			MatchAnnotationKwargWhen, MatchAnnotationKwargExpects)

	case a.missingOK != nil:
		if typedResult, ok := (*a.missingOK).(starlark.Bool); ok {
			if typedResult {
				allowedVals := []starlark.Value{starlark.MakeInt(0), starlark.MakeInt(1)}
				return a.checkValue(starlark.NewList(allowedVals), MatchAnnotationKwargMissingOK, matches)
			}
			return a.checkValue(starlark.MakeInt(1), MatchAnnotationKwargMissingOK, matches)
		}
		return fmt.Errorf("Expected keyword argument '%s' to be a boolean",
			MatchAnnotationKwargMissingOK)

	case a.when != nil:
		return a.checkValue(*a.when, MatchAnnotationKwargWhen, matches)

	case a.expects != nil:
		return a.checkValue(*a.expects, MatchAnnotationKwargExpects, matches)

	default:
		return a.checkValue(starlark.MakeInt(1), "", matches)
	}
}

func (a MatchAnnotationExpectsKwarg) checkValue(val interface{}, kwarg string, matches []*filepos.Position) error {
	switch typedVal := val.(type) {
	case starlark.Int:
		return a.checkInt(typedVal, matches)

	case starlark.String:
		return a.checkString(typedVal, matches)

	case *starlark.List:
		return a.checkList(typedVal, kwarg, matches)

	case starlark.Callable:
		result, err := starlark.Call(a.thread, typedVal, starlark.Tuple{starlark.MakeInt(len(matches))}, []starlark.Tuple{})
		if err != nil {
			return err
		}
		if typedResult, ok := result.(starlark.Bool); ok {
			if !bool(typedResult) {
				return MatchAnnotationNumMatchError{
					message:  "Expectation of number of matched nodes failed",
					fromWhen: a.when != nil,
				}
			}
			return nil
		}
		return fmt.Errorf("Expected keyword argument '%s' to have a function that returns a boolean", kwarg)

	default:
		return fmt.Errorf("Expected '%s' annotation keyword argument '%s' "+
			"to be either int, string or function, but was %T", AnnotationMatch, kwarg, typedVal)
	}
}

func (a MatchAnnotationExpectsKwarg) checkInt(typedVal starlark.Int, matches []*filepos.Position) error {
	i1, ok := typedVal.Int64()
	if ok {
		if i1 != int64(len(matches)) {
			errMsg := fmt.Sprintf("Expected number of matched nodes to be %d, but was %d%s",
				i1, len(matches), a.formatPositions(matches))
			return MatchAnnotationNumMatchError{message: errMsg, fromWhen: a.when != nil}
		}
		return nil
	}

	i2, ok := typedVal.Uint64()
	if ok {
		if i2 != uint64(len(matches)) {
			errMsg := fmt.Sprintf("Expected number of matched nodes to be %d, but was %d%s",
				i2, len(matches), a.formatPositions(matches))
			return MatchAnnotationNumMatchError{message: errMsg, fromWhen: a.when != nil}
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
			errMsg := fmt.Sprintf("Expected number of matched nodes to be >= %d, but was %d%s",
				typedInt, len(matches), a.formatPositions(matches))
			return MatchAnnotationNumMatchError{message: errMsg, fromWhen: a.when != nil}
		}

		return nil
	}

	return fmt.Errorf("Expected '%s' to be in format 'i+' where i is an integer", typedValStr)
}

func (a MatchAnnotationExpectsKwarg) checkList(typedVal *starlark.List, kwarg string, matches []*filepos.Position) error {
	var lastErr error
	var val starlark.Value

	iter := typedVal.Iterate()
	defer iter.Done()

	for iter.Next(&val) {
		lastErr = a.checkValue(val, kwarg, matches)
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
	sort.Strings(lines)
	return " (lines: " + strings.Join(lines, ", ") + ")"
}
