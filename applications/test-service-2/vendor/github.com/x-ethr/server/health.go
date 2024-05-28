package server

import (
	"encoding/json"
	"net/http"
)

var Health http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "Application/JSON")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)

	return
}
