package logs

import (
	"context"
	"log/slog"
	"net/http"
)

type Implementation interface {
	Value(ctx context.Context) *slog.Logger
	Configuration(options ...Variadic) Implementation
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{
		options: settings(),
	}
}
