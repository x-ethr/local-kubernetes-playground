package cors

import (
	"context"
	"net/http"
)

type Implementation interface {
	Value(ctx context.Context) bool
	Configuration(options ...Variadic) Implementation
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{
		options: settings(),
	}
}
