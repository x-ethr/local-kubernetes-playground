package login

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/x-ethr/pg"
	"github.com/x-ethr/server"
	"github.com/x-ethr/server/cookies"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"github.com/x-ethr/server/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/token"
	"authentication-service/models/users"
)

var handle server.Handle = func(x *types.CTX) {
	const name = "login"

	ctx := x.Request().Context()

	labeler := telemetry.Labeler(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	logger := slog.With(slog.String("name", name))

	generic, e := x.Input()
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Message: http.StatusText(http.StatusInternalServerError), Log: "Validator Failed to Hydrate CTX Input"})
		return
	}

	var input = generic.(*Body)

	logger.InfoContext(ctx, "Input", slog.Any("body", input))

	cookie, e := x.Request().Cookie("token")
	if e == nil {
		jwttoken, e := token.Verify(ctx, cookie.Value)
		if e == nil && jwttoken.Valid {
			if email, ok := jwttoken.Claims.(jwt.MapClaims)["sub"].(string); ok {
				span.AddEvent("authenticated-established-user-login-attempt", trace.WithAttributes(attribute.String("email", email)))
			}

			labeler.Add(attribute.Bool("error", true))
			x.Error(&types.Exception{Code: http.StatusBadRequest, Message: "Authenticated Session Already Exists for User", Log: "Authentication User Attempted to Login"})
			return
		}
	}

	dsn := pg.DSN()
	connection, e := pg.Connection(ctx, dsn)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e})
		return
	}

	defer connection.Release()

	count, e := users.New().Count(ctx, connection, input.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e})
		return
	} else if count == 0 {
		x.Error(&types.Exception{Code: http.StatusNotFound, Message: "User Not Found", Source: fmt.Errorf("user not found: %s", input.Email)})
		return
	}

	user, e := users.New().Get(ctx, connection, input.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Retrieve User Record", Source: e})
		return
	}

	if e := users.Verify(user.Password, input.Password); e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusUnauthorized, Log: "Invalid Authentication Attempt", Source: e})
		return
	}

	jwtstring, e := token.Create(ctx, user.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create JWT"})
		return
	}

	cookies.Secure(x.Writer(), "token", jwtstring)

	logger.DebugContext(ctx, "Successfully Generated JWT", slog.String("jwt", jwtstring))

	x.Complete(&types.Response{Status: http.StatusOK, Payload: jwtstring})

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	server.Validate[Body](w, r, v, handle)

	return
})
