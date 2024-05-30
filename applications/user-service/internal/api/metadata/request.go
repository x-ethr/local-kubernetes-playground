package metadata

import (
	"github.com/go-playground/validator/v10"
	"github.com/x-ethr/server/handler/types"
)

// v represents the request body's struct validator
var v = validator.New(validator.WithRequiredStructEnabled())

// Body represents the handler's structured request-body
type Body struct {
	types.Helper

	Username *string `json:"username"`                        // Username represents the user's optional username.
	Email    string  `json:"email" validate:"required,email"` // Email represents the user's required email address.
	Password string  `json:"password" validate:"required"`    // Password represents the user's required password.
}

func (b *Body) Help() types.Validators {
	var mapping = types.Validators{
		"username": {
			Value:   b.Username,
			Valid:   true,
			Message: "(Optional) The username of the user.",
		},
		"email": {
			Value:   b.Email,
			Valid:   b.Email != "",
			Message: "(Required) A valid, unique email address.",
		},
		"password": {
			Valid:   len(b.Password) >= 8 && len(b.Password) <= 72,
			Message: "(Required) The user's password. Password must be between 8 and 72 characters in length.",
		},
	}

	return mapping
}
