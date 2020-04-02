package core

import (
	"encoding/json"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	transfer       = 0
	deployContract = 1
	invokeContract = 2
	// vote           = 3
	// permission     = 4

	cnsTxType = 0x11 // Used for sending transactions without address
	fwTxType  = 0x12 // Used fot sending transactions  about firewall
	migTxType = 0x13 //Used for update system contract.

	DefaultConfigFilePath = "/config.json"

	DefaultAccountAdress = "0x0000000000000000000000000000000000000000"

	cnsProxyAddress = "0x0000000000000000000000000000000000000011"

	TxReceiptStatus_Success = "0x1"
	TxReceiptStatus_Failed = "0x0"

	DefaultConfigPath = "../../release/linux/conf/ctool.json"
	DefaultContractPath = "../../release/linux/conf/contracts/"
	DefaultPath = "../../release/linux/data/node-0/keystore"
)

var (
	config = Config{}
)

type JsonParam struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
}

type TxParams struct {
	From     common.Address `json:"from"`
	To       *common.Address `json:"to"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Value    string `json:"value"`
	Data     string `json:"data"`
	TxType   int    `json:"txType"`
}

type RawTxParams struct {
	TxParams
	Nonce int64 `json:"Nonce"`
}

type DeployParams struct {
	From     string `json:"from"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Data     string `json:"data"`
}

type Config struct {
	From     string `json:"from"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Url      string `json:"url"`
}

type FuncDesc struct {
	Name   string `json:"name"`
	Inputs []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"inputs"`
	Outputs []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"outputs"`
	Constant string `json:"constant"`
	Type     string `json:"type"`
}

type Response struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Id      int    `json:"id"`
	Error   struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Receipt struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  struct {
		BlockHash         string `json:"blockHash"`
		BlockNumber       string `json:"blockNumber"`
		ContractAddress   string `json:"contractAddress"`
		CumulativeGasUsed string `json:"cumulativeGas_used"`
		From              string `json:"from"`
		GasUsed           string `json:"gasUsed"`
		Root              string `json:"root"`
		To                string `json:"to"`
		TransactionHash   string `json:"transactionHash"`
		TransactionIndex  string `json:"transactionIndex"`
		Logs			  []struct{
			Address			string 		`json:"address"`
			Topics			[]string 	`json:"topics"`
			Data			string 		`json:"data"`
		} `json:"logs"`
		Status		  	string `json:"status"`
	} `json:"result"`
}

//----------------3.31---------------------------

type keystoreJson struct {
	Address string `json:"address""`
	Crypto  string `json:"crypto"`
}

/*
type logs struct {
	logs	[]*Log
}

type Log struct {
	Address			string `json"address"`
	Topics			[]string `json:"topics"`
	Data			[]byte `json:"data"`
}*/

func parseConfigJson(configPath string) error {
	if configPath == "" {
		dir, _ := os.Getwd()
		configPath = dir + DefaultConfigFilePath
	}

	if !filepath.IsAbs(configPath) {
		configPath, _ = filepath.Abs(configPath)
	}

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("parse config file error,%s", err.Error()))
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		panic(fmt.Errorf("parse config to json error,%s", err.Error()))
	}
	return nil
}

func parseAbiFromJson(fileName string) ([]FuncDesc, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("parse abi file error: %s", err.Error())
	}
	var a []FuncDesc
	if err := json.Unmarshal(bytes, &a); err != nil {
		return nil, fmt.Errorf("parse abi to json error: %s", err.Error())
	}
	return a, nil
}

func parseFuncFromAbi(fileName string, funcName string) (*FuncDesc, error) {
	funcs, err := parseAbiFromJson(fileName)
	if err != nil {
		return nil, err
	}

	for _, value := range funcs {
		if value.Name == funcName {
			return &value, nil
		}
	}
	return nil, fmt.Errorf("function %s not found in %s", funcName, fileName)
}

/**
  Find the method called by parsing abi
*/
func GetFuncNameAndParams(funcAndParams string) (string, []string) {
	strNoSpace := strings.Split(funcAndParams, " ")
	f := strings.Join(strNoSpace, "")

	funcName := string(f[0:strings.Index(f, "(")])

	paramString := string(f[strings.Index(f, "(")+1 : strings.LastIndex(f, ")")])
	if paramString == "" {
		return funcName, []string{}
	}

	symStack := []rune{}
	splitPos := []int{}

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
			if len(symStack) < 1{
				symStack = append(symStack, '\'')
			}else{
				if symStack[len(symStack) - 1] == '\''{
					symStack = symStack[:len(symStack) - 1]
				}else{
					symStack = append(symStack, '\'')
				}
			}
		}
	}
	params := make([]string,0)
	lastPos := 0
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
	return funcName, params
}

/**
  Self-test method for encrypting parameters
*/
func encodeParam(abiPath string, funcName string, funcParams string) error {
	abiFunc, err := parseFuncFromAbi(abiPath, funcName)
	if err != nil {
		return err
	}

	funcName, inputParams := GetFuncNameAndParams(funcParams)

	if len(abiFunc.Inputs) != len(inputParams) {
		return fmt.Errorf("incorrect number of parameters ,request=%d,get=%d\n", len(abiFunc.Inputs), len(inputParams))
	}

	paramArr := [][]byte{
		Int32ToBytes(111),
		[]byte(funcName),
	}

	for i, v := range inputParams {
		input := abiFunc.Inputs[i]
		p, e := StringConverter(v, input.Type)
		if e != nil {
			return err
		}
		paramArr = append(paramArr, p)
	}

	paramBytes, _ := rlp.EncodeToBytes(paramArr)

	fmt.Printf(hexutil.Encode(paramBytes))

	return nil
}


//---------------03.31 newly added--------------------------------

func parseFileToBytesDemo(filePath string) ([]byte, error) {

	fmt.Printf("In parseFileFunction\n")

	if filePath == ""{
		return nil, fmt.Errorf("file path cannot be null, %s")
	}

	if !filepath.IsAbs(filePath) {
		filePath, _ = filepath.Abs(filePath)
	}

	fmt.Printf("In IsAbs %s\n", filePath)

	file, err := os.Stat(filePath) //os.O_RDONLY, os.ModeDir

	if file.IsDir() {
		pathSep := string(os.PathSeparator)
		fileInfo, _ := ioutil.ReadDir(filePath)
		//f, _ := os.OpenFile(file.Name(), os.O_RDONLY, os.ModeDir)
		//fileInfo, _ := f.Readdir(-1)
		if fileInfo != nil {filePath = filePath + pathSep + fileInfo[0].Name()}
		fmt.Printf("In IsDir %s\n", filePath)
	}

	fmt.Printf("after IsDir %s\n", filePath)

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parse file error, %s", err.Error())
	}
	return bytes, nil
}
