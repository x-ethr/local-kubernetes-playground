package avatar

import (
	"log/slog"
	"net/http"

	"github.com/x-ethr/levels"
	"github.com/x-ethr/pg"
	"github.com/x-ethr/server"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"github.com/x-ethr/server/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"user-service/internal/api/avatar/types/update"
	"user-service/models/users"
)

var patch server.Handle = func(x *types.CTX) {
	const name = "avatar-update"

	ctx := x.Request().Context()

	labeler := telemetry.Labeler(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	generic, e := x.Input()
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Message: http.StatusText(http.StatusInternalServerError), Log: "Validator Failed to Hydrate CTX Input"})
		return
	}

	var input = generic.(*update.Body)

	slog.DebugContext(ctx, "Input", slog.Any("request", input))

	dsn := pg.DSN()
	connection, e := pg.Connection(ctx, dsn)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e})
		return
	}

	tx, e := connection.Begin(ctx)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Begin a Database Transaction", Source: e})
		return
	}

	count, e := users.New().Count(ctx, tx, input.Email)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e})
		return
	} else if count == 0 {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusNotFound, Message: "Account With Email Address Not Found"})
		return
	}

	arguments := &users.UpdateUserAvatarParams{Email: input.Email, Avatar: &input.Avatar}
	if e := users.New().UpdateUserAvatar(ctx, tx, arguments); e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Update User's Avatar"})
		return
	}

	// --> commit the transaction
	if e := tx.Commit(ctx); e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Commit Transaction"})
		return
	}

	slog.Log(ctx, levels.Trace, "Successfully Committed Database Transaction")

	x.Complete(&types.Response{Status: http.StatusOK, Payload: arguments})

	return
}

var Patch = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	server.Validate[update.Body](w, r, update.V, patch)

	return
})
