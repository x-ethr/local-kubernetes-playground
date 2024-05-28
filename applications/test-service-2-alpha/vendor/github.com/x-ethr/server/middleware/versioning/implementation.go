package versioning

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/x-ethr/text"

	"github.com/x-ethr/server/internal/keystore"
	"github.com/x-ethr/server/logging"
)

type generic struct {
	keystore.Valuer[string]

	options *Settings
}

func (*generic) Value(ctx context.Context) Version {
	if v, ok := ctx.Value(key).(Version); ok {
		return v
	}

	return Version{API: "N/A", Service: "N/A"}
}

func (g *generic) Configuration(options ...Variadic) Implementation {
	var o = settings()
	for _, option := range options {
		option(o)
	}

	g.options = o

	return g
}

func (g *generic) Middleware(next http.Handler) http.Handler {
	var name = text.Title(key.String(), func(o *text.Options) {
		o.Log = true
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := g.options.Version

			api, service := value.API, value.Service

			slog.Log(ctx, logging.Trace, "Middleware", slog.String("name", name), slog.Group("context", slog.String("key", string(key)), slog.Any("value", map[string]string{"api": api, "service": service})))

			ctx = context.WithValue(ctx, key, value)

			w.Header().Set("X-API-Version", api)
			w.Header().Set("X-Service-Version", service)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
