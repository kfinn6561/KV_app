package server

import (
	"net/http"
	"store/logging"
)

func PingEndpoint(w http.ResponseWriter, r *http.Request) {
	logging.LogAccessRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		WriteWithError(w, "pong", "ping")
		return
	} else {
		logging.WarningLogger.Println("attempted to access ping endpoint with method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		WriteWithError(w, "invalid http method", "ping")
		return
	}
}
