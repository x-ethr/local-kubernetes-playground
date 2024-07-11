package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Server initializes a http.Server.
func Server(ctx context.Context, handler http.Handler, port string) *http.Server {
	return &http.Server{
		Addr:                         fmt.Sprintf("0.0.0.0:%s", port),
		Handler:                      handler,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  15 * time.Second,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 60 * time.Second,
		IdleTimeout:                  30 * time.Second,
		MaxHeaderBytes:               http.DefaultMaxHeaderBytes,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		ConnContext: nil,
	}
}
