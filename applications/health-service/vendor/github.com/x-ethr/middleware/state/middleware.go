package state

import (
	"context"
	"net/http"
)

type Implementation interface {
	Value(ctx context.Context) string
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{}
}
