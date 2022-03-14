package server

import (
	"context"
	"fmt"
	"net/http"
	"store/KVStore"
	"store/logging"
	"time"
)

const timeoutTime = 10 * time.Second

func shutdownRoutine() {
	chanCloseErr := SafeClose(ShutdownChannel) //this could be replaced by a sync.once
	if chanCloseErr != nil {                   //this implies a shutdown has already begun
		logging.WarningLogger.Println("attempted to close the shutdown channel more than once")
		return
	}

	err := KVStore.Shutdown() //this will block until the store channel has been drained
	if err != nil {
		logging.ErrorLogger.Println("Error shutting down the KV store", err)
	}

	WaitWithTimeout(endpointWaitGroup, timeoutTime) //include timeout here in case one of the endpoints has crashed

	ctx, _ := context.WithTimeout(context.Background(), timeoutTime)
	errServer := server.Shutdown(ctx) //will attempt to shut down the server gracefully, but includes a timeout in case something's gone wrong
	if errServer != nil {
		logging.ErrorLogger.Println("unable to shut down server", errServer)
	}
}

func ShutdownEndpoint(w http.ResponseWriter, r *http.Request) {
	logging.LogAccessRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	//check for shutdown and either return or add self to waitgroup
	select {
	case <-ShutdownChannel:
		logging.WarningLogger.Println("attempted to access the shutdown endpoint after a shutdown")
		w.WriteHeader(http.StatusNotFound)
		WriteWithError(w, "Server is shutting down", "shutdown")
		return
	default:
		endpointWaitGroup.Add(1)
		defer endpointWaitGroup.Done()
	}

	//login with associated error handling
	username, errUsername := GetAuthorisation(r)
	if errUsername != nil {
		if errUsername == ErrInvalidAuth {
			w.WriteHeader(http.StatusForbidden)
			WriteWithError(w, "Forbidden", "shutdown")
			return
		} else if errUsername == ErrUnauthorised {
			w.WriteHeader(http.StatusUnauthorized)
			WriteWithError(w, "Unauthorised", "shutdown")
			return
		} else {
			logging.ErrorLogger.Println("unexpected error in authorisation", errUsername)
			w.WriteHeader(http.StatusInternalServerError)
			WriteWithError(w, "something has gone wrong", "shutdown")
			return
		}
	}

	if username != "admin" {
		logging.WarningLogger.Println("tried to shutdown without admin privileges")
		w.WriteHeader(http.StatusForbidden)
		WriteWithError(w, "Forbidden", "shutdown")
		return
	}

	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		WriteWithError(w, "OK", "shutdown")
		logging.InfoLogger.Println("Starting shutdown routine")
		fmt.Println("Shutting down server")
		go shutdownRoutine()
	} else {
		logging.WarningLogger.Println("attempted to access shutdown endpoint with method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		WriteWithError(w, "invalid http method", "shutdown")
		return
	}
}
