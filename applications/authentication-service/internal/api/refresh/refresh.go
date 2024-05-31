package refresh

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/x-ethr/server/cookies"
	"github.com/x-ethr/server/handler"
	"github.com/x-ethr/server/handler/types"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/token"
)

func processor(x *types.CTX) {
	const name = "refresh"

	ctx := x.Request().Context()

	labeler := telemetry.Labeler(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	var c bool // was cookie found?

	var tokenstring = ""

	cookie, e := x.Request().Cookie("token")
	if e != nil {
		c = false

		authorization := x.Request().Header.Get("Authorization")
		if authorization == "" && errors.Is(e, http.ErrNoCookie) {
			labeler.Add(attribute.Bool("error", true))
			x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Valid Authorization, Cookie Not Found", Source: e})
			return
		} else if authorization == "" {
			e = fmt.Errorf("null authorization header and unknown cookie error: %w", e)
			labeler.Add(attribute.Bool("error", true))
			x.Error(&types.Exception{Code: http.StatusBadRequest, Message: "Null Authorization Header & Unknown Cookie Error", Source: e})
			return
		}

		partials := strings.Split(authorization, " ")
		if len(partials) != 2 || partials[0] != "Bearer" {
			e := fmt.Errorf("invalid authorization header format: %s", strings.Join(partials, " "))
			labeler.Add(attribute.Bool("error", true))
			x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Invalid Authorization Format", Source: e})
			return
		}

		tokenstring = partials[1]
	} else {
		c = true
		span.AddEvent("cookie-refresh-request")
		tokenstring = cookie.Value
	}

	slog.DebugContext(ctx, "JWT Token", tokenstring)

	jwttoken, e := token.Verify(ctx, tokenstring)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))

		switch {
		case errors.Is(e, jwt.ErrTokenMalformed):
			span.RecordError(e, trace.WithStackTrace(false))
			x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Malformed JWT Token", Source: e})
			return
		case errors.Is(e, jwt.ErrTokenSignatureInvalid):
			span.RecordError(e, trace.WithStackTrace(false))
			x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Invalid JWT Token Signature", Source: e})
			return
		case errors.Is(e, jwt.ErrTokenExpired):
			span.RecordError(e, trace.WithStackTrace(false))
			x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Expired JWT Token", Source: e})
			return
		case errors.Is(e, jwt.ErrTokenNotValidYet):
			span.RecordError(e, trace.WithStackTrace(false))
			x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Invalid Future JWT Token", Source: e})
			return
		default:
			span.RecordError(e, trace.WithStackTrace(true))
			x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Unhandled JWT Error", Source: e})
			return
		}
	}

	slog.InfoContext(ctx, "Token Structure", slog.Any("jwt-header(s)", jwttoken.Header), slog.Any("jwt-claims", jwttoken.Claims))
	if !(c) { // --> regardless of outcome, because client doesn't have cookie but we have a jwttoken, set it.
		cookies.Secure(x.Writer(), "token", tokenstring)
	}

	expiration, e := jwttoken.Claims.GetExpirationTime()
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Error While Attempting to Get JWT Expiration Time", Source: e})
		return
	}

	remaining := time.Until(time.Unix(expiration.Unix(), 0))
	if remaining > 15*time.Minute {
		labeler.Add(attribute.Bool("error", true))

		x.Writer().Header().Set("Retry-After", strconv.Itoa(int(remaining.Seconds())))
		x.Writer().Header().Set("X-Retry-After-Unit", "Seconds")
		x.Error(&types.Exception{Code: http.StatusTooManyRequests, Message: "Invalid Validity Duration Remaining", Source: e})
		return
	}

	email, e := jwttoken.Claims.GetSubject()
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusUnauthorized, Message: "Error Attempting to Get JWT Subject", Source: e})
		return
	}

	update, e := token.Create(ctx, email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Message: "Error Attempting to Generate JWT Token", Source: e})
		return
	}

	cookies.Secure(x.Writer(), "token", update)

	x.Complete(&types.Response{Code: http.StatusOK, Payload: update})

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handler.Process(w, r, processor)

	return
})
