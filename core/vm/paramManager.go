package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)
var (
	gasContractNameKey = "GasContractName"
	isProduceEmptyBlockKey = "IsProduceEmptyBlock"
	txGasLimitKey = "TxGasLimit"
	blockGasLimitKey = "BlockGasLimit"
	isAllowAnyAccountDeployContractKey = "IsAllowAnyAccountDeployContract"
	isCheckContractDeployPermission="isCheckContractDeployPermission"
	isApproveDeployedContractKey = "IsApproveDeployedContract"
	isTxUseGasKey = "IsTxUseGas"
	//cbftTimeParamKey = "CBFTTimeParam"
)
const (
	txGasLimitMinValue uint64 = 12771596*100// 12771596 大致相当于 0.012772s
	txGasLimitMaxValue uint64 = 2e9      // 相当于 2s
	txGasLimitDefaultValue = 1.5e9        // 相当于 1.5s
	blockGasLimitMinValue uint64 = 12771596*100 // 12771596 大致相当于 0.012772s
	blockGasLimitMaxValue uint64 = 2e10         // 相当于 20s
	//blockGasLimitDefaultValue uint64 = 1e10        // 相当于 10s
	//produceDurationMaxValue = 60
	//produceDurationDefaultValue = 10
	//blockIntervalMinValue = 1
	//blockIntervalDefaultValue = 1
)

const (
	SUC = "0"
	FAL = "1"
)
type ParamManager struct {
	StateDB
	CodeAddr	*common.Address
}

func  encode(i interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(i)
}

func (u *ParamManager) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.ParamManagerGas
}

func (u *ParamManager) Run(input []byte) ([]byte, error) {
	gasLimit, err := rlp.EncodeToBytes(txGasLimitKey)
	if err!=nil {
		return nil,err
	}
	gasValue, err := rlp.EncodeToBytes(txGasLimitDefaultValue)
	if err!=nil {
		return nil,err
	}

	if u.getState(gasLimit) == nil {
		u.setState(gasLimit,gasValue)
	}
	return execSC(input, u.AllExportFns())
}

func (u *ParamManager) setState(key, value []byte) {
	u.StateDB.SetState(*u.CodeAddr, key, value)
}

func (u *ParamManager) getState(key []byte) []byte {
	return u.StateDB.GetState(*u.CodeAddr, key)
}

func (u *ParamManager) setGasContractName (contractName string) ([]byte, error){
	if !u.hasPermission() {
		//event here
		return encode(FAL)
	}
	if len(contractName) == 0 {
		//event here
		return encode(FAL)
	}
	key, err := rlp.EncodeToBytes(gasContractNameKey)
	if err!=nil {
		return nil,err
	}
	value, err := rlp.EncodeToBytes(contractName)
	if err!=nil {
		return nil,err
	}
	u.setState(key, value)
	return encode(SUC)
}

func (u *ParamManager) getGasContractName() ([]byte, error){
	key, err := encode(gasContractNameKey)
	if err!=nil {
		return nil,err
	}

	contractName := u.getState(key)
	return encode(contractName)
}

func (u *ParamManager) setIsProduceEmptyBlock(isProduceEmptyBlock uint8) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return encode(FAL)
	}
	if 0 != isProduceEmptyBlock && 1 != isProduceEmptyBlock {
		//event
		return encode(FAL)
	}
	key, err := encode(isProduceEmptyBlockKey)
	if err!=nil {
		return nil,err
	}
	value, err := encode(isProduceEmptyBlock)
	if err!=nil {
		return nil,err
	}
	u.setState(key, value)
	//event here
	return encode(SUC)
}

func (u *ParamManager) getIsProduceEmptyBlock() ([]byte, error){
	key, err := encode(isProduceEmptyBlockKey)
	if err!=nil {
		return nil,err
	}
	isProduceEmptyBlock := u.getState(key)
	return encode(isProduceEmptyBlock)
}

