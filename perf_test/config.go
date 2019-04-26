package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/BCOSnetwork/BCOS-Go/cmd/ctool/core"
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
	config = core.Config{}
)

func parseAbiFromJson(fileName string) ([]core.FuncDesc, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("parse abi file error: %s", err.Error())
	}
	var a []core.FuncDesc
	if err := json.Unmarshal(bytes, &a); err != nil {
		return nil, fmt.Errorf("parse abi to json error: %s", err.Error())
	}
	return a, nil
}

func parseFuncFromAbi(fileName string, funcName string) (*core.FuncDesc, error) {
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

func UpdateConfigUrl(url string) {
	config.Url = url
}
