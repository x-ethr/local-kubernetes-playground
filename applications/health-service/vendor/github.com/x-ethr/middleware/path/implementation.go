package path

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/x-ethr/middleware/types"
)

type generic struct {
	types.Valuer[string]
}

func (*generic) Value(ctx context.Context) string {
	return ctx.Value(key).(string)
}

func (*generic) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := r.URL.Path

			slog.Log(ctx, (slog.LevelDebug - 4), "Middleware", slog.Group("context", slog.String("key", string(key)), slog.String("value", value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
