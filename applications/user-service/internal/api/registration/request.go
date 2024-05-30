package registration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"

	"github.com/go-playground/validator/v10"
)

// Body represents the handler's structured request-body
type Body struct {
	Username *string `json:"username"`                        // Username represents the user's optional username.
	Email    string  `json:"email" validate:"required,email"` // Email represents the user's required email address.
	Password string  `json:"password" validate:"required"`    // Password represents the user's required password.
}

func (b *Body) Help() map[string]Validator {
	var mapping = make(map[string]Validator, 3)

	mapping["username"] = Validator{
		Value:   b.Username,
		Valid:   true,
		Message: "(Optional) The username of the user.",
	}

	mapping["email"] = Validator{
		Value:   b.Email,
		Valid:   b.Email != "",
		Message: "(Required) A valid, unique email address.",
	}

	mapping["password"] = Validator{
		Valid:   len(b.Password) >= 8 && len(b.Password) <= 72,
		Message: "(Required) The user's password. Password must be between 8 and 72 characters in length.",
	}

	return mapping
}

// v represents the request body's struct validator
var v atomic.Pointer[validator.Validate]

type Validator struct {
	Value   interface{} `json:"value,omitempty"`
	Valid   bool        `json:"valid"`
	Message string      `json:"message"`
}

type Helper interface {
	Help() map[string]Validator
}

func Validate(ctx context.Context, body io.Reader, data Helper) (string, map[string]Validator, error) {
	if v.Load() == nil {
		v.Store(validator.New(validator.WithRequiredStructEnabled()))
	}

	// invalid describes an invalid argument passed to `Struct`, `StructExcept`, StructPartial` or `Field`
	var invalid *validator.InvalidValidationError

	// Unmarshal request-body into "data".
	if e := json.NewDecoder(body).Decode(&data); e != nil {
		// Log an issue unmarshalling the body and return a Bad request exception.
		slog.Log(ctx, slog.LevelWarn, "Unable to Unmarshal Request Body",
			slog.String("error", e.Error()),
		)

		return "Valid JSON Required as Input", nil, e
	}

	// Validate "data" using "validation".
	if e := v.Load().Struct(data); e != nil {
		// Check if the error is due to an invalid validation configuration.
		if errors.As(e, &invalid) {
			// Log the issue and return an Internal server error exception.
			slog.ErrorContext(ctx, "Invalid Validator", slog.String("error", e.Error()))

			return "Internal Validation Error", nil, e
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

		return "", data.Help(), e
	}

	// If no exception was generated, log the "data" for debugging purposes.
	// slog.Log(ctx, slog.LevelDebug, "Request", logging.Structure("body", data))

	// Return nil if there were no exceptions generated.
	return "", nil, nil
}
