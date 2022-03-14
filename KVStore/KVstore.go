package KVStore

import (
	"errors"
	"time"
)

var (
	ErrKeyNotPresent = errors.New("key not present")
	ErrUnauthorized  = errors.New("user is not authorised to view this key")
	ErrBadRequest    = errors.New("bad store request")
	ErrShutdown      = errors.New("the KV store is shutting down")
)

var kvStore map[string]*Data

var (
	MaxDepth   int
	BufferSize int
)

const StewardTimeout = 10 * time.Second //may want to make this an argument of the startup function

var (
	//StoreChannel is the channel used to communicate with the actor
	StoreChannel chan StoreRequest
	//ShutdownChannel will unblock after a shutdown has been initiated
	ShutdownChannel chan struct{}
	//StoreGuardianDoneChan will only unblock after the store guardian has been shut down
	StoreGuardianDoneChan chan struct{}
)

const adminUser = "admin"

const (
	LookupString   = "lookup"
	PutString      = "put"
	DeleteString   = "delete"
	ListString     = "list"
	ShutdownString = "shutdown"
)

func Startup(bufferSize int, depth int) error {
	BufferSize = bufferSize
	StoreChannel = make(chan StoreRequest, BufferSize)
	kvStore = map[string]*Data{}
	MaxDepth = depth
	ShutdownChannel = make(chan struct{})
	StoreGuardianDoneChan = ListenForStoreRequests(StewardTimeout)
	return nil //no scope for errors currently, but may happen in the future iterations
}

func Shutdown() error {
	request := StoreRequest{command: ShutdownString}
	select {
	case <-ShutdownChannel: //Already initiated shutdown
		return ErrShutdown
	default:
		close(ShutdownChannel)
	}
	//closing the store guardian by sending a message instead of closing the store channel
	//This will prevent a panic if another endpoint tries to send a request before shutdown completes
	StoreChannel <- request //Tell store guardian to initiate shutdown
	<-StoreGuardianDoneChan //Wait for guardian to receive message and initiate shutdown
	close(StoreChannel)
	//could add a wait group in here to wait for all processes to receive and handle their results,
	//but probably unnecessary and could lead to deadlock if one of the processes dies
	return nil
}

func LookupValue(key, user string) (string, error) {
	request := StoreRequest{command: LookupString, data: StoreData{key: key, user: user}}
	response := MakeRequest(request)
	return response.value, response.err
}

func PutValue(key, user, value string) error {
	request := StoreRequest{command: PutString, data: StoreData{key: key, user: user, value: value}}
	response := MakeRequest(request)
	return response.err
}

func Delete(key, user string) error {
	request := StoreRequest{command: DeleteString, data: StoreData{key: key, user: user}}
	response := MakeRequest(request)
	return response.err
}

func ListStore() ([]byte, error) {
	request := StoreRequest{command: ListString, data: StoreData{}}
	response := MakeRequest(request)
	return response.json, response.err
}

func ListKey(key string) ([]byte, error) {
	request := StoreRequest{command: ListString, data: StoreData{key: key}}
	response := MakeRequest(request)
	return response.json, response.err
}
