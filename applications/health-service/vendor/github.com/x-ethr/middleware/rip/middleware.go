package rip

import (
	"context"
	"net/http"
)

type RIP struct {
	Remote string `json:"remote-address"`    // The original go's context remote address; see Real for the actual client's IP
	Real   string `json:"real-ip,omitempty"` // The real ip address of the client
}

type Implementation interface {
	Value(ctx context.Context) *RIP
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{}
}
