package server

import (
	"fmt"
	"net/http"
	"store/KVStore"
	"store/logging"
	"strconv"
	"sync"
)

var (
	ConnHost string
	ConnPort string
)

var server *http.Server

var ShutdownChannel chan struct{}
var endpointWaitGroup *sync.WaitGroup

//Setup creates the server, initialises the KV store and registers all the endpoints
func Setup(port int, host string, storeBufferSize int, storeDepth int) error {
	ConnPort = ":" + strconv.Itoa(port)
	ConnHost = host

	//initialise the KV Store
	err := KVStore.Startup(storeBufferSize, storeDepth)
	if err != nil {
		return err
	}

	//initialise server
	server = &http.Server{
		Addr: ConnPort,
	}

	//initialise shutdown channel and endpoint waitgroup
	ShutdownChannel = make(chan struct{})
	endpointWaitGroup = &sync.WaitGroup{}

	//setup endpoints
	http.HandleFunc("/ping", PingEndpoint)
	http.HandleFunc("/shutdown", ShutdownEndpoint)
	http.HandleFunc("/store/", StoreEndpoint)
	http.HandleFunc("/list/", ListEndpoint)
	http.HandleFunc("/login", LoginEndpoint)
	return nil
}

//Start starts the server. Will block until the server is shutdown
func Start() error {
	//start server
	fmt.Println("Starting Server - see", ConnHost+ConnPort)
	logging.InfoLogger.Println("Starting server on port", ConnPort)
	err := server.ListenAndServe()
	return err
}
