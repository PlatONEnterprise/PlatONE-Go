package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
)

const fileClearTime = 3600 * 24 * 7 // 7 Days

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

// PrintJson reformats the json printing style, easier for users to read
func PrintJson(marshalJson []byte) string {

	var addBytes = []byte{'\n'}
	var newJson = make([]byte, 0)

	for _, v := range marshalJson {
		switch v {
		case '}':
			addBytes = addBytes[:len(addBytes)-1]
			newJson = append(newJson, addBytes...)
			newJson = append(newJson, v)
		case '{':
			addBytes = append(addBytes, byte('\t'))
			newJson = append(newJson, v)
			newJson = append(newJson, addBytes...)
		case ',':
			newJson = append(newJson, v)
			newJson = append(newJson, addBytes...)
		default:
			newJson = append(newJson, v)
		}
	}

	return string(newJson)
}

// PrintRequest print the request to terminal or log for debugging usage
// it will limit the printing length if the request is too long
func PrintRequest(params interface{}) {
	paramJson, _ := json.Marshal(params)

	if len(paramJson) > 500 {
		fmt.Printf("\nrequest json data: %s... ...is too long\n", string(paramJson)[:500])
		/// path, _ := filepath.Abs(DEFAULT_LOG_DIRT)
		/// fmt.Printf("see full info at %s\n", path)
	} else {
		fmt.Printf("\nrequest json data: %s\n", string(paramJson))
	}
	/// Logger.Printf("request json data: %s\n", string(paramJson))
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

// FileDirectoryInit creates a new folder if the file directory is not exist
func FileDirectoryInit(filedirt string) {
	_, err := os.Stat(filedirt)
	if os.IsNotExist(err) {
		_ = os.Mkdir(filedirt, os.ModePerm)
	}
}

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

const (
	CnsIsName int32 = iota
	CnsIsAddress
	CnsIsUndefined
)

// TODO refactory
// IsNameOrAddress Judge whether the input string is an address or a name
func IsNameOrAddress(str string) int32 {
	var valid int32

	switch {
	case IsMatch(str, "address"):
		valid = CnsIsAddress
	case IsMatch(str, "name") &&
		!strings.HasPrefix(strings.ToLower(str), "0x"):
		valid = CnsIsName
	default:
		valid = CnsIsUndefined
	}

	return valid
}

// GetFileByKey search the file in the file directory by the search keywords provided
// if found, return the file name
func GetFileByKey(fileDirt, key string) string {

	fileList, _ := ioutil.ReadDir(fileDirt)

	for _, file := range fileList {

		switch {
		case file.IsDir():
			continue
		case strings.Contains(file.Name(), key):
			return file.Name()
		}
	}

	return ""
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

// TrimSpace trims all the space in the string
func TrimSpace(str string) string {
	strNoSpace := strings.Split(str, " ")
	return strings.Join(strNoSpace, "")
}

// FuncParse wraps the GetFuncNameAndParams
// it separates the function method name and the parameters
func FuncParse(funcName string, funcParams []string) (string, []string) {
	var funcParamsNew []string

	if funcName == "" {
		return "", nil
	}

	funcName, funcParamsNew = GetFuncNameAndParams(funcName)
	if len(funcParamsNew) != 0 && len(funcParams) != 0 {
		utils.Fatalf(ErrParamInValidSyntax, "function parameters")
	}
	funcParams = append(funcParams, funcParamsNew...)

	/// Logger.Printf("after function parse, the function is %s, %s", funcName, funcParams)
	return funcName, funcParams
}

// GetFuncNameAndParams parse the function params from the input string
func GetFuncNameAndParams(funcAndParams string) (string, []string) {
	// eliminate space
	f := TrimSpace(funcAndParams)

	hasBracket := strings.Contains(f, "(") && strings.Contains(f, ")")
	if !hasBracket {
		return f, nil
	}

	funcName := f[0:strings.Index(f, "(")]

	paramString := f[strings.Index(f, "(")+1 : strings.LastIndex(f, ")")]
	params := abi.GetFuncParams(paramString)

	return funcName, params
}

// deprecated, see abi/temp.go
func GetFuncParams(paramString string) []string {
	if paramString == "" {
		return nil
	}

	splitPos := recordFuncParamSplitPos(paramString)
	return splitFuncParamByPos(paramString, splitPos)

}

// splitFuncParamByPos splits the function params which is in string format
// by the position index recorded by the recordFuncParamSplitPos method
func splitFuncParamByPos(paramString string, splitPos []int) []string {
	var params = make([]string, 0)
	var lastPos = 0

	for _, i := range splitPos {
		params = append(params, paramString[lastPos:i])
		lastPos = i + 1
	}
	params = append(params, paramString[lastPos:])

	//params := strings.Split(paramString, ",")
	for index, param := range params {
		if strings.HasPrefix(param, "\"") {
			params[index] = param[strings.Index(param, "\"")+1 : strings.LastIndex(param, "\"")]
		}
		if strings.HasPrefix(param, "'") {
			params[index] = param[strings.Index(param, "'")+1 : strings.LastIndex(param, "'")]
		}
	}

	return params
}

// recordFuncParamSplitPos record the index of the end of each parameter
func recordFuncParamSplitPos(paramString string) []int {
	var symStack []rune
	var splitPos []int

	for i, s := range paramString {
		switch s {
		case ',':
			if len(symStack) == 0 {
				splitPos = append(splitPos, i)
			}
		case '{':
			symStack = append(symStack, '{')
		case '}':
			if len(symStack) < 1 {
				panic("parameter's format is not write!!!")
			}
			if symStack[len(symStack)-1] == '{' {
				symStack = symStack[:len(symStack)-1]
			}
		case '[':
			symStack = append(symStack, '[')
		case ']':
			if len(symStack) < 1 {
				panic("parameter's format is not write!!!")
			}
			if symStack[len(symStack)-1] == '[' {
				symStack = symStack[:len(symStack)-1]
			}
		case '(':
			symStack = append(symStack, '(')
		case ')':
			if len(symStack) < 1 {
				panic("parameter's format is not write!!!")
			}
			if symStack[len(symStack)-1] == '(' {
				symStack = symStack[:len(symStack)-1]
			}
		case '"':
			if len(symStack) < 1 {
				symStack = append(symStack, '"')
			} else {
				if symStack[len(symStack)-1] == '"' {
					symStack = symStack[:len(symStack)-1]
				} else {
					symStack = append(symStack, '"')
				}
			}
		case '\'':
			if len(symStack) < 1 {
				symStack = append(symStack, '\'')
			} else {
				if symStack[len(symStack)-1] == '\'' {
					symStack = symStack[:len(symStack)-1]
				} else {
					symStack = append(symStack, '\'')
				}
			}
		}
	}

	return splitPos
}
