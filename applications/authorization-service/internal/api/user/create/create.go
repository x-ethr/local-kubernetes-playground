package create

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"authorization-service/internal/database"
	"authorization-service/models/users"
)

func process(ctx context.Context, reader io.ReadCloser, body chan *Body, channel chan<- *users.CreateRow, exception chan<- *Exception, invalid chan<- *Invalid) {
	// defer close(body)
	// defer close(channel)
	// defer close(exception)
	// defer close(invalid)

	var data Body

	body <- &data

	if message, validators, e := Validate(ctx, reader, &data); e != nil {
		invalid <- &Invalid{Validators: validators, Message: message, Source: e}
		return
	}

	user := &users.CreateParams{Email: data.Email}

	connection, e := database.Connection(ctx)
	if e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e}
		return
	}

	defer connection.Release()

	count, e := users.New(connection).Count(ctx, user.Email)
	if e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e}
		return
	} else if count >= 1 {
		exception <- &Exception{Code: http.StatusConflict, Message: "Account With Email Address Already Exists"}
		return
	}

	password := data.Password
	user.Password, e = users.Hash(password)
	if e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unknown Exception - Unable to Hash User's Password"}
		return
	}

	result, e := users.New(connection).Create(ctx, user)
	if e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User"}
		return
	}

	channel <- result

	return
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := make(chan *Body, 1)                                                                            // -> buffered, asynchronous channel(s)
	channel, exception, invalid := make(chan *users.CreateRow), make(chan *Exception), make(chan *Invalid) // -> unbuffered, synchronous channel(s)

	var input *Body // only used for logging

	go process(ctx, r.Body, body, channel, exception, invalid)

	for {
		select {
		case <-ctx.Done():
			return
		case input = <-body: // continue waiting for one of the other primitives to complete
			continue
		case response := <-channel:
			if response == nil {
				slog.ErrorContext(ctx, "Response Returned Unexpected, Null Result", slog.String("path", r.URL.Path), slog.String("method", r.Method), slog.Any("input", input))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			slog.DebugContext(ctx, "Successfully Processed Request", slog.Any("response", response))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			json.NewEncoder(w).Encode(response)

			return
		case e := <-invalid:
			slog.WarnContext(ctx, "Invalid Request", slog.String("error", e.Error()), slog.String("path", r.URL.Path), slog.String("method", r.Method), slog.Any("input", input))
			e.Response(w)
			return
		case e := <-exception:
			var err error = e.Source
			if e.Source == nil {
				err = fmt.Errorf("N/A")
			}

			slog.ErrorContext(ctx, "Error While Processing Request", slog.String("error", err.Error()), slog.String("internal-message", e.Log), slog.String("path", r.URL.Path), slog.String("method", r.Method), slog.Any("input", input))
			http.Error(w, e.Error(), e.Code)

			return
		}
	}
}
