package main

import (
	"log"
	"os"
	"path"
)

var (
	// default error logger
	logger = log.New(os.Stderr, "", 0)

	// remote events logger
	RemoteLoggerFile = "goneypot.log"
	remoteLogger     *log.Logger

	// credentials logger
	DisableCredsLog = false
	CredsLoggerFile = "credentials.log"
	credsLogger     *log.Logger
)

func setupLoggers() {
	setupRemoteLogger()

	if !DisableCredsLog {
		setupCredsLogger()
	}
}

func setupRemoteLogger() {
	p := path.Join(LoggingRoot, RemoteLoggerFile)
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatal("error opening file: ", err)
	}
	remoteLogger = log.New(f, "", log.Ldate|log.Ltime)
	logger.Printf("set up remote logging to: %s", p)
}

func setupCredsLogger() {
	p := path.Join(LoggingRoot, CredsLoggerFile)
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatal("error opening file: ", err)
	}
	credsLogger = log.New(f, "", 0)
	logger.Printf("set up credential logging to: %s", p)
}

func logRemoteEvent(remoteAddr string, message string) {
	remoteLogger.Printf("%s %s", remoteAddr, message)
}
