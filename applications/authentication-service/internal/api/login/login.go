package login

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/x-ethr/pg"
	"github.com/x-ethr/server/cookies"
	"github.com/x-ethr/server/handler/input"
	"github.com/x-ethr/server/handler/types"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/token"
	"authentication-service/models/users"
)

func processor(w http.ResponseWriter, r *http.Request, input *Body, output chan<- *types.Response, exception chan<- *types.Exception, options *types.Options) {
	const name = "login"

	ctx := r.Context()

	labeler := telemetry.Labeler(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	cookie, e := r.Cookie("token")
	if e == nil {
		jwttoken, e := token.Verify(ctx, cookie.Value)
		if e == nil && jwttoken.Valid {
			if email, ok := jwttoken.Claims.(jwt.MapClaims)["sub"].(string); ok {
				span.AddEvent("authenticated-established-user-login-attempt", trace.WithAttributes(attribute.String("email", email)))
			}

			labeler.Add(attribute.Bool("error", true))
			exception <- &types.Exception{Code: http.StatusBadRequest, Message: "Authenticated Session Already Exists for User", Log: "Authentication User Attempted to Login"}
			return
		}
	}

	dsn := pg.DSN()
	connection, e := pg.Connection(ctx, dsn)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e}
		return
	}

	defer connection.Release()

	count, e := users.New().Count(ctx, connection, input.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e}
		return
	} else if count == 0 {
		exception <- &types.Exception{Code: http.StatusNotFound, Message: "User Not Found", Source: fmt.Errorf("user not found: %s", input.Email)}
		return
	}

	user, e := users.New().Get(ctx, connection, input.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Retrieve User Record", Source: e}
		return
	}

	if e := users.Verify(user.Password, input.Password); e != nil {
		labeler.Add(attribute.Bool("error", true))
		exception <- &types.Exception{Code: http.StatusUnauthorized, Log: "Invalid Authentication Attempt", Source: e}
		return
	}

	jwt, e := token.Create(ctx, user.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create JWT"}
		return
	}

	cookies.Secure(w, "token", jwt)

	output <- &types.Response{Code: http.StatusOK, Payload: jwt}

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	input.Process(w, r, v, processor)

	return
})
