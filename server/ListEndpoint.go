package server

import (
	"net/http"
	"store/KVStore"
	"store/logging"
	"strings"
)

func ListEndpoint(w http.ResponseWriter, r *http.Request) {
	logging.LogAccessRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	//check for shutdown and either return or add self to waitgroup
	select {
	case <-ShutdownChannel:
		logging.WarningLogger.Println("attempted to access the list endpoint after a shutdown")
		w.WriteHeader(http.StatusNotFound)
		WriteWithError(w, "Server is shutting down", "list")
		return
	default:
		endpointWaitGroup.Add(1)
		defer endpointWaitGroup.Done()
	}

	if r.Method != http.MethodGet {
		logging.WarningLogger.Println("attempted to access list endpoint with method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		WriteWithError(w, "invalid http method", "list")
		return
	}

	//extract the key. Will always take the first argument after "/list/" as the key and ignore all others
	pathArgString := strings.TrimPrefix(r.URL.Path, "/list")
	trimmedPathArgString := strings.Trim(pathArgString, "/ ")
	pathArgs := strings.Split(trimmedPathArgString, "/")
	key := pathArgs[0] //will be "" if no key provided

	//login with associated error handling
	//currently don't need the username in the list endpoint, so just checks that they are a valid user
	_, errUsername := GetAuthorisation(r)
	if errUsername != nil {
		if errUsername == ErrInvalidAuth {
			w.WriteHeader(http.StatusForbidden)
			WriteWithError(w, "Forbidden", "list")
			return
		} else if errUsername == ErrUnauthorised {
			w.WriteHeader(http.StatusUnauthorized)
			WriteWithError(w, "Unauthorised", "list")
			return
		} else {
			logging.ErrorLogger.Println("unexpected error in authorisation", errUsername)
			w.WriteHeader(http.StatusInternalServerError)
			WriteWithError(w, "something has gone wrong", "list")
			return
		}
	}

	//interact with the KV store (actor modelling is handled by the KVStore package)
	var storeResponse []byte
	var storeErr error
	if key == "" {
		storeResponse, storeErr = KVStore.ListStore()
	} else {
		storeResponse, storeErr = KVStore.ListKey(key)
	}

	//handle any error returned from the KV store
	switch storeErr {
	case KVStore.ErrShutdown:
		logging.WarningLogger.Println("Server entered shutdown routine. Unable to Process request")
		w.WriteHeader(http.StatusNotFound)
		WriteWithError(w, "Server is shutting down", "list")
		return
	case KVStore.ErrKeyNotPresent:
		w.WriteHeader(http.StatusNotFound)
		WriteWithError(w, "404 key not found", "list")
		return
	case KVStore.ErrUnauthorized:
		w.WriteHeader(http.StatusForbidden)
		WriteWithError(w, "Forbidden", "list")
		return
	case nil:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(storeResponse)
		if err != nil {
			logging.ErrorLogger.Println("error writing in the list endpoint.", err)
			return
		}
		return
	default:
		logging.ErrorLogger.Println("unexpected error from the list interface", storeErr)
		w.WriteHeader(http.StatusInternalServerError)
		WriteWithError(w, "something went wrong", "list")
		return
	}
}
