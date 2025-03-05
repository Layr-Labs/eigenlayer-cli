package types

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strconv"
)

type JSONInt32 int32
type JSONInt64 int64
type JSONFloat float32
type JSONFloat64 float64
type JSONString string

func (f *JSONString) UnmarshalJSON(b []byte) error {

	if len(b) == 0 {
		return errors.New("input is empty")
	}

	if b[0] == '"' {
		// If it's a quoted string, unmarshal directly to a string
		var str string
		err := json.Unmarshal(b, &str)
		if err != nil {
			return err
		}
		*f = JSONString(str)
		return nil
	} else {
		// If it's a number, convert it to string
		var num json.Number
		err := json.Unmarshal(b, &num)
		if err != nil {
			return err
		}
		*f = JSONString(num.String())
		return nil
	}
}

func (f *JSONInt32) UnmarshalJSON(b []byte) error {
	val, err := unmarshalJSON[int32](reflect.TypeOf(int32(0)), b, func(val string) (int32, error) {
		v, e := strconv.ParseInt(val, 10, 64)
		if e != nil {
			return 0, e
		}
		return safeIntToInt32(v)
	}, 0)
	if err == nil {
		*f = JSONInt32(val)
	}
	return err
}

func (f *JSONInt64) UnmarshalJSON(b []byte) error {
	val, err := unmarshalJSON[int64](reflect.TypeOf(int64(0)), b, func(val string) (int64, error) {
		return strconv.ParseInt(val, 10, 64)
	}, 0)
	if err == nil {
		*f = JSONInt64(val)
	}
	return err
}

func (f *JSONFloat) UnmarshalJSON(b []byte) error {
	val, err := unmarshalJSON[float32](reflect.TypeOf(int(0)), b, func(val string) (float32, error) {
		v, e := strconv.ParseFloat(val, 32)
		return float32(v), e
	}, 0)
	if err == nil {
		*f = JSONFloat(val)
	}
	return err
}

func (f *JSONFloat64) UnmarshalJSON(b []byte) error {
	val, err := unmarshalJSON[float64](reflect.TypeOf(int(0)), b, func(val string) (float64, error) {
		return strconv.ParseFloat(val, 64)
	}, 0)
	if err == nil {
		*f = JSONFloat64(val)
	}
	return err
}

func unmarshalJSON[T any](
	reflectType reflect.Type,
	b []byte,
	parse func(val string) (T, error),
	defaultValue T,
) (T, error) {
	switch {
	case len(b) == 0:
		return defaultValue, &json.UnmarshalTypeError{
			Value:  string(""),  // The raw data that caused the error
			Type:   reflectType, // The expected Go type
			Offset: 0,           // The offset in the JSON input
		}
	case b[0] == '"':
		var targetStr string
		err := json.Unmarshal(b, &targetStr)
		if err != nil {
			return defaultValue, err
		}
		parsedInt, err := parse(targetStr)
		if err != nil {
			return defaultValue, &json.UnmarshalTypeError{
				Value:  targetStr,   // The raw data that caused the error
				Type:   reflectType, // The expected Go type
				Offset: 0,           // The offset in the JSON input
			}
		}

		return parsedInt, nil
	default:
		var targetInt T
		err := json.Unmarshal(b, &targetInt)
		if err != nil {
			return defaultValue, err
		}

		return targetInt, nil
	}
}

func safeIntToInt32(value int64) (int32, error) {
	if value > math.MaxInt32 || value < math.MinInt32 {
		return 0, errors.New("integer overflow: cannot convert int to int32")
	}
	// no lint gosec,  as the int32 casting has been safely done here by checking with max values
	// no lint nolintlint as the git-hub says no usage of gosec and fails pull request's build pipeline
	int32Value := int32(value) //nolint:gosec,nolintlint

	return int32Value, nil
}
