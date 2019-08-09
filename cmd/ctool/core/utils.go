package core

import (
	"encoding/json"
	"fmt"
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
	vote           = 3
	permission     = 4

	cnsTxType = 0x11 // Used for sending transactions without address
	fwTxType  = 0x12 // Used fot sending transactions  about firewall

	DefaultConfigFilePath = "/config.json"
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
	From     string `json:"from"`
	To       string `json:"to"`
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
	} `json:"result"`
}

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
	params := []string{}
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
