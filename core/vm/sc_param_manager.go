package vm

import (
	"fmt"
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

var (
	gasContractNameKey                 = "GasContractName"
	isProduceEmptyBlockKey             = "IsProduceEmptyBlock"
	txGasLimitKey                      = "TxGasLimit"
	blockGasLimitKey                   = "BlockGasLimit"
	isAllowAnyAccountDeployContractKey = "IsAllowAnyAccountDeployContract"
	isCheckContractDeployPermission    = "isCheckContractDeployPermission"
	isApproveDeployedContractKey       = "IsApproveDeployedContract"
	isTxUseGasKey                      = "IsTxUseGas"
)

//var (
//	ErrHasNoPermission = errors.New("has no permission")
//	ErrParamInvalid         = errors.New("something invalid")
//)
//
const (
	paramTrue  uint32 = 1
	paramFalse uint32 = 0
)

const (
	isCheckContractDeployPermissionDefault = paramFalse
	isTxUseGasDefault                      = paramFalse
	isApproveDeployedContractDefault       = paramFalse
	isProduceEmptyBlockDefault             = paramFalse
	isAllowAnyAccountDeployContractDefault = paramFalse
	gasContractNameDefault                 = ""
)

const (
	txGasLimitMinValue        uint64 = 12771596 * 100 // 12771596 大致相当于 0.012772s
	txGasLimitMaxValue        uint64 = 2e9            // 相当于 2s
	txGasLimitDefaultValue    uint64 = 1.5e9          // 相当于 1.5s
	blockGasLimitMinValue     uint64 = 12771596 * 100 // 12771596 大致相当于 0.012772s
	blockGasLimitMaxValue     uint64 = 2e10           // 相当于 20s
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
	//err := u.setDefaultValue()
	//if nil != err {
	//	return nil, err
	//}
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
		return failFlag, ErrParamInvalid
	}
	ret, err := u.doParamSet(gasContractNameKey, contractName)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

//pass
func (u *ParamManager) getGasContractName() (string, error) {
	key, err := encode(gasContractNameKey)
	if err != nil {
		return gasContractNameDefault, err
	}
	contractName := u.getState(key)
	if len(contractName) == 0 {
		return gasContractNameDefault, nil
	}
	var ret string
	if err := rlp.DecodeBytes(contractName, &ret); nil != err {
		return gasContractNameDefault, err
	}
	return ret, nil
	//return encode(contractName)
}

//pass
func (u *ParamManager) setIsProduceEmptyBlock(isProduceEmptyBlock uint32) (int32, error) {
	if isProduceEmptyBlock/2 != 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))
		return failFlag, ErrParamInvalid
	}
	ret, err := u.doParamSet(isProduceEmptyBlockKey, isProduceEmptyBlock)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

func (u *ParamManager) getIsProduceEmptyBlock() (uint32, error) {
	key, err := encode(isProduceEmptyBlockKey)
	if err != nil {
		return isProduceEmptyBlockDefault, err
	}
	isProduceEmptyBlock := u.getState(key)
	if len(isProduceEmptyBlock) == 0 {
		return isProduceEmptyBlockDefault, nil
	}
	var ret uint32
	if err := rlp.DecodeBytes(isProduceEmptyBlock, &ret); nil != err {
		return isProduceEmptyBlockDefault, err
	}
	return ret, nil
}

func (u *ParamManager) setTxGasLimit(txGasLimit uint64) (int32, error) {
	if txGasLimit < txGasLimitMinValue || txGasLimit > txGasLimitMaxValue {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))

		return failFlag, ErrParamInvalid
	}
	// 获取区块 gas limit，其值应大于或等于每笔交易 gas limit 参数的值

	key, err := encode(blockGasLimitKey)
	if err != nil {
		return failFlag, err
	}
	currentBlockGasLimit := u.getState(key)
	blockGasLimit := blockGasLimitDefaultValue
	if currentBlockGasLimit != nil {
		if err := rlp.DecodeBytes(currentBlockGasLimit, &blockGasLimit); nil != err {
			return failFlag, err
		}
		if txGasLimit > blockGasLimit {
			return failFlag, ErrParamInvalid
		}
	}
	ret, err := u.doParamSet(txGasLimitKey, txGasLimit)
	if err != nil {
		return failFlag, err
	}
	return ret, err

}

func (u *ParamManager) getTxGasLimit() (uint64, error) {
	key, err := encode(txGasLimitKey)
	if err != nil {
		return txGasLimitDefaultValue, err
	}
	txGasLimit := u.getState(key)
	if len(txGasLimit) == 0 {
		return txGasLimitDefaultValue, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(txGasLimit, &ret); nil != err {
		return txGasLimitDefaultValue, err
	}
	return ret, nil
}

func (u *ParamManager) setBlockGasLimit(blockGasLimit uint64) (int32, error) {
	if blockGasLimit < blockGasLimitMinValue || blockGasLimit > blockGasLimitMaxValue {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))

		return failFlag, ErrParamInvalid
	}
	key, err := encode(txGasLimitKey)
	if err != nil {
		return failFlag, err
	}
	currentTxGasLimit := u.getState(key)
	txGasLimit := txGasLimitDefaultValue
	if currentTxGasLimit != nil {
		if err := rlp.DecodeBytes(currentTxGasLimit, &txGasLimit); nil != err {
			return failFlag, err
		}
		if txGasLimit > blockGasLimit {
			return failFlag, err
		}
	}
	ret, err := u.doParamSet(blockGasLimitKey, blockGasLimit)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

