package registration

import (
	"net/http"

	"github.com/x-ethr/pg"
	"github.com/x-ethr/server/handler/input"
	"github.com/x-ethr/server/handler/types"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/token"
	"authentication-service/models/users"
)

func processor(r *http.Request, input *Body, output chan<- *types.Response, exception chan<- *types.Exception, options *types.Options) {
	const name = "registration"

	ctx := r.Context()

	tracing := middleware.New().Tracer().Value(ctx)
	version := middleware.New().Version().Value(ctx)
	service := middleware.New().Service().Value(ctx)

	ctx, span := tracing.Start(ctx, name, trace.WithAttributes(attribute.String("component", name)), trace.WithAttributes(telemetry.Resources(ctx, service, version.Service).Attributes()...))

	user := &users.CreateParams{Email: input.Email}

	dsn := pg.DSN()
	connection, e := pg.Connection(ctx, dsn)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e}
		return
	}

	defer connection.Release()

	count, e := users.New(connection).Count(ctx, user.Email)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e}
		return
	} else if count >= 1 {
		exception <- &types.Exception{Code: http.StatusConflict, Message: "Account With Email Address Already Exists"}
		return
	}

	password := input.Password
	user.Password, e = users.Hash(password)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unknown Exception - Unable to Hash User's Password"}
		return
	}

	result, e := users.New(connection).Create(ctx, user)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User"}
		return
	}

	jwt, e := token.Create(ctx, result.Email)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(false))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create JWT"}
		return
	}

	var payload = map[string]interface{}{
		"token": jwt,
	}

	output <- &types.Response{Code: http.StatusCreated, Payload: payload}

	return
}

func Handler(w http.ResponseWriter, r *http.Request) {
	input.Process(w, r, v, processor)

	return
}
