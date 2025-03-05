package validators

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

// uuidv7Regex defines the regex pattern to validate a UUIDv7 (time-ordered UUID based on Unix time).
// It checks for a UUID that adheres to the expected UUIDv7 structure as per its specification.
var uuidv7Regex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// validateUUIDv7 is a custom validation function for validating UUIDv7 strings.
//
// Parameters:
//   - fl: validator.FieldLevel, contains information about the field to be validated.
//
// Returns:
//   - bool: returns true if the field value matches the UUIDv7 pattern, otherwise returns false.
func validateUUIDv7(fl validator.FieldLevel) bool {
	uuid, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return uuidv7Regex.MatchString(uuid)
}

// NewValidate creates and returns a new validator instance with custom validations registered.
// It registers the following custom validations and associated error messages:
//   - "uuidv7": Validates that a field is a valid UUIDv7.
//   - "oneof": Validates that a field value is one of the predefined valid values.
//
// Returns:
//   - *validator.Validate: a validator instance with custom validations registered.
func NewValidate() *validator.Validate {
	validate := validator.New()
	_ = validate.RegisterValidation("uuidv7", validateUUIDv7)

	// Register custom validation error messages
	registerValidationErrorMessage("uuidv7", "%s should be a valid UUID-V7")
	registerValidationErrorMessage("oneof", "%s should be one of valid enum values")

	return validate
}
