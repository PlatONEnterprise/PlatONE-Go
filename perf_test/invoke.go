package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PlatONnetwork/PlatON-Go/cmd/ctool/core"
	"github.com/PlatONnetwork/PlatON-Go/common/hexutil"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
)

func invoke(contractAddress string, abiPath string, funcParams string, txType int) error {

	if contractAddress == "" {
		fmt.Printf("contract address can't be empty!")
		return errors.New("contract address can't be empty!")
	}

	if abiPath == "" {
		fmt.Printf("abi can't be empty!")
		return errors.New("abi can't be empty!")
	}

	if funcParams == "" {
		fmt.Printf("func can't be empty!")
		return errors.New("func can't be empty!")
	}

	parseConfigJson(*configPath)

	//Judging whether this contract exists or not
	// if !getContractByAddress(contractAddress) {
	// 	panic("the contract address is not exist ...")
	// }

	err := InvokeContract(contractAddress, abiPath, funcParams, txType)
	if err != nil {
		panic(fmt.Errorf("invokeContract contract error,%s", err.Error()))
	}
	return nil

}

func InvokeContract(contractAddr string, abiPath string, funcParams string, txType int) error {

	funcName, inputParams := core.GetFuncNameAndParams(funcParams)

	//Judging whether this method exists or not
	abiFunc, err := parseFuncFromAbi(abiPath, funcName)
	if err != nil {
		return err
	}

	if len(abiFunc.Inputs) != len(inputParams) {
		return fmt.Errorf("incorrect number of parameters ,request=%d,get=%d\n",
			len(abiFunc.Inputs), len(funcParams))
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
			return fmt.Errorf("incorrect param type: %s,index:%d", v, i)
		}
		paramArr = append(paramArr, p)
	}

	paramBytes, e := rlp.EncodeToBytes(paramArr)
	if e != nil {
		return fmt.Errorf("rpl encode error,%s", e.Error())
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

		paramJson, _ := json.Marshal(params)
		fmt.Printf("\n request json data：%s \n", string(paramJson))
		r, err = Send(params, "eth_call")
	} else {
		params := make([]interface{}, 1)
		params[0] = txParams

		paramJson, _ := json.Marshal(params)
		fmt.Printf("\n request json data：%s \n", string(paramJson))
		r, err = Send(params, "eth_sendTransaction")
	}

	fmt.Printf("\n response json：%s \n", r)

	if err != nil {
		return fmt.Errorf("send http post to invokeContract contract error,%s", e.Error())
	}
	resp := parseResponse(r)

	//parse the return type through adi
	if abiFunc.Constant == "true" {
		if len(abiFunc.Outputs) != 0 && abiFunc.Outputs[0].Type != "void" {
			bytes, _ := hexutil.Decode(resp.Result)
			result := core.BytesConverter(bytes, abiFunc.Outputs[0].Type)
			fmt.Printf("\nresult: %v\n", result)
			return nil
		}
		fmt.Printf("\n result: []\n")
	} else {
		fmt.Printf("\n trasaction hash: %s\n", resp.Result)
	}
	return nil
}
