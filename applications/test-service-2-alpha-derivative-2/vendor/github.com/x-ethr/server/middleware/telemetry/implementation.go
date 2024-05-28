package telemetry

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/x-ethr/text"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/x-ethr/server/internal/keystore"
	"github.com/x-ethr/server/logging"
)

type generic struct {
	keystore.Valuer[string]
}

func (generic) Value(ctx context.Context) string {
	return ctx.Value(key).(string)
}

func (generic) Middleware(next http.Handler) http.Handler {
	var name = text.Title(key.String(), func(o *text.Options) {
		o.Log = true
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		server := ctx.Value(keystore.Keys().Server()).(string)

		// --> benefit of interfaces includes avoiding cyclic dependencies.
		mux := ctx.Value(http.ServerContextKey).(*http.Server).Handler.(interface {
			Pattern(r *http.Request) string
		})

		pattern := mux.Pattern(r)

		{
			value := "enabled"

			slog.Log(ctx, logging.Trace, "Middleware", slog.String("name", name), slog.Group("context", slog.String("key", string(key)), slog.Any("value", map[string]string{"enabled": value, "pattern": pattern})))

			ctx = context.WithValue(ctx, key, value)
		}

		handler := otelhttp.NewHandler(otelhttp.WithRouteTag(pattern, next), pattern, otelhttp.WithServerName(server), otelhttp.WithFilter(func(request *http.Request) (filter bool) {
			ctx := request.Context()

			if request.URL.Path == "/health" {
				filter = true

				slog.Log(ctx, logging.Trace, "Health Telemetry Exclusion", slog.Bool("filter", filter))
			}

			return
		}))

		handler.ServeHTTP(w, r)
	})
}