// 获取区块 gaslimit
func (u *ParamManager) getBlockGasLimit() (uint64, error) {
	key, err := encode(blockGasLimitKey)
	if err != nil {
		return blockGasLimitDefaultValue, err
	}
	blockGasLimit := u.getState(key)
	if len(blockGasLimit) == 0 {
		return blockGasLimitDefaultValue, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(blockGasLimit, &ret); nil != err {
		return blockGasLimitDefaultValue, err
	}
	return ret, nil
}


// 设置是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *ParamManager) setCheckContractDeployPermission(permission uint32) (int32, error) {
	if permission/2 != 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))

		return failFlag, ErrParamInvalid
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
	key, err := encode(isCheckContractDeployPermission)
	if err != nil {
		return isCheckContractDeployPermissionDefault, err
	}
	checkPermission := u.getState(key)
	if len(checkPermission) == 0 {
		return isCheckContractDeployPermissionDefault, nil
	}
	var ret uint32
	if err := rlp.DecodeBytes(checkPermission, &ret); nil != err {
		return isCheckContractDeployPermissionDefault, err
	}
	return ret, nil
}

// 设置是否审核已部署的合约
// @isApproveDeployedContract:
// 1: 审核已部署的合约  0: 不审核已部署的合约
func (u *ParamManager) setIsApproveDeployedContract(isApproveDeployedContract uint32) (int32, error) {
	if isApproveDeployedContract/2 != 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))

		return failFlag, ErrParamInvalid
	}
	ret, err := u.doParamSet(isApproveDeployedContractKey, isApproveDeployedContract)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

// 获取是否审核已部署的合约的标志
func (u *ParamManager) getIsApproveDeployedContract() (uint32, error) {
	key, err := encode(isApproveDeployedContractKey)
	if err != nil {
		return isApproveDeployedContractDefault, err
	}
	isApproveDeployedContract := u.getState(key)
	if len(isApproveDeployedContract) == 0 {
		return isApproveDeployedContractDefault, nil
	}
	var ret uint32
	if err := rlp.DecodeBytes(isApproveDeployedContract, &ret); nil != err {
		return isApproveDeployedContractDefault, err
	}
	return ret, nil
}

// 本参数根据最新的讨论（2019.03.06之前）不再需要，即交易需要消耗gas。但是计费相关如消耗特定合约代币的参数由 setGasContractName 进行设置
// 设置交易是否消耗 gas
// @isTxUseGas:
// 1:  交易消耗 gas  0: 交易不消耗 gas
func (u *ParamManager) setIsTxUseGas(isTxUseGas uint64) (int32, error) {
	if isTxUseGas/2 != 0 {
		u.emitNotifyEventInParam(paramInvalid, fmt.Sprintf("param is invalid."))

		return failFlag, ErrParamInvalid
	}
	ret, err := u.doParamSet(isTxUseGasKey, isTxUseGas)
	if err != nil {
		return failFlag, err
	}
	return ret, err
}

// 获取交易是否消耗 gas
func (u *ParamManager) getIsTxUseGas() (uint32, error) {
	key, err := encode(isTxUseGasKey)
	if err != nil {
		return isTxUseGasDefault, err
	}
	isTxUseGas := u.getState(key)
	if len(isTxUseGas) == 0 {
		return isTxUseGasDefault, nil
	}
	var ret uint32
	if err := rlp.DecodeBytes(isTxUseGas, &ret); nil != err {
		return isTxUseGasDefault, err
	}
	return ret, nil
}
func (u *ParamManager) doParamSet(inputKey, inputValue interface{}) (int32, error) {
	if !hasParamOpPermission(u.stateDB, u.caller) {
		u.emitNotifyEventInParam(callerHasNoPermission, fmt.Sprintf("%s has no permission to adjust param.", u.caller.String()))
		return failFlag, ErrNoPermission
	}
	key, err := encode(inputKey)
	if err != nil {
		u.emitNotifyEventInParam(encodeFailure, fmt.Sprintf("%s failed to encode.", string(key)))
		return failFlag, err
	}
	value, err := encode(inputValue)
	if err != nil {
		u.emitNotifyEventInParam(encodeFailure, fmt.Sprintf("%s failed to encode.", string(value)))
		return failFlag, err
	}
	u.setState(key, value)
	u.emitNotifyEventInParam(doParamSetSuccess, fmt.Sprintf("param set successful."))
	return sucFlag, nil

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
