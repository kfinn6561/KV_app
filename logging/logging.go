package logging

import (
	"log"
	"net/http"
	"os"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	AccessLogger  *log.Logger
)

var (
	serverLogFile *os.File
	accessLogFile *os.File
)

func LogAccessRequest(r *http.Request) {
	AccessLogger.Printf("url: %s, HTTP method: %s, Source IP address: %s\n", r.Host+r.URL.Path, r.Method, r.RemoteAddr)
}

func SetupLoggers(serverLogName string, accessLogName string) {
	var err error
	serverLogFile, err = os.OpenFile(serverLogName,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644)
	if err != nil {
		log.Fatal(err)
	}

	accessLogFile, err = os.OpenFile(accessLogName,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644)
	if err != nil {
		log.Fatal(err)
	}
	InfoLogger = log.New(serverLogFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(serverLogFile, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(serverLogFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	AccessLogger = log.New(accessLogFile, "INFO: ", log.Ldate|log.Ltime)
}

func Shutdown() {
	serverLogFile.Close()
	accessLogFile.Close()
}
