package utils

import (
	"io/ioutil"
	"strings"
)

const(
	DEAFAULT_ABI_DIRT = "./abi"
)

func init(){
	DeleteOldFile(DEAFAULT_ABI_DIRT)
}

// StoreAbiFile stores the abi files in the default abi file directory
// if the abi file is already exist, the abi file will not be stored
func StoreAbiFile(key string, abiBytes []byte) {

	if len(abiBytes) == 0 {
		return
	}

	FileDirectoryInit(DEAFAULT_ABI_DIRT)
	fileName := GetFileByKey(DEAFAULT_ABI_DIRT, key)

	if fileName == ""{
		filePath := DEAFAULT_ABI_DIRT +  "/" + key + ".abi.json"
		WriteFile(abiBytes, filePath)
	}
}

// GetFileByKey search the file in the file directory by the search keywords provided
// if found, return the file name
func GetFileByKey(fileDirt, key string) string {

	fileList, _ := ioutil.ReadDir(fileDirt)

	for _, file := range fileList{

		switch {
		case file.IsDir():
			continue
		case strings.Contains(file.Name(), key):
			return file.Name()
		}
	}

	return ""
}
