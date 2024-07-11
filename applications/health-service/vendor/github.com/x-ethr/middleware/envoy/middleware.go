package envoy

import (
	"context"
	"net/http"
)

type Implementation interface {
	Value(ctx context.Context) *Envoy
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{}
}
