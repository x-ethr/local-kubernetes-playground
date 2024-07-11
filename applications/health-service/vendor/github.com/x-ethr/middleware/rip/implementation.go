package rip

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/x-ethr/middleware/types"
)

var trueClientIP = http.CanonicalHeaderKey("True-Client-IP")
var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

func evaluate(r *http.Request) string {
	var ip string

	if tcip := r.Header.Get(trueClientIP); tcip != "" {
		ip = tcip
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ",")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	}

	if ip == "" || net.ParseIP(ip) == nil {
		return ""
	}

	return ip
}

type generic struct {
	types.Valuer[*RIP]
}

func (*generic) Value(ctx context.Context) *RIP {
	return ctx.Value(key).(*RIP)
}

func (*generic) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			var ip string

			if rip := evaluate(r); rip != "" {
				// r.RemoteAddr = rip
				ip = rip
			}

			value := &RIP{
				Remote: r.RemoteAddr,
				Real:   ip,
			}

			slog.Log(ctx, (slog.LevelDebug - 4), "Middleware", slog.Group("context", slog.String("key", string(key)), slog.Any("value", value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
