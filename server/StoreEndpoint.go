package server

import (
	"io"
	"io/ioutil"
	"net/http"
	"store/KVStore"
	"store/logging"
	"strings"
)

func StoreEndpoint(w http.ResponseWriter, r *http.Request) {
	logging.LogAccessRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	//check for shutdown and either return or add self to waitgroup
	select {
	case <-ShutdownChannel:
		logging.WarningLogger.Println("attempted to access the store endpoint after a shutdown")
		w.WriteHeader(http.StatusNotFound)
		WriteWithError(w, "Server is shutting down", "store")
		return
	default:
		endpointWaitGroup.Add(1)
		defer endpointWaitGroup.Done()
	}

	//extract the key. Will always take the first argument after "/store/" as the key and ignore all others
	pathArgString := strings.TrimPrefix(r.URL.Path, "/store")
	trimmedPathArgString := strings.Trim(pathArgString, "/ ")
	pathArgs := strings.Split(trimmedPathArgString, "/")
	key := pathArgs[0]
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteWithError(w, "must provide a key in the url path", "store")
		return
	}

	//login with associated error handling
	username, errUsername := GetAuthorisation(r)
	if errUsername != nil {
		if errUsername == ErrInvalidAuth {
			w.WriteHeader(http.StatusForbidden)
			WriteWithError(w, "Forbidden", "store")
			return
		} else if errUsername == ErrUnauthorised {
			w.WriteHeader(http.StatusUnauthorized)
			WriteWithError(w, "Unauthorised", "store")
			return
		} else {
			logging.ErrorLogger.Println("unexpected error in authorisation", errUsername)
			w.WriteHeader(http.StatusInternalServerError)
			WriteWithError(w, "something has gone wrong", "store")
			return
		}
	}

	//interact with the KV store (actor modelling is handled by the KVStore package)
	var responseErr error
	var outputBody = "OK"
	switch r.Method {
	case http.MethodGet:
		value, err := KVStore.LookupValue(key, username)
		responseErr = err
		outputBody = value
	case http.MethodPut:
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logging.ErrorLogger.Println("Unable to close the request body")
			}
		}(r.Body)
		value, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logging.WarningLogger.Println("failed to read body of a put request", err)
			w.WriteHeader(http.StatusBadRequest)
			WriteWithError(w, "must provide a body", "store")
			return
		}
		responseErr = KVStore.PutValue(key, username, string(value))
	case http.MethodDelete:
		responseErr = KVStore.Delete(key, username)
	default:
		logging.WarningLogger.Println("received bad request on the store endpoint. Method was", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		WriteWithError(w, "invalid http method", "store")
		return
	}

	//handle any error returned from the KV store
	switch responseErr {
	case KVStore.ErrShutdown:
		logging.WarningLogger.Println("Server entered shutdown routine. Unable to Process request")
		w.WriteHeader(http.StatusNotFound)
		WriteWithError(w, "Server is shutting down", "store")
		return
	case KVStore.ErrKeyNotPresent:
		w.WriteHeader(http.StatusNotFound)
		WriteWithError(w, "404 key not found", "store")
		return
	case KVStore.ErrUnauthorized:
		w.WriteHeader(http.StatusForbidden)
		WriteWithError(w, "Forbidden", "store")
		return
	case nil:
		w.WriteHeader(http.StatusOK)
		WriteWithError(w, outputBody, "store")
		return
	default:
		logging.ErrorLogger.Println("unexpected error from the store interface", responseErr)
		w.WriteHeader(http.StatusInternalServerError)
		WriteWithError(w, "something went wrong", "store")
		return
	}
}
