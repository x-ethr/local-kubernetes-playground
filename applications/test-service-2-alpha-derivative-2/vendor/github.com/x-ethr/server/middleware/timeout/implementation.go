package timeout

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/x-ethr/text"

	"github.com/x-ethr/server/internal/keystore"
	"github.com/x-ethr/server/logging"
)

type generic struct {
	keystore.Valuer[string]

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

func (*generic) Value(ctx context.Context) string {
	return ctx.Value(key).(string)
}

func (g *generic) Middleware(next http.Handler) http.Handler {
	var name = text.Title(key.String(), func(o *text.Options) {
		o.Log = true
	})

	if g.options.Timeout <= 0 {
		g.options.Timeout = (time.Second * 30)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := strconv.FormatInt(int64(g.options.Timeout), 10)

			slog.Log(ctx, logging.Trace, "Middleware", slog.String("name", name), slog.Group("context", slog.String("key", string(key)), slog.String("value", g.options.Timeout.String())))

			ctx = context.WithValue(ctx, key, value)

			w.Header().Set("X-Timeout", g.options.Timeout.String())
		}

		ctx, cancel := context.WithTimeout(ctx, g.options.Timeout)
		defer func() {
			cancel()
			e := ctx.Err()
			if errors.Is(e, context.DeadlineExceeded) {
				http.Error(w, "gateway-timeout", http.StatusGatewayTimeout)
				return
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