func (u *ParamManager) setTxGasLimit(txGasLimit uint64) ([]byte, error) {
	if !u.hasPermission() {
		//event here
		return encode(FAL)
	}
	if txGasLimit < txGasLimitMinValue || txGasLimit > txGasLimitMaxValue {
		//event
		return encode(FAL)
	}
	// 获取区块 gas limit，其值应大于或等于每笔交易 gas limit 参数的值
	key, err := encode(blockGasLimitKey)
	if err!=nil {
		return nil,err
	}
	blockGasLimit1 := u.getState(key)
	var blockGasLimit uint64
	if err := rlp.DecodeBytes(blockGasLimit1, blockGasLimit); nil != err {
		return nil, err
	}
	if txGasLimit > blockGasLimit{
		//event
		return encode(FAL)
	}
	key, err = encode(txGasLimitKey)
	if err!=nil {
		return nil,err
	}
	value, err := encode(txGasLimit)
	if err!=nil {
		return nil,err
	}
	u.setState(key, value)
	//event
	return  encode(SUC)
}

func (u *ParamManager) getTxGasLimit() ([]byte, error){
	key, err := encode(txGasLimitKey)
	if err!=nil {
		return nil,err
	}
	txGasLimit := u.getState(key)
	return encode(txGasLimit)
}

func (u *ParamManager) setBlockGasLimit(blockGasLimit uint64) ([]byte, error){
	if !u.hasPermission() {
		//event here
		return encode(FAL)
	}
	if blockGasLimit < blockGasLimitMinValue || blockGasLimit > blockGasLimitMaxValue {
		//event
		return encode(FAL)
	}
	key, err := encode(txGasLimitKey)
	if err!=nil {
		return nil,err
	}
	txGasLimit1 := u.getState(key)
	var txGasLimit uint64
	if err := rlp.DecodeBytes(txGasLimit1, txGasLimit); nil != err {
		return nil, err
	}
	if  txGasLimit > blockGasLimit{
		//event
		return encode(FAL)
	}
	key, err = encode(blockGasLimitKey)
	if err!=nil {
		return nil,err
	}
	value, err := encode(blockGasLimit)
	if err!=nil {
		return nil,err
	}
	u.setState(key, value)
	//event here
	return encode(SUC)
}

// 获取区块 gaslimit
func (u *ParamManager) getBlockGasLimit() ([]byte, error) {
	key ,err := encode(blockGasLimitKey)
	if err!=nil {
		return nil,err
	}
	blockGasLimit := u.getState(key)

	return encode(blockGasLimit)
}

// 设置是否允许任意用户部署合约
// @isAllowAnyAccountDeployContract:
// 0: 允许任意用户部署合约  1: 用户具有相应权限才可以部署合约
// 默认为0，即允许任意用户部署合约
func (u *ParamManager) setAllowAnyAccountDeployContract(isAllowAnyAccountDeployContract uint8) ([]byte, error){
	if !u.hasPermission() {
		//event here
		return encode(FAL)
	}
	if 0 != isAllowAnyAccountDeployContract && 1 != isAllowAnyAccountDeployContract {
		//event here
		return encode(FAL)
	}
	key, err := encode(isAllowAnyAccountDeployContractKey)
	if err!=nil {
		return nil,err
	}
	value, err := encode(isAllowAnyAccountDeployContract)
	if err!=nil {
		return nil,err
	}
	u.setState(key, value)
	//event here
	return encode(SUC)
}

// 设置是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *ParamManager) setCheckContractDeployPermission(checkPermission uint8) ([]byte, error){
	if !u.hasPermission() {
		//event here
		return encode(FAL)
	}
	if 0 != checkPermission && 1 != checkPermission {
		//event here
		return encode(FAL)
	}
	key, err := encode(isCheckContractDeployPermission)
	if err!=nil {
		return nil,err
	}
	value, err := encode(checkPermission)
	if err!=nil {
		return nil,err
	}
	u.setState(key, value)
	//event here
	return encode(SUC)
}

// 获取是否是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *ParamManager) getCheckContractDeployPermission() ([]byte, error) {
	key ,err := encode(isCheckContractDeployPermission)
	if err!=nil {
		return nil,err
	}
	checkPermission := u.getState(key)

	return encode(checkPermission)
}

