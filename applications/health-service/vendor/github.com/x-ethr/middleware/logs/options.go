package logs

import (
	"log/slog"

	"github.com/x-ethr/middleware/types"
)

type Settings struct {
	Logger *slog.Logger
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{
		Logger: slog.Default(),
	}
}
