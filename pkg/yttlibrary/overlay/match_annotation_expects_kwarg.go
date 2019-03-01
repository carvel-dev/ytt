package overlay

import (
	"fmt"
	"strconv"
	"strings"

	"go.starlark.net/starlark"
)

type MatchAnnotationExpectsKwarg struct {
	expects   *starlark.Value
	missingOK *starlark.Value
	thread    *starlark.Thread
}

func (a MatchAnnotationExpectsKwarg) Check(num int) error {
	switch {
	case a.missingOK != nil && a.expects != nil:
		return fmt.Errorf("Expected only one of keyword arguments ('missing_ok', 'expects') specified")

	case a.missingOK != nil:
		if typedResult, ok := (*a.missingOK).(starlark.Bool); ok {
			if typedResult {
				allowedVals := []starlark.Value{starlark.MakeInt(0), starlark.MakeInt(1)}
				return a.checkValue(starlark.NewList(allowedVals), num)
			}
			return a.checkValue(starlark.MakeInt(1), num)
		}
		return fmt.Errorf("Expected keyword argument 'missing_ok' to be a boolean")

	case a.expects != nil:
		return a.checkValue(*a.expects, num)

	default:
		return a.checkValue(starlark.MakeInt(1), num)
	}
}

func (a MatchAnnotationExpectsKwarg) checkValue(val interface{}, num int) error {
	switch typedVal := val.(type) {
	case starlark.Int:
		return a.checkInt(typedVal, num)

	case starlark.String:
		return a.checkString(typedVal, num)

	case *starlark.List:
		return a.checkList(typedVal, num)

	case starlark.Callable:
		result, err := starlark.Call(a.thread, typedVal, starlark.Tuple{starlark.MakeInt(num)}, []starlark.Tuple{})
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

func (a MatchAnnotationExpectsKwarg) checkInt(typedVal starlark.Int, num int) error {
	i1, ok := typedVal.Int64()
	if ok {
		if i1 != int64(num) {
			return fmt.Errorf("Expected number of matched nodes to be %d, but was %d", i1, num)
		}
		return nil
	}

	i2, ok := typedVal.Uint64()
	if ok {
		if i2 != uint64(num) {
			return fmt.Errorf("Expected number of matched nodes to be %d, but was %d", i2, num)
		}
		return nil
	}

	panic("Unsure how to convert starlark.Int to int")
}

func (a MatchAnnotationExpectsKwarg) checkString(typedVal starlark.String, num int) error {
	typedValStr := string(typedVal)

	if strings.HasSuffix(typedValStr, "+") {
		typedInt, err := strconv.Atoi(strings.TrimSuffix(typedValStr, "+"))
		if err != nil {
			return fmt.Errorf("Expected '%s' to be in format 'i+' where i is an integer", typedValStr)
		}

		if num < typedInt {
			return fmt.Errorf("Expected number of matched nodes to be >= %d, but was %d", typedInt, num)
		}

		return nil
	}

	return fmt.Errorf("Expected '%s' to be in format 'i+' where i is an integer", typedValStr)
}

func (a MatchAnnotationExpectsKwarg) checkList(typedVal *starlark.List, num int) error {
	var lastErr error
	var val starlark.Value

	iter := typedVal.Iterate()
	defer iter.Done()

	for iter.Next(&val) {
		lastErr = a.checkValue(val, num)
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}