// 获取是否允许任意用户部署合约的标志
func (u *ParamManager) getAllowAnyAccountDeployContract() ([]byte, error) {
	key ,err := encode(isAllowAnyAccountDeployContractKey)
	if err!=nil {
		return nil,err
	}
	isAllowAnyAccountDeployContract := u.getState(key)

	return encode(isAllowAnyAccountDeployContract)
}

// 设置是否审核已部署的合约
// @isApproveDeployedContract:
// 1: 审核已部署的合约  0: 不审核已部署的合约
func (u *ParamManager) setIsApproveDeployedContract(isApproveDeployedContract uint8) ([]byte, error){
	if !u.hasPermission() {
		//event here
		return encode(FAL)
	}
	if 0 != isApproveDeployedContract && 1 != isApproveDeployedContract {
		//event here
		return encode(FAL)
	}
	key, err := encode(isApproveDeployedContractKey)
	if err!=nil {
		return nil,err
	}
	value, err := encode(isApproveDeployedContract)
	if err!=nil {
		return nil,err
	}
	u.setState(key, value)
	//event here
	return encode(SUC)
}

// 获取是否审核已部署的合约的标志
func (u *ParamManager) getIsApproveDeployedContract() ([]byte, error) {
	key ,err := encode(isApproveDeployedContractKey)
	if err!=nil {
		return nil,err
	}
	isApproveDeployedContract := u.getState(key)

	return encode(isApproveDeployedContract)
}

// 本参数根据最新的讨论（2019.03.06之前）不再需要，即交易需要消耗gas。但是计费相关如消耗特定合约代币的参数由 setGasContractName 进行设置
// 设置交易是否消耗 gas
// @isTxUseGas:
// 1:  交易消耗 gas  0: 交易不消耗 gas
func (u *ParamManager) setIsTxUseGas(isTxUseGas uint8) ([]byte, error){
	if !u.hasPermission() {
		//event here
		return encode(FAL)
	}
	if 0 != isTxUseGas && 1 != isTxUseGas {
		//event here
		return encode(FAL)
	}
	key, err := encode(isTxUseGasKey)
	if err!=nil {
		return nil,err
	}
	value, err := encode(isTxUseGas)
	if err!=nil {
		return nil,err
	}
	u.setState(key, value)
	//event here
	return encode(SUC)
}

// 获取交易是否消耗 gas
func (u *ParamManager) getIsTxUseGas() ([]byte, error) {
	key ,err := encode(isTxUseGasKey)
	if err!=nil {
		return nil,err
	}
	isTxUseGas := u.getState(key)

	return encode(isTxUseGas)
}

func (u *ParamManager)hasPermission() bool{
	//在角色合约接口中查询对应角色信息并判断是否有权限
	return true
}

//for access control
func (u *ParamManager) AllExportFns() SCExportFns {
	return SCExportFns{
		"setGasContractName": u.setGasContractName,
		"getGasContractName": u.getGasContractName,
		"setIsProduceEmptyBlock": u.setIsProduceEmptyBlock,
		"getIsProduceEmptyBlock": u.getIsProduceEmptyBlock,
		"setTxGasLimit": u.setTxGasLimit,
		"getTxGasLimit": u.getTxGasLimit,
		"setBlockGasLimit": u.setBlockGasLimit,
		"getBlockGasLimit": u.getBlockGasLimit,
		"setAllowAnyAccountDeployContract": u.setAllowAnyAccountDeployContract,
		"setCheckContractDeployPermission": u.setCheckContractDeployPermission,
		"getCheckContractDeployPermission": u.getCheckContractDeployPermission,
		"getAllowAnyAccountDeployContract": u.getAllowAnyAccountDeployContract,
		"setIsApproveDeployedContract": u.setIsApproveDeployedContract,
		"getIsApproveDeployedContract": u.getIsApproveDeployedContract,
		"setIsTxUseGas": u.setIsTxUseGas,
		"getIsTxUseGas": u.getIsTxUseGas,
		
	}
}
