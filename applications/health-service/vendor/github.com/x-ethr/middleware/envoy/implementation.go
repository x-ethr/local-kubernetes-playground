package envoy

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/x-ethr/middleware/types"
)

type generic struct {
	types.Valuer[string]
}

type Envoy struct {
	Attempts *int    `json:"x-envoy-attempt-count,omitempty"`
	Original *string `json:"x-envoy-original-path,omitempty"`
	Internal *bool   `json:"x-envoy-internal,omitempty"`
}

func (*generic) Value(ctx context.Context) *Envoy {
	if v, ok := ctx.Value(key).(*Envoy); ok {
		return v
	}

	return nil
}

func (*generic) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			var internal *bool
			if v := r.Header.Get("X-Envoy-Internal"); v == "true" {
				assignment := true
				internal = &assignment
			}

			var attempts *int
			if v := r.Header.Get("X-Envoy-Request-Count"); v != "" {
				assignment, e := strconv.Atoi(v)
				if e == nil {
					attempts = &assignment
				}
			}

			var original *string
			if v := r.Header.Get("X-Envoy-Original-Path"); v != "" {
				original = &v
			}

			value := &Envoy{Original: original, Attempts: attempts, Internal: internal}

			slog.Log(ctx, (slog.LevelDebug - 4), "Middleware", slog.Group("context", slog.String("key", string(key)), slog.Any("value", value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
