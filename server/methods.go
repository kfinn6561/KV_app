package server

import (
	"errors"
	"net/http"
	"store/logging"
	"store/users"
	"strings"
	"sync"
	"time"
)

var ErrInvalidAuth = errors.New("invalid authorisation token")
var ErrUnauthorised = errors.New("unauthorised")

//logging.ErrorLogger.Println("store guardian received a bad request command", storeRequest.command)

func GetAuthorisation(r *http.Request) (username string, err error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		logging.ErrorLogger.Println("token not in required format", reqToken)
		return "", ErrInvalidAuth
	}
	tokenString := strings.TrimSpace(splitToken[1])
	username, ok := users.ValidateJWT(tokenString)
	if !ok {
		return "", ErrUnauthorised
	} else {
		return username, nil
	}
}

func WriteWithError(w http.ResponseWriter, value string, endpointName string) {
	_, err := w.Write([]byte(value))
	if err != nil {
		logging.ErrorLogger.Printf("error writing in the %s endpoint. %v\n", endpointName, err)
	}
}

func WaitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()
	select {
	case <-waitChan:
		return true
	case <-time.After(timeout):
		return false
	}
}

//SafeClose normally closes a channel and returns nil.
//But if the channel is already closed it will recover from the panic and return an error.
func SafeClose(channel chan struct{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("channel is closed")
		}
	}()
	close(channel)
	return nil
}
