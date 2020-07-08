package vm

import (
	"errors"
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
var (
	ErrHasNoPermission = errors.New("has no permission")
	ErrInvalid         = errors.New("something invalid")
)

const (
	txGasLimitMinValue     uint64 = 12771596 * 100 // 12771596 大致相当于 0.012772s
	txGasLimitMaxValue     uint64 = 2e9            // 相当于 2s
	txGasLimitDefaultValue uint64      = 1.5e9          // 相当于 1.5s
	blockGasLimitMinValue  uint64 = 12771596 * 100 // 12771596 大致相当于 0.012772s
	blockGasLimitMaxValue  uint64 = 2e10           // 相当于 20s
	blockGasLimitDefaultValue uint64 = 1e10        // 相当于 10s
	//produceDurationMaxValue = 60
	//produceDurationDefaultValue = 10
	//blockIntervalMinValue = 1
	//blockIntervalDefaultValue = 1
	errFlag  = -1
	failFlag = 0
	sucFlag  = 1

)

type ParamManager struct {
	state      StateDB
	codeAddr   *common.Address
	callerAddr common.Address
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
	err :=u.setDefaultValue()
	if nil != err{
		return nil, err
	}
	return execSC(input, u.AllExportFns())
}

func (u *ParamManager) setState(key, value []byte) {
	u.state.SetState(*u.codeAddr, key, value)
}

func (u *ParamManager) getState(key []byte) []byte {
	return u.state.GetState(*u.codeAddr, key)
}

func (u *ParamManager) setGasContractName(contractName string) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return nil, ErrHasNoPermission
	}
	if len(contractName) == 0 {
		//event here
		return nil, ErrInvalid
	}
	key, err := encode(gasContractNameKey)
	if err != nil {
		return nil, err
	}
	value, err := encode(contractName)
	if err != nil {
		return nil, err
	}
	u.setState(key, value)
	return nil, nil
}

//pass
func (u *ParamManager) getGasContractName() (string, error) {
	key, err := encode(gasContractNameKey)
	if err != nil {
		return "", err
	}
	contractName := u.getState(key)
	if len(contractName) == 0{
		return "", nil
	}
	var ret string
	if err := rlp.DecodeBytes(contractName, &ret); nil != err {
		return "", err
	}
	return ret, nil
	//return encode(contractName)
}

//pass
func (u *ParamManager) setIsProduceEmptyBlock(isProduceEmptyBlock uint32) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return nil, ErrHasNoPermission
	}
	if 0 != isProduceEmptyBlock && 1 != isProduceEmptyBlock {
		//event
		return nil, ErrInvalid
	}
	key, err := encode(isProduceEmptyBlockKey)
	if err != nil {
		return nil, err
	}
	value, err := encode(isProduceEmptyBlock)
	if err != nil {
		return nil, err
	}
	u.setState(key, value)
	//event here
	return nil, nil
}

