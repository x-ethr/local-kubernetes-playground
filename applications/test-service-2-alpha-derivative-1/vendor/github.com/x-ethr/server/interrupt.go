package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// Interrupt is a graceful interrupt + signal handler for an HTTP server.
func Interrupt(ctx context.Context, cancel context.CancelFunc, server *http.Server) {
	// Listen for syscall signals for process to interrupt/quit
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-interrupt

		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			fmt.Print("\r")
		}

		slog.DebugContext(ctx, "Initializing Server Shutdown ...")

		// Shutdown signal with grace period of 30 seconds
		shutdown, timeout := context.WithTimeout(ctx, 30*time.Second)
		defer timeout()
		go func() {
			<-shutdown.Done()
			if errors.Is(shutdown.Err(), context.DeadlineExceeded) {
				slog.Log(ctx, slog.LevelError, "Graceful Server Shutdown Timeout - Forcing an Exit ...")

				os.Exit(99)
			}
		}()

		// Trigger graceful shutdown
		if e := server.Shutdown(shutdown); e != nil {
			slog.ErrorContext(ctx, "Exception During Server Shutdown", slog.String("error", e.Error()))
		}

		cancel()
	}()
}
