package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// FormatValidationError converts raw binding and validation errors into clean, human-readable messages.
func FormatValidationError(err error) string {
	if err == nil {
		return ""
	}

	// Handle empty body
	if errors.Is(err, io.EOF) {
		return "request body cannot be empty"
	}

	// Handle JSON unmarshal type error
	var unmarshalTypeError *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeError) {
		return fmt.Sprintf("invalid type for field '%s', expected %s", unmarshalTypeError.Field, unmarshalTypeError.Type.String())
	}

	// Handle JSON syntax error
	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return "malformed JSON payload"
	}

	// Handle go-playground validation errors
	var valErrors validator.ValidationErrors
	if errors.As(err, &valErrors) {
		var errorMessages []string
		for _, e := range valErrors {
			fieldName := toLowerCamel(e.Field())
			switch e.Tag() {
			case "required":
				errorMessages = append(errorMessages, fmt.Sprintf("%s is required", fieldName))
			case "email":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid email address", fieldName))
			case "min":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least %s characters long", fieldName, e.Param()))
			case "max":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must not exceed %s characters", fieldName, e.Param()))
			default:
				errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation on '%s'", fieldName, e.Tag()))
			}
		}
		if len(errorMessages) > 0 {
			return strings.Join(errorMessages, ", ")
		}
	}

	return err.Error()
}

func toLowerCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
