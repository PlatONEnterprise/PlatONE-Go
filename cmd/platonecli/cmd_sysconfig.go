package main

import (
	"fmt"
	"strconv"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"gopkg.in/urfave/cli.v1"
)

var (
	SysConfigCmd = cli.Command{
		Name:  "sysconfig",
		Usage: "Manage the system configurations",
		Subcommands: []cli.Command{
			setCfg,
			getCfg,
		},
	}

	setCfg = cli.Command{
		Name:   "set",
		Usage:  "set the system configurations",
		Action: setSysConfig,
		Flags:  sysConfigCmdFlags,
	}

	getCfg = cli.Command{
		Name:   "get",
		Usage:  "get the system configurations",
		Action: getSysConfig,
		Flags:  getSysConfigCmdFlags,
	}
)

func checkConfigParam(param string, key string) bool {

	switch key {
	case "TxGasLimit":
		// number check
		num, err := strconv.ParseUint(param, 10, 0)
		if err != nil {

		}

		// param check
		isInRange := vm.TxGasLimitMinValue < num && vm.TxGasLimitMaxValue > num
		if !isInRange {
			fmt.Printf("the transaction gas limit should be within (%d, %d)\n", vm.TxGasLimitMinValue, vm.TxGasLimitMaxValue)
			return false
		}
	case "BlockGasLimit":
		num, err := strconv.ParseUint(param, 10, 0)
		if err != nil {

		}

		isInRange := vm.BlockGasLimitMinValue < num && vm.BlockGasLimitMaxValue > num
		if !isInRange {
			fmt.Printf("the block gas limit should be within (%d, %d)\n", vm.BlockGasLimitMinValue, vm.BlockGasLimitMaxValue)
			return false
		}
	}

	return true
}

func setSysConfig(c *cli.Context) {

	txGasLimit := c.String(TxGasLimitFlags.Name)
	blockGasLimit := c.String(BlockGasLimitFlags.Name)
	isTxUseGas := c.String(IsTxUseGasFlags.Name)
	isApproveDeployedContract := c.String(IsApproveDeployedContractFlags.Name)
	isCheckContractDeployPermission := c.String(IsCheckContractDeployPermissionFlags.Name)
	isProduceEmptyBlock := c.String(IsProduceEmptyBlockFlags.Name)
	gasContractName := c.String(GasContractNameFlags.Name)

	/*
		if len(c.Args()) > 1 {
			utils.Fatalf("please set one system configuration at a time")
		}*/

	setConfig(c, txGasLimit, "TxGasLimit")
	setConfig(c, blockGasLimit, "BlockGasLimit")
	setConfig(c, isTxUseGas, "IsTxUseGas")
	setConfig(c, isApproveDeployedContract, "IsApproveDeployedContract")
	setConfig(c, isCheckContractDeployPermission, "CheckContractDeployPermission")
	setConfig(c, isProduceEmptyBlock, "IsProduceEmptyBlock")
	setConfig(c, gasContractName, "GasContractName")

}

func setConfig(c *cli.Context, param string, name string) {

	if !checkConfigParam(param, name) {
		return
	}

	funcName := "set" + name
	funcParams := CombineFuncParams(param)

	result := contractCall(c, funcParams, funcName, parameterManagementAddress)
	fmt.Printf("result: %s\n", result)
}

func getSysConfig(c *cli.Context) {

	txGasLimit := c.String(TxGasLimitFlags.Name)
	blockGasLimit := c.String(BlockGasLimitFlags.Name)
	isTxUseGas := c.String(IsTxUseGasFlags.Name)
	isApproveDeployedContract := c.String(IsApproveDeployedContractFlags.Name)
	isCheckContractDeployPermission := c.String(IsCheckContractDeployPermissionFlags.Name)
	isProduceEmptyBlock := c.String(IsProduceEmptyBlockFlags.Name)
	gasContractName := c.String(GasContractNameFlags.Name)

	if len(c.Args()) > 1 {
		utils.Fatalf("please set one system configuration at a time")
	}

	getConfig(c, txGasLimit != "", "getTxGasLimit")
	getConfig(c, blockGasLimit != "", "getBlockGasLimit")
	getConfig(c, isTxUseGas != "", "getIsTxUseGas")
	getConfig(c, isApproveDeployedContract != "", "getIsApproveDeployedContract")
	getConfig(c, isCheckContractDeployPermission != "", "getIsCheckContractDeployPermission")
	getConfig(c, isProduceEmptyBlock != "", "getIsProduceEmptyBlock")
	getConfig(c, gasContractName != "", "getGasContractName")

}

func getConfig(c *cli.Context, isGet bool, funcName string) {

	if isGet {
		result := contractCall(c, nil, funcName, parameterManagementAddress)
		fmt.Printf("%s result: %v\n", funcName, result)
	}
}
