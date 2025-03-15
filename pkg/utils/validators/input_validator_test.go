package validators

import (
	"encoding/json"
	"errors"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils/codec/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalValid(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required"`
		ID   string `json:"id"   validate:"uuidv7"`
	}

	msg := `{"name":"TR", "id":"01927a2c-3286-78d9-a29b-00991eff773b"}`
	iv := NewInputValidator()
	tst := &TestStruct{}
	err := json.Unmarshal([]byte(msg), tst)
	assert.Nil(t, err)
	err2 := iv.Validate(tst)
	assert.Nil(t, err2)
}

func TestUnmarshalInvalidUUIDV7(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required"`
		ID   int    `json:"id"   validate:"uuidv7"`
	}

	msg := `{"name":"TR", "id":"01927a2c-3286-78d9-a29b-00991eff773b"}`
	iv := NewInputValidator()
	tst := &TestStruct{}
	err := json.Unmarshal([]byte(msg), tst)
	assert.NotNil(t, err)
	err2 := iv.Validate(tst)
	assert.NotNil(t, err2)
}

func TestUnmarshalError(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required"`
	}

	msg := `{"namew":"TR"}`
	iv := NewInputValidator()
	tst := &TestStruct{}
	err := json.Unmarshal([]byte(msg), tst)
	assert.Nil(t, err)
	err2 := iv.Validate(tst)
	assert.NotNil(t, err2)
	valErr := &FieldValidationError{}
	assert.True(t, errors.As(err2, &valErr))
	assert.Equal(t, "Name is required", (*valErr)["name"])
}

func TestUnmarshalErrorEmbed(t *testing.T) {
	type TestStruct struct {
		Age string `json:"age" validate:"required"`
	}
	type TestStruct2 struct {
		TestStruct
		Name string `json:"name" validate:"required"`
	}
	msg := `{"namew":"TR"}`
	iv := NewInputValidator()
	tst := &TestStruct2{}
	err := json.Unmarshal([]byte(msg), tst)
	assert.Nil(t, err)
	err2 := iv.Validate(tst)
	assert.NotNil(t, err2)
	valErr := &FieldValidationError{}
	assert.True(t, errors.As(err2, &valErr))
	assert.Equal(t, "Name is required", (*valErr)["name"])
	assert.Equal(t, "Age is required", (*valErr)["age"])
}

func TestUnmarshalErrorEmbedField(t *testing.T) {
	type TestStruct struct {
		Age string `json:"age" validate:"required"`
	}
	type TestStruct2 struct {
		Test TestStruct `json:"test" validate:"required"`
		Name string     `json:"name" validate:"required"`
	}
	msg := `{"namew":"TR"}`
	iv := NewInputValidator()
	tst := &TestStruct2{}
	err := json.Unmarshal([]byte(msg), tst)
	assert.Nil(t, err)
	err2 := iv.Validate(tst)
	assert.NotNil(t, err2)
	valErr := &FieldValidationError{}
	assert.True(t, errors.As(err2, &valErr))
	assert.Equal(t, "Name is required", (*valErr)["name"])
	assert.Equal(t, "Age is required", (*valErr)["age"])
}

func TestUnmarshalCustomError(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required" errors:"required=Grrrr"`
		Age  int
	}

	msg := `{"namew":"TR"}`
	iv := NewInputValidator()
	tst := &TestStruct{}
	err := json.Unmarshal([]byte(msg), tst)
	assert.Nil(t, err)
	err2 := iv.Validate(tst)
	assert.NotNil(t, err2)
	valErr := &FieldValidationError{}
	assert.True(t, errors.As(err2, &valErr))
	assert.Equal(t, "Grrrr", (*valErr)["name"])
}

func TestUnmarshalNoValidation(t *testing.T) {
	iv := NewInputValidator()
	err2 := iv.Validate(5)
	assert.Nil(t, err2)
}

func TestUnmarshalCustomTypes(t *testing.T) {
	type TestStruct struct {
		Name types.JSONInt32 `json:"age" validate:"required"`
	}

	msg := `{"age":"12"}`
	iv := NewInputValidator()
	tst := &TestStruct{}
	err := json.Unmarshal([]byte(msg), tst)
	assert.Nil(t, err)
	err2 := iv.Validate(tst)
	assert.Nil(t, err2)
	assert.Equal(t, 12, int(tst.Name))
}

type TestType int

func (f *TestType) UnmarshalJSON(b []byte) error {
	return errors.New("TestType")
}