func (u *ParamManager) getIsProduceEmptyBlock() (int32, error) {
	key, err := encode(isProduceEmptyBlockKey)
	if err != nil {
		return failFlag, err
	}
	isProduceEmptyBlock := u.getState(key)
	if len(isProduceEmptyBlock) == 0{
		return errFlag, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(isProduceEmptyBlock, &ret); nil != err {
		return 0, err
	}
	if ret == 1 {
		return sucFlag, nil
	}else {
		return failFlag, nil
	}
}

func (u *ParamManager) setTxGasLimit(txGasLimit uint64) ([]byte, error) {
	var err error
	if !u.hasPermission() {
		//event here
		//return encode(FAL)
		return nil, err
	}
	if txGasLimit < txGasLimitMinValue || txGasLimit > txGasLimitMaxValue {
		//event
		return nil, err
	}
	// 获取区块 gas limit，其值应大于或等于每笔交易 gas limit 参数的值
	key, err := encode(blockGasLimitKey)
	if err != nil {
		return nil, err
	}
	blockGasLimit := u.getState(key)
	if blockGasLimit != nil {
		var ci uint64
		if err := rlp.DecodeBytes(blockGasLimit, &ci); nil != err {
			return nil, err
		}
		if txGasLimit > ci {
			//event
			return nil, ErrInvalid
		}
	}

	key, err = encode(txGasLimitKey)
	if err != nil {
		return nil, err
	}
	value, err := encode(txGasLimit)
	if err != nil {
		return nil, err
	}
	u.setState(key, value)
	//event
	return nil, nil
	//return  encode(SUC)
}

func (u *ParamManager) getTxGasLimit() (uint64, error) {
	key, err := encode(txGasLimitKey)
	if err != nil {
		return failFlag, err
	}
	txGasLimit := u.getState(key)
	if len(txGasLimit) == 0{
		return failFlag, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(txGasLimit, &ret); nil != err {
		return 0, err
	}
	return ret, nil
}

func (u *ParamManager) setBlockGasLimit(blockGasLimit uint64) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return nil, ErrHasNoPermission
	}
	if blockGasLimit < blockGasLimitMinValue || blockGasLimit > blockGasLimitMaxValue {
		//event
		return nil, ErrInvalid
	}
	key, err := encode(txGasLimitKey)
	if err != nil {
		return nil, err
	}
	txGasLimit1 := u.getState(key)
	if txGasLimit1 != nil {
		var ci uint64
		if err := rlp.DecodeBytes(txGasLimit1, &ci); nil != err {
			return nil, err
		}
		if ci > blockGasLimit {
			//event
			return nil, err
		}
	}
	key, err = encode(blockGasLimitKey)
	if err != nil {
		return nil, err
	}
	value, err := encode(blockGasLimit)
	if err != nil {
		return nil, err
	}
	u.setState(key, value)
	//event here
	return nil, nil
}

// 获取区块 gaslimit
func (u *ParamManager) getBlockGasLimit() (uint64, error) {
	key, err := encode(blockGasLimitKey)
	if err != nil {
		return failFlag, err
	}
	blockGasLimit := u.getState(key)
	if len(blockGasLimit) == 0{
		return failFlag, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(blockGasLimit, &ret); nil != err {
		return 0, err
	}
	return ret, nil
}

// 设置是否允许任意用户部署合约
// @isAllowAnyAccountDeployContract:
// 0: 允许任意用户部署合约  1: 用户具有相应权限才可以部署合约
// 默认为0，即允许任意用户部署合约
func (u *ParamManager) setAllowAnyAccountDeployContract(isAllowAnyAccountDeployContract uint32) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return nil, ErrHasNoPermission
	}
	if 0 != isAllowAnyAccountDeployContract && 1 != isAllowAnyAccountDeployContract {
		//event here
		return nil, ErrInvalid
	}
	key, err := encode(isAllowAnyAccountDeployContractKey)
	if err != nil {
		return nil, err
	}
	value, err := encode(isAllowAnyAccountDeployContract)
	if err != nil {
		return nil, err
	}
	u.setState(key, value)
	//event here
	return nil, nil
}

// 获取是否允许任意用户部署合约的标志
func (u *ParamManager) getAllowAnyAccountDeployContract() (int32, error) {
	key, err := encode(isAllowAnyAccountDeployContractKey)
	if err != nil {
		return failFlag, err
	}
	isAllowAnyAccountDeployContract := u.getState(key)
	if len(isAllowAnyAccountDeployContract) == 0{
		return failFlag, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(isAllowAnyAccountDeployContract, &ret); nil != err {
		return failFlag, err
	}
	if ret == 1 {
		return sucFlag, nil
	}else {
		return failFlag, nil
	}
}

// 设置是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *ParamManager) setCheckContractDeployPermission(checkPermission uint32) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return nil, ErrHasNoPermission
	}
	if 0 != checkPermission && 1 != checkPermission {
		//event here
		return nil, ErrInvalid
	}
	key, err := encode(isCheckContractDeployPermission)
	if err != nil {
		return nil, err
	}
	value, err := encode(checkPermission)
	if err != nil {
		return nil, err
	}
	u.setState(key, value)
	//event here
	return nil, nil
}

// 获取是否是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *ParamManager) getCheckContractDeployPermission() (int32, error) {
	key, err := encode(isCheckContractDeployPermission)
	if err != nil {
		return failFlag, err
	}
	checkPermission := u.getState(key)
	if len(checkPermission) == 0{
		return failFlag, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(checkPermission, &ret); nil != err {
		return failFlag, err
	}
	if ret == 1 {
		return sucFlag, nil
	}else {
		return failFlag, nil
	}

}

