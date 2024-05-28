package servername

import "github.com/x-ethr/server/internal/keystore"

type Settings struct {

	// Server represents the "Server" [http.Header].
	Server string `json:"server" yaml:"server"`
}

type Variadic keystore.Variadic[Settings]

func settings() *Settings {
	return &Settings{}
}
