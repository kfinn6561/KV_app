package server

import (
	"net/http"
	"store/logging"
	"store/users"
)

func LoginEndpoint(w http.ResponseWriter, r *http.Request) {
	logging.LogAccessRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != http.MethodGet {
		logging.WarningLogger.Println("attempted to access login endpoint with method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		WriteWithError(w, "invalid http method", "login")
		return
	}
	user, password, ok := r.BasicAuth()
	if !ok {
		logging.WarningLogger.Println("tried to login without providing basic auth")
		w.WriteHeader(http.StatusBadRequest)
		WriteWithError(w, "Must provide basic auth", "login")
		return
	}
	token, err := users.GenerateJWT(user, password) //this includes a password check
	if err != nil {
		logging.WarningLogger.Println("attempt to login with invalid details")
		w.WriteHeader(http.StatusUnauthorized)
		WriteWithError(w, "Unauthorised", "login")
		return
	}
	// should I add it as a cookie?
	logging.InfoLogger.Printf("user %s successfully logged in\n", user)
	output := "Bearer " + token
	w.WriteHeader(http.StatusOK)
	WriteWithError(w, output, "login")
}
