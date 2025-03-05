package codec

import (
	"encoding/json"
	"errors"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils/validators"

	"io"
	"net/url"
	"reflect"
)

type JSONCodec struct {
	inputValidator *validators.InputValidator
}

func NewJSONCodec() *JSONCodec {
	return &JSONCodec{inputValidator: validators.NewInputValidator()}
}

func (j *JSONCodec) Unmarshal(data []byte, v any) error {
	err := json.Unmarshal(data, v)

	if err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			return j.inputValidator.ErrorMessageCreator.GetUnmarshalTypeError(v, unmarshalTypeError)
		}

		return err
	}

	return j.inputValidator.Validate(v)
}

func (j *JSONCodec) UnmarshalQueryParam(queryParams url.Values, v any) error {
	val := reflect.Indirect(reflect.ValueOf(v))
	t := val.Type()

	creator := j.inputValidator.ErrorMessageCreator.GetErrorMessageCreator(t)

	jsonData := make(map[string]interface{})
	for key, values := range queryParams {
		if creator.SliceFields[key] {
			jsonData[key] = values
		} else {
			jsonData[key] = values[0]
		}
	}

	marshal, errm := json.Marshal(jsonData)
	if errm != nil {
		return errm
	}

	err := json.Unmarshal(marshal, v)

	if err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			return j.inputValidator.ErrorMessageCreator.GetUnmarshalTypeError(v, unmarshalTypeError)
		}

		return err
	}

	return j.inputValidator.Validate(v)
}

func (j *JSONCodec) Decode(r io.ReadCloser, v any) error {
	defer func() { _ = r.Close() }()

	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		return err
	}

	return j.inputValidator.Validate(v)
}
