package utils

import (
	"log"
	"os"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
)

var Logger *log.Logger
var LogErr *log.Logger
var LogDeg *log.Logger
var logFileDirt string

const (
	defaultLogDirt = "./platonecli_log"
)

// LogInit is used while debugging utils, packet packages
func LogInit() {
	runPath := GetRunningTimePath()
	logFileDirt = runPath + defaultLogDirt

	FileDirectoryInit(logFileDirt)
	pathSep := string(os.PathSeparator)
	logFilePath := logFileDirt + pathSep + time.Now().Format("2006-01-02") + ".log"

	// create or append
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		utils.Fatalf(ErrOpenFileFormat, "log", err.Error())
	}

	Logger = log.New(logFile, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	LogErr = log.New(logFile, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	LogDeg = log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	logStart := log.New(logFile, "", 0)
	logStart.Println("[Record of cmd]")

	err = DeleteOldFile(logFileDirt)
	if err != nil {
		LogErr.Printf("Delete %s file error: %s\n", logFileDirt, err.Error())
	}
}
