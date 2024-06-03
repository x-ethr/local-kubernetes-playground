package registration

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/x-ethr/levels"
	"github.com/x-ethr/pg"
	"github.com/x-ethr/server"
	"github.com/x-ethr/server/cookies"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"github.com/x-ethr/server/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	tx, e := connection.Begin(ctx)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Begin a Database Transaction", Source: e})
		return
	}

	count, e := users.New().Count(ctx, tx, user.Email)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

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
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unknown Exception - Unable to Hash User's Password"})
		return
	}

	result, e := users.New().Create(ctx, tx, user)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User"})
		return
	}

	// --> create user in user-service

	client := otelhttp.DefaultClient

	var reader bytes.Buffer
	if e := json.NewEncoder(&reader).Encode(map[string]string{"email": result.Email}); e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User in User-Service"})
		return
	}

	request, e := http.NewRequestWithContext(ctx, http.MethodPost, "http://user-service.development.svc.cluster.local:8080/register", &reader)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User in User-Service"})
		return
	}

	response, e := client.Do(request)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User in User-Service"})
		return
	}

	defer response.Body.Close()

	content, e := io.ReadAll(response.Body)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Error Reading Response Body from User-Service"})
		return
	}

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated:
		var mapping map[string]interface{}
		if e := json.Unmarshal(content, &mapping); e != nil {
			pg.Disconnect(ctx, connection, tx)

			labeler.Add(attribute.Bool("error", true))
			x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Error Unmarshalling Response from User-Service"})
			return
		}

		slog.InfoContext(ctx, "User-Service Returned a Successful Response", slog.Any("value", mapping))
	case http.StatusConflict:
		slog.WarnContext(ctx, "User Already Exists in User-Service Database", slog.Bool("continue", true), slog.String("value", string(content)))
	default:
		pg.Disconnect(ctx, connection, tx)

		e := errors.New(response.Status)
		slog.ErrorContext(ctx, "User-Service Returned a Fatal Error", slog.Int("status", response.StatusCode), slog.String("response", string(content)))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Error Unmarshalling Response from User-Service"})
		return
	}

	jwt, e := token.Create(ctx, result.Email)
	if e != nil {
		pg.Disconnect(ctx, connection, tx)

		labeler.Add(attribute.Bool("error", true))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create JWT"})
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

	cookies.Secure(x.Writer(), "token", jwt)

	x.Complete(&types.Response{Status: http.StatusCreated, Payload: jwt})

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	server.Validate[Body](w, r, v, handle)

	return
})
