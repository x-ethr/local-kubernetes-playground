package cors

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/rs/cors"

	"github.com/x-ethr/middleware/types"
)

type generic struct {
	types.Valuer[string]

	options *Settings
}

func (g *generic) Configuration(options ...Variadic) Implementation {
	var o = settings()
	for _, option := range options {
		option(o)
	}

	g.options = o

	return g
}

func (*generic) Value(ctx context.Context) bool {
	return ctx.Value(key).(bool)
}

func (g *generic) Middleware(next http.Handler) http.Handler {
	wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := true

			slog.Log(ctx, (slog.LevelDebug - 4), "Middleware", slog.Group("context", slog.String("key", string(key)), slog.Bool("value", value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})

	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool { return true },
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
		Debug:            g.options.Debug,
	})

	return c.Handler(wrapper)
}
