package utils

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"log"
	"os"
	"path/filepath"
	"time"
)

var Logger *log.Logger
var LogErr *log.Logger
var LogDeg *log.Logger
var logFileDirt string

const (
	FILE_CLEAR_TIME = 3600 * 24 * 7 // 7 Days
	DEFAULT_LOG_DIRT    = "./log"
)

//TODO LogFileSetup
func init() {
	runPath := GetRunningTimePath()
	logFileDirt = runPath + DEFAULT_LOG_DIRT

	FileDirectoryInit(logFileDirt)
	pathSep := string(os.PathSeparator)
	logFilePath := logFileDirt + pathSep + time.Now().Format("2006-01-02") + ".log"

	// create or append
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		utils.Fatalf(ErrOpenFileFormat, "log", err.Error())
	}
	//defer logFile.Close()

	Logger = log.New(logFile, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	LogErr = log.New(logFile, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	LogDeg = log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	logStart := log.New(logFile, "", 0)
	logStart.Println("")

	DeleteOldFile(logFileDirt)
}

func DeleteOldFile(fileDirt string) {
	currentTime := time.Now().Unix()

	err := filepath.Walk(fileDirt, func(path string, fileInfo os.FileInfo, err error) error {

		if fileInfo == nil {
			return err
		}
		fileTime := fileInfo.ModTime().Unix()

		if (currentTime - fileTime) > FILE_CLEAR_TIME {
			_ = os.RemoveAll(path)
		}

		return nil
	})

	if err != nil {
		LogErr.Printf("Delete %s file error: %s\n", fileDirt, err.Error())
	}
}
