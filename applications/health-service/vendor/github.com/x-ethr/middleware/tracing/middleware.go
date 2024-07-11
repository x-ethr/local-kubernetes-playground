package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type Implementation interface {
	Value(ctx context.Context) trace.Tracer
	Configuration(options ...Variadic) Implementation
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{
		options: settings(),
	}
}
