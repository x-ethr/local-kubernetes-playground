package login

import (
	"github.com/go-playground/validator/v10"
	"github.com/x-ethr/server/types"
)

// Body represents the handler's structured request-body
type Body struct {
	types.Helper `json:"-"`
	Email        string `json:"email" validate:"required,email"`           // Email represents the user's required email address.
	Password     string `json:"password" validate:"required,min=8,max=72"` // Password represents the user's required password.
}

func (b *Body) Help() types.Validators {
	var mapping = types.Validators{
		"email": {
			Value:   b.Email,
			Valid:   b.Email != "",
			Message: "(Required) A valid, unique email address.",
		},
		"password": {
			Valid:   len(b.Password) >= 8 && len(b.Password) <= 72,
			Message: "(Required) The user's password.",
		},
	}

	return mapping
}

// v represents the request body's struct validator
var v = validator.New(validator.WithRequiredStructEnabled())
