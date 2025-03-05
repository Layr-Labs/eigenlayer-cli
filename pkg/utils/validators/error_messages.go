// Package validators provides utilities for validating structs and generating error messages
package validators

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// FieldValidationError represents a map of field names to their corresponding error messages
// It implements the Error interface to return error messages in JSON format
// Example: {"fieldName": "error message"}
type FieldValidationError map[string]string

// Error returns the FieldValidationError as a JSON-encoded string
func (f *FieldValidationError) Error() string {
	ret, _ := json.Marshal(f)
	return string(ret)
}

// defaultErrorMessages contains default validation error messages for common validation rules
var defaultErrorMessages = map[string]string{
	"required": "%s is required",
}

// registerValidationErrorMessage allows custom error messages to be registered for specific validation tags
func registerValidationErrorMessage(validator string, message string) {
	defaultErrorMessages[validator] = message
}

// structErrorMessageCreator is responsible for creating structured error messages for a specific struct
// It keeps track of field names, JSON names, slice fields, and custom error messages
// Useful for creating custom validation errors based on tags
type structErrorMessageCreator struct {
	jsonFieldNamesByFieldName map[string]string // Mapping from struct field names to JSON field names
	fieldNamesByJSONFieldName map[string]string // Mapping from JSON field names to struct field names
	SliceFields               map[string]bool   // Fields that are slices
	errorMsgs                 map[string]string // Custom error messages defined in struct tags
}

// camelToSpaceSeparated converts a camelCase or PascalCase string into a space-separated string
// Example: camelToSpaceSeparated("MyFieldName") -> "My Field Name"
func camelToSpaceSeparated(text string) string {
	re := regexp.MustCompile("([A-Z]+)([A-Z][a-z])|([a-z])([A-Z])")
	spaced := re.ReplaceAllString(text, "$1$3 $2$4")
	return spaced
}

// newStructErrorMessageCreator creates an instance of structErrorMessageCreator for a given type
// It analyzes the struct fields and populates necessary maps to facilitate error message creation
func newStructErrorMessageCreator(t reflect.Type) *structErrorMessageCreator {
	jsonFieldNames := make(map[string]string)
	fieldNames := make(map[string]string)
	sliceFields := make(map[string]bool)
	errorMsgs := make(map[string]string)

	fillInfo(t.Name()+".", t, jsonFieldNames, fieldNames, sliceFields, errorMsgs)

	return &structErrorMessageCreator{jsonFieldNames, fieldNames, sliceFields, errorMsgs}
}

// fillInfo recursively fills field info for a struct type into the provided maps
// This function handles nested struct fields and populates JSON field names, field names, slice fields, and custom
// error messages
func fillInfo(
	prefix string,
	t reflect.Type,
	jsonFieldNames map[string]string,
	fieldNames map[string]string,
	sliceFields map[string]bool,
	errorMsgs map[string]string,
) {
	if prefix == "." {
		prefix = ""
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			fillInfo(prefix+field.Name+".", field.Type, jsonFieldNames, fieldNames, sliceFields, errorMsgs)
			continue
		}
		fieldName := field.Tag.Get("json")
		if fieldName != "" {
			jsonFieldNames[prefix+field.Name] = fieldName
			fieldNames[fieldName] = field.Name
		} else {
			jsonFieldNames[prefix+field.Name] = field.Name
			fieldNames[field.Name] = field.Name
		}

		sliceFields[fieldName] = field.Type.Kind() == reflect.Slice

		errs := field.Tag.Get("errors")
		if errs != "" {
			sections := strings.Split(errs, "|")
			for _, section := range sections {
				parts := strings.Split(section, "=")
				if len(parts) == 2 {
					names := strings.Split(parts[0], ",")
					for _, name := range names {
						errorMsgs[prefix+field.Name+"."+name] = parts[1]
					}
				}
			}
		}
	}
}

// createErrorMessage creates an error message for a set of validation errors
// It uses default error messages and custom error messages defined in the struct tags
func (c *structErrorMessageCreator) createErrorMessage(validationErrors validator.ValidationErrors) error {
	var errorList = make(FieldValidationError)
	for _, ve := range validationErrors {
		fieldName := c.jsonFieldNamesByFieldName[ve.StructNamespace()]
		name := camelToSpaceSeparated(ve.StructField())
		message := name + " is invalid (" + ve.ActualTag() + ")"
		errMsg, found := c.errorMsgs[ve.StructNamespace()+"."+ve.ActualTag()]
		if !found {
			defaultMsg, defaultMsgFound := defaultErrorMessages[ve.ActualTag()]
			if defaultMsgFound {
				message = fmt.Sprintf(defaultMsg, name)
			}
		} else {
			message = errMsg
		}
		errorList[fieldName] = message
	}
	return &errorList
}

// createUnmarshalTypeErrorMessage creates an error message for JSON unmarshal type errors
// It returns an error with a helpful message indicating the expected data type
func (c *structErrorMessageCreator) createUnmarshalTypeErrorMessage(err *json.UnmarshalTypeError) error {
	var errorList = make(FieldValidationError)
	fieldName := c.fieldNamesByJSONFieldName[err.Field]
	name := camelToSpaceSeparated(fieldName)
	message := name + " should be a valid " + err.Type.Name()
	errorList[err.Field] = message
	return &errorList
}

// errorMessageCreator manages the creation of error messages for multiple struct types
// It ensures thread-safe access to its underlying map of struct error creators
type errorMessageCreator struct {
	structMap map[string]*structErrorMessageCreator
	mutex     sync.RWMutex
}

// newErrorMessageCreator creates an instance of errorMessageCreator
func newErrorMessageCreator() *errorMessageCreator {
	return &errorMessageCreator{make(map[string]*structErrorMessageCreator), sync.RWMutex{}}
}

// GetErrorMessageCreator retrieves or creates a structErrorMessageCreator for a given type
// It ensures that error message creators are cached for reuse
func (c *errorMessageCreator) GetErrorMessageCreator(t reflect.Type) *structErrorMessageCreator {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := t.Name()
	if v, ok := c.structMap[key]; ok {
		return v
	}
	creator := newStructErrorMessageCreator(t)
	c.structMap[key] = creator
	return creator
}

// GetUnmarshalTypeError creates an error message for JSON unmarshalling errors related to type mismatches
// If the error is an instance of json.UnmarshalTypeError, it retrieves a struct error message based on the provided
// input struct
func (c *errorMessageCreator) GetUnmarshalTypeError(inputStruct interface{}, err error) error {
	if err == nil {
		return nil
	}

	var validationErrors *json.UnmarshalTypeError
	if errors.As(err, &validationErrors) {
		val := reflect.Indirect(reflect.ValueOf(inputStruct))
		t := val.Type()

		message := c.GetErrorMessageCreator(t).createUnmarshalTypeErrorMessage(validationErrors)

		return message
	}

	return nil
}
