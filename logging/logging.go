package logging

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/logging"
)

const logName = "kv-store-log"

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	AccessLogger  *log.Logger
)

var (
	serverLogFile *os.File
	accessLogFile *os.File

	client *logging.Client
)

func LogAccessRequest(r *http.Request) {
	AccessLogger.Printf("url: %s, HTTP method: %s, Source IP address: %s\n", r.Host+r.URL.Path, r.Method, r.RemoteAddr)
}

func SetupLoggers(serverLogName string, accessLogName string, cloudLogs bool) {
	if cloudLogs {
		setupCloudLoggers()
	} else {
		setupLocalLoggers(serverLogName, accessLogName)
	}
}

func setupLocalLoggers(serverLogName string, accessLogName string) {
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

func setupCloudLoggers() {
	projectID, ok := os.LookupEnv("GOOGLE_CLOUD_PROJECT")
	if !ok {
		log.Fatal("GOOGLE_CLOUD_PROJECT not set")
		return
	}
	ctx := context.Background()
	var err error
	client, err = logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return
	}
	InfoLogger = client.Logger(logName).StandardLogger(logging.Info)
	WarningLogger = client.Logger(logName).StandardLogger(logging.Warning)
	ErrorLogger = client.Logger(logName).StandardLogger(logging.Error)

	AccessLogger = client.Logger("kv-store-access-log").StandardLogger(logging.Info)
}

func Shutdown() {
	fmt.Println("Shutting down loggers")
	if serverLogFile != nil {
		fmt.Println("closing serverlogfile")
		serverLogFile.Close()
	}
	if accessLogFile != nil {
		fmt.Println("Closing accessLogFile")
		accessLogFile.Close()
	}
	if client != nil {
		fmt.Println("Closing cloud logging client")
		client.Close()
	}
}
