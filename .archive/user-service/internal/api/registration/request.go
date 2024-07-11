package registration

import (
	"github.com/go-playground/validator/v10"
	"github.com/x-ethr/server/types"
)

// Body represents the handler's structured request-body
type Body struct {
	types.Helper `json:"-"`
	Email        string  `json:"email" validate:"required,email"` // Email represents the user's required email address.
	Avatar       *string `json:"avatar"`
}

func (b *Body) Help() types.Validators {
	var mapping = types.Validators{
		"email": {
			Value:   b.Email,
			Valid:   b.Email != "",
			Message: "(Required) A valid, unique email address.",
		},
		"avatar": {
			Value:   b.Avatar,
			Valid:   true,
			Message: "(Optional) An avatar URL.",
		},
	}

	return mapping
}

// v represents the request body's struct validator
var v = validator.New(validator.WithRequiredStructEnabled())
