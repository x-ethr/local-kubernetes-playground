package registration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"user-service/internal/database"
	"user-service/models/users"
)

func process(ctx context.Context, reader io.ReadCloser, body chan *Body, channel chan<- *users.CreateRow, exception chan<- *Exception, invalid chan<- *Invalid) {
	// defer close(body)
	// defer close(channel)
	// defer close(exception)
	// defer close(invalid)

	var data Body

	body <- &data

	if message, validators, e := Validate(ctx, v, reader, &data); e != nil {
		invalid <- &Invalid{Validators: validators, Message: message, Source: e}
		return
	}

	user := &users.CreateParams{Email: data.Email, Username: data.Username}
	// password := data.Password

	connection, e := database.Connection(ctx)
	if e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e}
		return
	}

	defer connection.Close(ctx)

	count, e := users.New(connection).Count(ctx, user.Email)
	if e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e}
		return
	} else if count >= 1 {
		exception <- &Exception{Code: http.StatusConflict, Message: "Account With Email Address Already Exists"}
		return
	}

	var payload = map[string]string{"email": user.Email, "password": data.Password}

	content, e := json.Marshal(payload)
	if e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Unable to JSON Marshal Payload to Authorization Service", Source: e}
		return
	}

	var buffer bytes.Buffer
	if size, e := buffer.Write(content); e != nil || size == 0 {
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Unable to Write Payload Content to Buffer", Source: e, Metadata: map[string]interface{}{"size": size}}
		return
	}

	client := otelhttp.DefaultClient
	request, e := http.NewRequestWithContext(ctx, http.MethodPost, "http://api.authorization-service.svc.cluster.local:8080/users", &buffer)
	if e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Unable to Create Request to Authorization Service", Source: e}
		return
	}

	response, e := client.Do(request)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Make Request to Authorization Service", slog.String("error", e.Error()))
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Error Making Request to Authorization Service", Source: e}
		return
	} else if response.StatusCode >= 400 {
		slog.ErrorContext(ctx, "Authorization Service Returned Status-Code >= 400", slog.Int("status-code", response.StatusCode))

		content, e := io.ReadAll(response.Body)
		if e != nil {
			exception <- &Exception{Code: http.StatusInternalServerError, Log: "Unable to Read Error Response from Authorization Service", Source: e}
			return
		}

		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Error Making Request to Authorization Service", Source: fmt.Errorf("response: %s", string(content))}
		return
	}

	defer response.Body.Close()

	var structure map[string]interface{}
	if e := json.NewDecoder(response.Body).Decode(&structure); e != nil {
		exception <- &Exception{Code: http.StatusInternalServerError, Log: "Error Decoding Authorization Service's Response Body", Source: e}
		return
	}

	slog.InfoContext(ctx, "Authorization Service", slog.Any("response", structure))

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

			slog.InfoContext(ctx, "Successfully Processed Request", slog.Any("response", response))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			json.NewEncoder(w).Encode(response)

			return
		case e := <-invalid:
			slog.WarnContext(ctx, "Invalid Request", slog.String("error", e.Error()), slog.String("path", r.URL.Path), slog.String("method", r.Method), slog.Any("input", input))
			e.Response(w)
			return
		case e := <-exception:
			slog.ErrorContext(ctx, "Error While Processing Request", slog.Any("error", e), slog.String("path", r.URL.Path), slog.String("method", r.Method), slog.Any("input", input))
			http.Error(w, e.Error(), e.Code)

			return
		}
	}
}
