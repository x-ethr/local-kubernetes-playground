package registration

import (
	"net/http"

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
	const name = "registration"

	ctx := x.Request().Context()

	labeler := telemetry.Labeler(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	cookie, e := x.Request().Cookie("token")
	if e == nil {
		jwttoken, e := token.Verify(ctx, cookie.Value)
		if e == nil && jwttoken.Valid {
			labeler.Add(attribute.Bool("error", true))
			x.Error(&types.Exception{Code: http.StatusBadRequest, Message: "Authenticated Session Already Exists for User", Log: "Authentication User Attempted to Register"})
			return
		}
	}

	generic, e := x.Input()
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Message: http.StatusText(http.StatusInternalServerError), Log: "Validator Failed to Hydrate CTX Input"})
		return
	}

	var input = generic.(*Body)

	user := &users.CreateParams{Email: input.Email}

	dsn := pg.DSN()
	connection, e := pg.Connection(ctx, dsn)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e})
		return
	}

	defer connection.Release()

	count, e := users.New().Count(ctx, connection, user.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e})
		return
	} else if count >= 1 {
		x.Error(&types.Exception{Code: http.StatusConflict, Message: "Account With Email Address Already Exists"})
		return
	}

	password := input.Password
	user.Password, e = users.Hash(password)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unknown Exception - Unable to Hash User's Password"})
		return
	}

	result, e := users.New().Create(ctx, connection, user)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User"})
		return
	}

	jwt, e := token.Create(ctx, result.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create JWT"})
		return
	}

	var payload = map[string]interface{}{
		"token": jwt,
	}

	cookies.Secure(x.Writer(), "token", jwt)

	x.Complete(&types.Response{Status: http.StatusCreated, Payload: payload})

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	server.Validate[Body](w, r, v, handle)

	return
})