// 设置是否审核已部署的合约
// @isApproveDeployedContract:
// 1: 审核已部署的合约  0: 不审核已部署的合约
func (u *ParamManager) setIsApproveDeployedContract(isApproveDeployedContract uint32) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return nil, ErrHasNoPermission
	}
	if 0 != isApproveDeployedContract && 1 != isApproveDeployedContract {
		//event here
		return nil, nil
	}
	key, err := encode(isApproveDeployedContractKey)
	if err != nil {
		return nil, err
	}
	value, err := encode(isApproveDeployedContract)
	if err != nil {
		return nil, err
	}
	u.setState(key, value)
	//event here
	return nil, nil
}

// 获取是否审核已部署的合约的标志
func (u *ParamManager) getIsApproveDeployedContract() (int32, error) {
	key, err := encode(isApproveDeployedContractKey)
	if err != nil {
		return failFlag, err
	}
	isApproveDeployedContract := u.getState(key)
	if len(isApproveDeployedContract) == 0{
		return failFlag, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(isApproveDeployedContract, &ret); nil != err {
		return failFlag, err
	}
	if ret == 1 {
		return sucFlag, nil
	}else {
		return failFlag, nil
	}
}

// 本参数根据最新的讨论（2019.03.06之前）不再需要，即交易需要消耗gas。但是计费相关如消耗特定合约代币的参数由 setGasContractName 进行设置
// 设置交易是否消耗 gas
// @isTxUseGas:
// 1:  交易消耗 gas  0: 交易不消耗 gas
func (u *ParamManager) setIsTxUseGas(isTxUseGas uint64) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return nil, ErrHasNoPermission
	}
	if 0 != isTxUseGas && 1 != isTxUseGas {
		//event here
		return nil, ErrInvalid
	}
	key, err := encode(isTxUseGasKey)
	if err != nil {
		return nil, err
	}
	value, err := encode(isTxUseGas)
	if err != nil {
		return nil, err
	}
	u.setState(key, value)
	//event here
	return nil, nil
}

// 获取交易是否消耗 gas
func (u *ParamManager) getIsTxUseGas() (uint64, error) {
	key, err := encode(isTxUseGasKey)
	if err != nil {
		return failFlag, err
	}
	isTxUseGas := u.getState(key)
	if len(isTxUseGas) == 0 {
		return failFlag, nil
	}
	var ret uint64
	if err := rlp.DecodeBytes(isTxUseGas, &ret); nil != err {
		return failFlag, err
	}
	return ret, nil
}

func (u *ParamManager)setDefaultValue() error{
	gasLimit, err := rlp.EncodeToBytes(txGasLimitKey)
	if err != nil {
		return err
	}
	gasValue, err := rlp.EncodeToBytes(txGasLimitDefaultValue)
	if err != nil {
		return err
	}
	data := u.getState(gasLimit)
	if len(data) == 0 {
		u.setState(gasLimit, gasValue)
	}
	blockGasLimit, err := rlp.EncodeToBytes(blockGasLimitKey)
	if err != nil {
		return err
	}
	blockGasValue, err := rlp.EncodeToBytes(blockGasLimitDefaultValue)
	if err != nil {
		return err
	}
	data = u.getState(blockGasLimit)
	if len(data) == 0 {
		u.setState(blockGasLimit, blockGasValue)
	}
	return err
}


func (u *ParamManager) hasPermission() bool {
	//在角色合约接口中查询对应角色信息并判断是否有权限
	return checkPermission(u.state, u.callerAddr, 1) || checkPermission(u.state, u.callerAddr, 0)
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
		"setAllowAnyAccountDeployContract": u.setAllowAnyAccountDeployContract,
		"setCheckContractDeployPermission": u.setCheckContractDeployPermission,
		"getCheckContractDeployPermission": u.getCheckContractDeployPermission,
		"getAllowAnyAccountDeployContract": u.getAllowAnyAccountDeployContract,
		"setIsApproveDeployedContract":     u.setIsApproveDeployedContract,
		"getIsApproveDeployedContract":     u.getIsApproveDeployedContract,
		"setIsTxUseGas":                    u.setIsTxUseGas,
		"getIsTxUseGas":                    u.getIsTxUseGas,
	}
}
