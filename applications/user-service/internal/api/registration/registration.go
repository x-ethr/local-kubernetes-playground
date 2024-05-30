package registration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/x-ethr/server/handler/input"
	"github.com/x-ethr/server/handler/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"user-service/internal/database"
	"user-service/models/users"
)

var processor input.Processor[Body, users.CreateRow] = func(ctx context.Context, input *Body, output chan<- *users.CreateRow, exception chan<- *types.Exception, options *types.Options) {
	user := &users.CreateParams{Email: input.Email, Username: input.Username}

	connection, e := database.Connection(ctx)
	if e != nil {
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e}
		return
	}

	defer connection.Release()

	count, e := users.New(connection).Count(ctx, user.Email)
	if e != nil {
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e}
		return
	} else if count >= 1 {
		exception <- &types.Exception{Code: http.StatusConflict, Message: "Account With Email Address Already Exists"}
		return
	}

	var payload = map[string]string{"email": user.Email, "password": input.Password}

	content, e := json.Marshal(payload)
	if e != nil {
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to JSON Marshal Payload to Authorization Service", Source: e}
		return
	}

	var buffer bytes.Buffer
	if size, e := buffer.Write(content); e != nil || size == 0 {
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Write Payload Content to Buffer", Source: e, Metadata: map[string]interface{}{"size": size}}
		return
	}

	client := otelhttp.DefaultClient
	request, e := http.NewRequestWithContext(ctx, http.MethodPost, "http://authorization-service.development.svc.cluster.local:8080/users", &buffer)
	if e != nil {
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Create Request to Authorization Service", Source: e}
		return
	}

	response, e := client.Do(request)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Make Request to Authorization Service", slog.String("error", e.Error()))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Error Making Request to Authorization Service", Source: e}
		return
	} else if response.StatusCode >= 400 {
		slog.ErrorContext(ctx, "Authorization Service Returned Status-Code >= 400", slog.Int("status-code", response.StatusCode))

		content, e := io.ReadAll(response.Body)
		if e != nil {
			exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Read Error Response from Authorization Service", Source: e}
			return
		}

		slog.ErrorContext(ctx, "Authorization Service Error Message", slog.String("message", string(content)))

		exception <- &types.Exception{Code: response.StatusCode, Log: "Error Making Request to Authorization Service", Source: fmt.Errorf("response: %s", string(content))}
		return
	}

	defer response.Body.Close()

	var structure map[string]interface{}
	if e := json.NewDecoder(response.Body).Decode(&structure); e != nil {
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Error Decoding Authorization Service's Response Body", Source: e}
		return
	}

	slog.InfoContext(ctx, "Authorization Service", slog.Any("response", structure))

	result, e := users.New(connection).Create(ctx, user)
	if e != nil {
		exception <- &types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User"}
		return
	}

	output <- result

	return
}

func Handler(tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input.Process(w, r, v, processor, func(o *types.Options) {
			o.Tracer = tracer
		})

		return
	}
}
