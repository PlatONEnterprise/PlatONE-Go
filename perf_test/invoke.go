package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/core"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

var (
	txHashList []string
	lastTxHash string
)

func invoke(contractAddress string, abiPath string, funcParams string, txType int) (error, interface{}) {

	if contractAddress == "" {
		fmt.Printf("contract address can't be empty!")
		return errors.New("contract address can't be empty!"), nil
	}

	if abiPath == "" {
		fmt.Printf("abi can't be empty!")
		return errors.New("abi can't be empty!"), nil
	}

	if funcParams == "" {
		fmt.Printf("func can't be empty!")
		return errors.New("func can't be empty!"), nil
	}

	err, ret := InvokeContract(contractAddress, abiPath, funcParams, txType)
	if err != nil {
		panic(fmt.Errorf("invokeContract contract error,%s", err.Error()))
	}
	return nil, ret

}

func InvokeContract(contractAddr string, abiPath string, funcParams string, txType int) (error, interface{}) {

	funcName, inputParams := core.GetFuncNameAndParams(funcParams)

	//Judging whether this method exists or not
	abiFunc, err := parseFuncFromAbi(abiPath, funcName)
	if err != nil {
		return err, nil
	}

	if len(abiFunc.Inputs) != len(inputParams) {
		return fmt.Errorf("incorrect number of parameters ,request=%d,get=%d\n",
			len(abiFunc.Inputs), len(funcParams)), nil
	}

	if txType == 0 {
		txType = invokeContract
	}

	paramArr := [][]byte{
		core.Int64ToBytes(int64(txType)),
		[]byte(funcName),
	}

	for i, v := range inputParams {
		input := abiFunc.Inputs[i]
		p, e := core.StringConverter(v, input.Type)
		if e != nil {
			return fmt.Errorf("incorrect param type: %s,index:%d", v, i), nil
		}
		paramArr = append(paramArr, p)
	}

	paramBytes, e := rlp.EncodeToBytes(paramArr)
	if e != nil {
		return fmt.Errorf("rpl encode error,%s", e.Error()), nil
	}

	txParams := core.TxParams{
		From:     config.From,
		To:       contractAddr,
		GasPrice: config.GasPrice,
		Gas:      config.Gas,
		Data:     hexutil.Encode(paramBytes),
		TxType:   txType,
	}

	var r string
	if abiFunc.Constant == "true" {
		params := make([]interface{}, 2)
		params[0] = txParams
		params[1] = "latest"

		//paramJson, _ := json.Marshal(params)
		//fmt.Printf("\n request json data：%s \n", string(paramJson))
		r, err = Send(params, "eth_call")
	} else {
		params := make([]interface{}, 1)
		params[0] = txParams

		//paramJson, _ := json.Marshal(params)
		//fmt.Printf("\n request json data：%s \n", string(paramJson))
		r, err = Send(params, "eth_sendTransaction")
	}

	//fmt.Printf("\n response json：%s \n", r)

	if err != nil {
		return fmt.Errorf("send http post to invokeContract contract error,%s", e.Error()), nil
	}
	resp := parseResponse(r)
	if resp == nil {
		return nil, nil
	}

	//parse the return type through adi
	if abiFunc.Constant == "true" {
		if len(abiFunc.Outputs) != 0 && abiFunc.Outputs[0].Type != "void" {
			bytes, _ := hexutil.Decode(resp.Result)
			result := core.BytesConverter(bytes, abiFunc.Outputs[0].Type)
			//fmt.Printf("\nresult: %v\n", result)
			return nil, result
		}
	} else {
		//fmt.Printf("\n trasaction hash: %s\n", resp.Result)
		pos := strings.Index(resp.Result, "0x")
		lastTxHash = resp.Result[pos : pos+common.HashLength*2+2]
		txHashList = append(txHashList, lastTxHash)
	}
	return nil, nil
}

func assembleForWs(contractAddress string, abiPath string, funcParams string, txType int) []interface{} {

	if contractAddress == "" {
		fmt.Printf("contract address can't be empty!")
		return nil
	}

	if abiPath == "" {
		fmt.Printf("abi can't be empty!")
		return nil
	}

	if funcParams == "" {
		fmt.Printf("func can't be empty!")
		return nil
	}

	return AssembleParamsByWs(contractAddress, abiPath, funcParams, txType)

}

func AssembleParamsByWs(contractAddr string, abiPath string, funcParams string, txType int) []interface{} {

	funcName, inputParams := core.GetFuncNameAndParams(funcParams)

	//Judging whether this method exists or not
	abiFunc, err := parseFuncFromAbi(abiPath, funcName)
	if err != nil {
		return nil
	}

	if len(abiFunc.Inputs) != len(inputParams) {
		return nil
	}

	if txType == 0 {
		txType = invokeContract
	}

	paramArr := [][]byte{
		core.Int64ToBytes(int64(txType)),
		[]byte(funcName),
	}

	for i, v := range inputParams {
		input := abiFunc.Inputs[i]
		p, e := core.StringConverter(v, input.Type)
		if e != nil {
			return nil
		}
		paramArr = append(paramArr, p)
	}

	paramBytes, e := rlp.EncodeToBytes(paramArr)
	if e != nil {
		return nil
	}

	txParams := core.TxParams{
		From:     config.From,
		To:       contractAddr,
		GasPrice: config.GasPrice,
		Gas:      config.Gas,
		Data:     hexutil.Encode(paramBytes),
		TxType:   txType,
	}

	if abiFunc.Constant == "true" {
		params := make([]interface{}, 2)
		params[0] = txParams
		params[1] = "latest"
		return params
	} else {
		params := make([]interface{}, 1)
		params[0] = txParams
		return params
	}

}
