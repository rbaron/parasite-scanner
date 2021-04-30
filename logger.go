package main

import (
	"log"
	"os"
)

var logger *log.Logger
var loggerIOWriter *os.File

// Logging will default to stdout. If a graphical user interface is used,
// logging will instead be written to this file.
const kLogFile = "parasite-scanner.log"

func InitLogger(shouldLogToFile bool) (err error) {
	ioWriter := os.Stdout
	if shouldLogToFile {
		ioWriter, err = os.OpenFile(kLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	logger = log.New(ioWriter, "", 0)
	return nil
}

func DeInitLogger() {
	if loggerIOWriter != os.Stdout {
		loggerIOWriter.Close()
	}
}
