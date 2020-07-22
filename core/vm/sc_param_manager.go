package vm

import (
	"fmt"
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

var (
	gasContractNameKey              = generateStateKey("GasContractName")
	isProduceEmptyBlockKey          = generateStateKey("IsProduceEmptyBlock")
	txGasLimitKey                   = generateStateKey("TxGasLimit")
	blockGasLimitKey                = generateStateKey("BlockGasLimit")
	isCheckContractDeployPermission = generateStateKey("isCheckContractDeployPermission")
	isApproveDeployedContractKey    = generateStateKey("IsApproveDeployedContract")
	isTxUseGasKey                   = generateStateKey("IsTxUseGas")
)

const (
	paramTrue  uint32 = 1
	paramFalse uint32 = 0
)

const (
	isCheckContractDeployPermissionDefault = paramFalse
	isTxUseGasDefault                      = paramFalse
	isApproveDeployedContractDefault       = paramFalse
	isProduceEmptyBlockDefault             = paramFalse
	gasContractNameDefault                 = ""
)

const (
	TxGasLimitMinValue        uint64 = 12771596 * 100 // 12771596 大致相当于 0.012772s
	TxGasLimitMaxValue        uint64 = 2e9            // 相当于 2s
	txGasLimitDefaultValue    uint64 = 1.5e9          // 相当于 1.5s
	BlockGasLimitMinValue     uint64 = 12771596 * 100 // 12771596 大致相当于 0.012772s
	BlockGasLimitMaxValue     uint64 = 2e10           // 相当于 20s
	blockGasLimitDefaultValue uint64 = 1e10           // 相当于 10s
	failFlag                         = -1
	sucFlag                          = 0
)
const (
	callerHasNoPermission CodeType = 0
	encodeFailure         CodeType = 1
	doParamSetSuccess     CodeType = 2
	paramInvalid          CodeType = 3
)

type ParamManager struct {
	stateDB      StateDB
	contractAddr *common.Address
	caller       common.Address
	blockNumber  *big.Int
}

func encode(i interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(i)
}

func (u *ParamManager) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.ParamManagerGas
}

func (u *ParamManager) Run(input []byte) ([]byte, error) {
	return execSC(input, u.AllExportFns())
}

func (u *ParamManager) setState(key, value []byte) {
	u.stateDB.SetState(*u.contractAddr, key, value)
}

func (u *ParamManager) getState(key []byte) []byte {
	return u.stateDB.GetState(*u.contractAddr, key)
}

func (u *ParamManager) setGasContractName(contractName string) (int32, error) {
	if len(contractName) == 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))
		return failFlag, errParamInvalid
	}
	ret, err := u.doParamSet(gasContractNameKey, contractName)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

//pass
func (u *ParamManager) getGasContractName() (string, error) {
	contractName := gasContractNameDefault
	err := u.getParam(gasContractNameKey, &contractName)
	if err != nil && err != errEmptyValue {
		return "", err
	}
	return contractName, nil
}

//pass
func (u *ParamManager) setIsProduceEmptyBlock(isProduceEmptyBlock uint32) (int32, error) {
	if isProduceEmptyBlock/2 != 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))
		return failFlag, errParamInvalid
	}
	ret, err := u.doParamSet(isProduceEmptyBlockKey, isProduceEmptyBlock)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

func (u *ParamManager) getIsProduceEmptyBlock() (uint32, error) {
	isProduceEmptyBlock := isProduceEmptyBlockDefault
	err := u.getParam(isProduceEmptyBlockKey, &isProduceEmptyBlock)
	if err != nil && err != errEmptyValue {
		return isProduceEmptyBlockDefault, err
	}
	return isProduceEmptyBlock, nil
}

func (u *ParamManager) setTxGasLimit(txGasLimit uint64) (int32, error) {
	if txGasLimit < TxGasLimitMinValue || txGasLimit > TxGasLimitMaxValue {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))

		return failFlag, errParamInvalid
	}
	// 获取区块 gas limit，其值应大于或等于每笔交易 gas limit 参数的值
	blockGasLimit := blockGasLimitDefaultValue
	err := u.getParam(blockGasLimitKey, &blockGasLimit)
	if err != nil && err != errEmptyValue {
		return failFlag, err
	}
	if txGasLimit > blockGasLimit {
		return failFlag, errParamInvalid
	}

	ret, err := u.doParamSet(txGasLimitKey, txGasLimit)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

func (u *ParamManager) getTxGasLimit() (uint64, error) {
	txGasLimit := txGasLimitDefaultValue
	err := u.getParam(txGasLimitKey, &txGasLimit)
	if err != nil && err != errEmptyValue {
		return txGasLimitDefaultValue, err
	}
	return txGasLimit, nil
}

func (u *ParamManager) setBlockGasLimit(blockGasLimit uint64) (int32, error) {
	if blockGasLimit < BlockGasLimitMinValue || blockGasLimit > BlockGasLimitMaxValue {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))
		return failFlag, errParamInvalid
	}
	key := txGasLimitKey
	txGasLimit := txGasLimitDefaultValue
	err := u.getParam(key, &txGasLimit)
	if err != nil && err != errEmptyValue {
		return failFlag, err
	}
	if txGasLimit > blockGasLimit {
		return failFlag, nil
	}

	ret, err := u.doParamSet(blockGasLimitKey, blockGasLimit)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

