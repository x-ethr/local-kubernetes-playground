package telemetrics

import (
	"context"
	"net/http"
)

type Telemetry struct {
	Headers map[string]string // Headers represents the headers derived from an http(s) request relating to telemetry.
}

type Implementation interface {
	Value(ctx context.Context) Telemetry
	Configuration(options ...Variadic) Implementation
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{
		options: settings(),
	}
}
