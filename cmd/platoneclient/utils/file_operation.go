package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// get the path where the executable is executed
func GetRunningTimePath() string {
	cur, _ := os.Executable()

	path := os.Args[0]
	if cur == path {
		return ""
	} else {
		index := strings.Index(path, "/")
		path = path[index+1:]
		return cur[:len(cur)-len(path)]
	}
}

// FileDirectoryInit creates a new folder if the file directory is not exist
func FileDirectoryInit(filedirt string) {
	_, err := os.Stat(filedirt)
	if os.IsNotExist(err) {
		_ = os.Mkdir(filedirt, os.ModePerm)
	}
}

// GetFileInDirt get the first file at the file directory
func GetFileInDirt(fileDirt string) (string, error) {

	var filePath string

	file, err := os.Stat(fileDirt)
	if err != nil {
		return "", err
	}

	if file.IsDir() {
		pathSep := string(os.PathSeparator)
		fileInfo, _ := ioutil.ReadDir(fileDirt)
		if fileInfo != nil {
			filePath = fileDirt + pathSep + fileInfo[0].Name()
		}
	} else {
		filePath = fileDirt
	}

	return filePath, nil
}

const fileClearTime = 3600 * 24 * 7 // 7 Days

func DeleteOldFile(fileDirt string) error {
	currentTime := time.Now().Unix()

	return filepath.Walk(fileDirt, func(path string, fileInfo os.FileInfo, err error) error {

		if fileInfo == nil {
			return err
		}
		fileTime := fileInfo.ModTime().Unix()

		if (currentTime - fileTime) > fileClearTime {
			_ = os.RemoveAll(path)
		}

		return nil
	})
}

// GetFileByKey search the file in the file directory by the search keywords provided
// if found, return the file name
func GetFileByKey(fileDirt, key string) (string, error) {

	if key == "" {
		return "", errors.New("the key can not be null")
	}

	fileList, err := ioutil.ReadDir(fileDirt)
	if err != nil {
		return "", err
	}

	for _, file := range fileList {
		if file.IsDir() {
			continue
		}

		if strings.Contains(strings.ToLower(file.Name()), strings.ToLower(key)) {
			return file.Name(), nil
		}
	}

	return "", errors.New("file not found")
}

// WriteFile writes the new data in the file, the old data will be override
func WriteFile(fileBytes []byte, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		// utils.Fatalf(ErrOpenFileFormat, filePath, err.Error())
		return err
	}

	_, err = file.Write(fileBytes)
	if err != nil {
		// utils.Fatalf(ErrWriteFileFormat, err.Error())
		return err
	}

	return nil
}

// ParseFileToBytes read the file and return the file bytes
func ParseFileToBytes(filePath string) ([]byte, error) {

	if filePath == "" {
		return nil, ErrFileNull
	}

	if !filepath.IsAbs(filePath) {
		filePath, _ = filepath.Abs(filePath)
	}

	_, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf(ErrFindFileFormat, err.Error())
	}
	/// Logger.Printf("the file being parsed: %s\n", filePath)

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf(ErrReadFileFormat, filePath, err.Error())
	}
	return bytes, nil
}
