package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

// Validators is a type that represents a map of string keys to Validator values.
// Each key-value pair in the map corresponds to a validation check for a specific field.
// The string key is the field name, and the Validator value contains information about the validation result.
type Validators map[string]Validator

// Validator is a type that represents a validation result for a specific field.
// It contains information about the validated value, validity, and an optional message.
//
//   - The [Validator.Value] field stores the value that was validated.
//   - The [Validator.Valid] field indicates whether the validation check was successful or not.
//   - The [Validator.Message] field holds an optional message providing additional information about the validation result.
type Validator struct {
	Value   interface{} `json:"value,omitempty"` // Value is the value that was validated.
	Valid   bool        `json:"valid"`           // Valid is a boolean field indicating whether the validation check was successful or not.
	Message string      `json:"message"`         // Message is a field in the Validator struct that holds an optional message providing additional information about the validation result.
}

// Helper is an interface that defines a single method, Help().
// Help() returns a map of string keys to Validator values, representing validation checks for specific fields.
type Helper interface {
	Help() Validators // Help is a method of the Helper interface that returns a map of string keys to Validator values.
}

// Validate is a function that takes a context, validator, request body reader, and data interface as arguments.
// It performs the following steps:
// 1. Unmarshals the request body into the data interface.
// 2. Validates the data using the validator.
// 3. If there are validation errors, logs each error and returns an appropriate response.
// 4. If the data implements the Helper interface, returns the result of the Help method.
// 5. Logs the data for debugging purposes.
// 6. Returns nil if there were no exceptions generated.
// The function returns a string message, a map of Validators, and an error.
func Validate(ctx context.Context, v *validator.Validate, body io.Reader, data interface{}) (Validators, error) {
	// invalid describes an invalid argument passed to `Struct`, `StructExcept`, StructPartial` or `Field`
	var invalid *validator.InvalidValidationError

	// Unmarshal request-body into "data".
	if e := json.NewDecoder(body).Decode(&data); e != nil {
		if typecast, ok := data.(Helper); ok {
			return typecast.Help(), e
		}

		// Log an issue unmarshalling the body and return a Bad request exception.
		slog.Log(ctx, slog.LevelWarn, "Unable to Unmarshal Request Body",
			slog.String("error", e.Error()),
		)

		return nil, e
	}

	// Validate "data" using "validation".
	if e := v.Struct(data); e != nil {
		// Check if the error is due to an invalid validation configuration.
		if errors.As(e, &invalid) {
			// Log the issue and return an Internal server error exception.
			slog.ErrorContext(ctx, "Invalid Validator", slog.String("error", e.Error()))

			return nil, e
		}

		// Loop through the validation errors, logging each one.
		for key, e := range e.(validator.ValidationErrors) {
			slog.Log(ctx, slog.LevelWarn, fmt.Sprintf("Validator (%d)", key), slog.Group("error",
				slog.String("tag", e.Tag()),
				slog.String("actual-tag", e.ActualTag()),
				slog.String("parameter", e.Param()),
				slog.String("field", e.Field()),
				slog.String("namespace", e.Namespace()),
				slog.String("struct-namespace", e.StructNamespace()),
				slog.String("struct-field", e.StructField()),
				slog.Group("reflection",
					slog.Attr{
						Key:   "kind",
						Value: slog.AnyValue(e.Kind()),
					},
					slog.Attr{
						Key:   "type",
						Value: slog.AnyValue(e.Type()),
					},
					slog.Attr{
						Key:   "value",
						Value: slog.AnyValue(e.Value()),
					},
				),
			))
		}

		if typecast, ok := data.(Helper); ok {
			return typecast.Help(), e
		}

		return nil, e
	}

	// If no exception was generated, log the "data" for debugging purposes.
	// slog.Log(ctx, slog.LevelDebug, "Request", logging.Structure("body", data))

	// Return nil if there were no exceptions generated.
	return nil, nil
}
