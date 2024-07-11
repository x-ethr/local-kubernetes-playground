package authentication

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type Authentication struct {
	Token *jwt.Token

	Audience string // Audience is the target recipient the JWT is intended for, as set by the "aud" jwt-claims structure.
	Issuer   string // Issuer is the issuing service that generated the jwt, as set by the "iss" jwt-claims structure.
	Email    string // Email represents the user's email address as set by the "sub" jwt-claims structure.
	Raw      string // Raw represents the raw jwt token as submitted by the client.

	// ID interface{} // ID represents a custom identifier field, as set by the "id" non-standard jwt-claims structure. Defaults to nil.
}

type Implementation interface {
	Value(ctx context.Context) *Authentication
	Configuration(options ...Variadic) Implementation
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{
		options: settings(),
	}
}
