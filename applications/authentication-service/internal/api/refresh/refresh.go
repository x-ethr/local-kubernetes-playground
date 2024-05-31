package login

import (
	"fmt"
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
	const name = "login"

	ctx := r.Context()

	tracing := middleware.New().Tracer().Value(ctx)
	version := middleware.New().Version().Value(ctx)
	service := middleware.New().Service().Value(ctx)

	ctx, span := tracing.Start(ctx, name, trace.WithAttributes(attribute.String("component", name)), trace.WithAttributes(telemetry.Resources(ctx, service, version.Service).Attributes()...))

	dsn := pg.DSN()
	connection, e := pg.Connection(ctx, dsn)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e}
		return
	}

	defer connection.Release()

	count, e := users.New().Count(ctx, connection, input.Email)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e}
		return
	} else if count == 0 {
		exception <- &types.Exception{Code: http.StatusNotFound, Message: "User Not Found", Source: fmt.Errorf("user not found: %s", input.Email)}
		return
	}

	user, e := users.New().Get(ctx, connection, input.Email)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Retrieve User Record", Source: e}
		return
	}

	if e := users.Verify(user.Password, input.Password); e != nil {
		span.RecordError(e, trace.WithStackTrace(false))
		exception <- &types.Exception{Code: http.StatusUnauthorized, Log: "Invalid Authentication Attempt", Source: e}
		return
	}

	jwt, e := token.Create(ctx, user.Email)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(false))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create JWT"}
		return
	}

	output <- &types.Response{Code: http.StatusOK, Payload: jwt}

	return
}

func Handler(w http.ResponseWriter, r *http.Request) {
	input.Process(w, r, v, processor)

	return
}
