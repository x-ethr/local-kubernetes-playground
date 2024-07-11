# `http-server` - HTTP Routing, Logging & Telemetry

## Documentation

Official `godoc` documentation (with examples) can be found at the [Package Registry](https://pkg.go.dev/github.com/x-ethr/go-http-server/v2).

- See https://pkg.go.dev/github.com/x-ethr/go-http-server/v2 for additional documentation.
- Please refer to the [code examples](./example_test.go) for additional usage and implementation details.

## Usage

###### Add Package Dependency

```bash
go get -u github.com/x-ethr/go-http-server/v2
```

###### Import & Implement

`main.go`

```go
package main

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "os"

    "github.com/x-ethr/go-http-server/v2"
    "github.com/x-ethr/go-http-server/v2/writer"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())

    mux := http.NewServeMux()

    mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        var response = map[string]interface{}{
            "key": "value",
        }

        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })

    // Add Response Writer
    handler := writer.Handle(mux)

    // Start the HTTP server
    slog.Info("Starting Server ...", slog.String("local", fmt.Sprintf("http://localhost:%s", "8080")))

    api := server.Server(ctx, handler, "8080")

    // Issue Cancellation Handler
    server.Interrupt(ctx, cancel, api)

    // <-- Blocking
    if e := api.ListenAndServe(); e != nil && !(errors.Is(e, http.ErrServerClosed)) {
        slog.ErrorContext(ctx, "Error During Server's Listen & Serve Call ...", slog.String("error", e.Error()))

        os.Exit(100)
    }

    // --> Exit
    {
        slog.InfoContext(ctx, "Graceful Shutdown Complete")

        // Waiter
        <-ctx.Done()
    }
}
```

## Contributions

See the [**Contributing Guide**](./CONTRIBUTING.md) for additional details on getting started.
