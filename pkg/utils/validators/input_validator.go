// Package validators provides a struct validation utility for generating error messages
package validators

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"reflect"
)

// InputValidator is a struct that helps in validating input structs and generating error messages
// It uses the go-playground/validator package for performing validations
// errorMessageCreator is used to create custom error messages based on validation tags
type InputValidator struct {
	validate            *validator.Validate
	ErrorMessageCreator *errorMessageCreator
}

// NewInputValidator creates a new instance of InputValidator
// It initializes the validator instance and the errorMessageCreator
func NewInputValidator() *InputValidator {
	return &InputValidator{
		validate:            NewValidate(),
		ErrorMessageCreator: newErrorMessageCreator(),
	}
}

// Validate validates the given value (struct) and returns an error if validation fails.
// If the value is not a struct, no validation is performed
// If validation fails, it returns a formatted error message based on the validation errors
func (iv *InputValidator) Validate(value any) error {
	indirectValue := reflect.Indirect(reflect.ValueOf(value))
	noValidation := indirectValue.Kind() != reflect.Struct
	if noValidation {
		return nil
	}
	err := iv.validate.Struct(value)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		t := indirectValue.Type()
		message := iv.ErrorMessageCreator.GetErrorMessageCreator(t).createErrorMessage(validationErrors)
		return message
	}
	return nil
}