// 获取区块 gaslimit
func (u *ParamManager) getBlockGasLimit() (uint64, error) {
	var b = blockGasLimitDefaultValue
	if err := u.getParam(blockGasLimitKey, &b); err != nil && err != errEmptyValue {
		return blockGasLimitDefaultValue, err
	}
	return b, nil
}

// 设置是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *ParamManager) setCheckContractDeployPermission(permission uint32) (int32, error) {
	if permission/2 != 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))

		return failFlag, errParamInvalid
	}
	ret, err := u.doParamSet(isCheckContractDeployPermission, permission)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

// 获取是否是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *ParamManager) getCheckContractDeployPermission() (uint32, error) {
	var b = isCheckContractDeployPermissionDefault
	if err := u.getParam(isCheckContractDeployPermission, &b); err != nil && err != errEmptyValue {
		return isCheckContractDeployPermissionDefault, err
	}
	return b, nil
}

// 设置是否审核已部署的合约
// @isApproveDeployedContract:
// 1: 审核已部署的合约  0: 不审核已部署的合约
func (u *ParamManager) setIsApproveDeployedContract(isApproveDeployedContract uint32) (int32, error) {
	if isApproveDeployedContract/2 != 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))

		return failFlag, errParamInvalid
	}
	ret, err := u.doParamSet(isApproveDeployedContractKey, isApproveDeployedContract)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

// 获取是否审核已部署的合约的标志
func (u *ParamManager) getIsApproveDeployedContract() (uint32, error) {
	var b = isApproveDeployedContractDefault
	if err := u.getParam(isApproveDeployedContractKey, &b); err != nil && err != errEmptyValue {
		return isApproveDeployedContractDefault, err
	}
	return b, nil
}

// 本参数根据最新的讨论（2019.03.06之前）不再需要，即交易需要消耗gas。但是计费相关如消耗特定合约代币的参数由 setGasContractName 进行设置
// 设置交易是否消耗 gas
// @isTxUseGas:
// 1:  交易消耗 gas  0: 交易不消耗 gas
func (u *ParamManager) setIsTxUseGas(isTxUseGas uint64) (int32, error) {
	if isTxUseGas/2 != 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))
		return failFlag, errParamInvalid
	}
	ret, err := u.doParamSet(isTxUseGasKey, isTxUseGas)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

// 获取交易是否消耗 gas
func (u *ParamManager) getIsTxUseGas() (uint32, error) {
	var isUseGas = isTxUseGasDefault
	if err := u.getParam(isTxUseGasKey, &isUseGas); err != nil && err != errEmptyValue {
		return isTxUseGasDefault, err
	}
	return isUseGas, nil
}
func (u *ParamManager) doParamSet(key []byte, value interface{}) (int32, error) {
	if !hasParamOpPermission(u.stateDB, u.caller) {
		u.emitNotifyEventInParam(callerHasNoPermission, fmt.Sprintf("%s has no permission to adjust param.", u.caller.String()))
		return failFlag, errNoPermission
	}
	if err := u.setParam(key, value); err != nil {
		u.emitNotifyEventInParam(encodeFailure, fmt.Sprintf("%v failed to encode.", value))
		return failFlag, err
	}
	u.emitNotifyEventInParam(doParamSetSuccess, fmt.Sprintf("param set successful."))
	return sucFlag, nil
}

func (u *ParamManager) setParam(key []byte, val interface{}) error {
	value, err := rlp.EncodeToBytes(val)
	if err != nil {
		return err
	}
	u.setState(key, value)
	return nil
}

func (u *ParamManager) getParam(key []byte, val interface{}) error {
	value := u.getState(key)
	if len(value) == 0 {
		return errEmptyValue
	}
	if err := rlp.DecodeBytes(value, val); err != nil {
		return err
	}
	return nil
}

func (u *ParamManager) emitNotifyEventInParam(code CodeType, msg string) {
	topic := "Notify"
	emitEvent(*u.contractAddr, u.stateDB, u.blockNumber.Uint64(), topic, code, msg)
}

//for access control
func (u *ParamManager) AllExportFns() SCExportFns {
	return SCExportFns{
		"setGasContractName":               u.setGasContractName,
		"getGasContractName":               u.getGasContractName,
		"setIsProduceEmptyBlock":           u.setIsProduceEmptyBlock,
		"getIsProduceEmptyBlock":           u.getIsProduceEmptyBlock,
		"setTxGasLimit":                    u.setTxGasLimit,
		"getTxGasLimit":                    u.getTxGasLimit,
		"setBlockGasLimit":                 u.setBlockGasLimit,
		"getBlockGasLimit":                 u.getBlockGasLimit,
		"setCheckContractDeployPermission": u.setCheckContractDeployPermission,
		"getCheckContractDeployPermission": u.getCheckContractDeployPermission,
		"setIsApproveDeployedContract":     u.setIsApproveDeployedContract,
		"getIsApproveDeployedContract":     u.getIsApproveDeployedContract,
		"setIsTxUseGas":                    u.setIsTxUseGas,
		"getIsTxUseGas":                    u.getIsTxUseGas,
	}
}
