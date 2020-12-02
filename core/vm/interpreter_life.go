package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/math"
	"github.com/PlatONEnetwork/PlatONE-Go/core/lru"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"

	"github.com/PlatONEnetwork/PlatONE-Go/life/exec"
	"github.com/PlatONEnetwork/PlatONE-Go/life/resolver"
	//gomath "math"
)

var (
	errReturnInvalidRlpFormat   = errors.New("interpreter_life: invalid rlp format.")
	errReturnInsufficientParams = errors.New("interpreter_life: invalid input. ele must greater than 2")
	errReturnInvalidAbi         = errors.New("interpreter_life: invalid abi, encoded fail.")
	errFuncNameNotInTheAbis     = errors.New("interpreter_life: the FuncName is not in the Abi list")
)

var DEFAULT_VM_CONFIG = exec.VMConfig{
	EnableJIT:          false,
	DefaultMemoryPages: exec.DefaultMemoryPages,
	DynamicMemoryPages: exec.DynamicMemoryPages,
}

// WASMInterpreter represents an WASM interpreter
type WASMInterpreter struct {
	evm         *EVM
	cfg         Config
	wasmStateDB *WasmStateDB
	WasmLogger  log.Logger
	resolver    exec.ImportResolver
	returnData  []byte
}

// NewWASMInterpreter returns a new instance of the Interpreter
func NewWASMInterpreter(evm *EVM, cfg Config) *WASMInterpreter {

	wasmStateDB := &WasmStateDB{
		StateDB: evm.StateDB,
		evm:     evm,
		cfg:     &cfg,
	}
	return &WASMInterpreter{
		evm:         evm,
		cfg:         cfg,
		WasmLogger:  NewWasmLogger(cfg, log.WasmRoot()),
		wasmStateDB: wasmStateDB,
		resolver:    resolver.NewResolver(0x01),
	}
}

// check if the called functin in the abi
func (in *WASMInterpreter) preCheckFunction(contract *Contract, input []byte, abi []byte) (bool, *common.Address, error) {

	var txData [][]byte
	containFunc := false

	if err := rlp.DecodeBytes(input, &txData); err != nil {
		return containFunc, nil, err
	}
	if len(txData) < 2 {
		return containFunc, nil, errors.New("Too few elements in tx.data")
	}

	funcName := string(txData[1])
	wasmabi := new(utils.WasmAbi)

	err := wasmabi.FromJson(abi)
	if err != nil {
		return containFunc, nil, err
	}

	for _, obj := range wasmabi.AbiArr {
		if obj.Name == funcName {
			containFunc = true
			break
		}
	}
	if !containFunc {
		statedb := NewWasmStateDB(in.wasmStateDB, contract)
		key := []byte("currentManagerName")
		key = append(key, byte(0))
		key = append([]byte{byte(len(key))}, key...)

		cnsManagerBytes := statedb.GetState(key)
		if len(cnsManagerBytes) <= 1 {
			return false, &common.Address{}, nil
		}
		cnsManagerAddr := common.HexToAddress(string(cnsManagerBytes[1:]))

		return false, &cnsManagerAddr, nil
	}

	return containFunc, nil, nil
}

// Run loops and evaluates the contract's code with the given input data and returns.
// the return byte-slice and an error if one occurred
//
// It's important to note that any errors returned by the interpreter should be
// considered a revert-and-consume-all-gas operations except for
// errExecutionReverted which means revert-and-keep-gas-left.
func (in *WASMInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error) {
	defer func() {
		if er := recover(); er != nil {
			ret, err = nil, fmt.Errorf("VM execute fail：%v", er)
		}
	}()
	in.evm.depth++
	defer func() {
		in.evm.depth--
		if in.evm.depth == 0 {
			logger, ok := in.WasmLogger.(*WasmLogger)
			if ok {
				logger.Flush()
			}
		}
	}()

	if len(contract.Code) == 0 {
		return nil, nil
	}
	_, abi, code, er := parseRlpData(contract.Code)
	if er != nil {
		return nil, er
	}

	// 2020.12.2 yzk
	// this is reserved because we depend on it to extract cns info from old version chain(0.9.x)
	if contract.self.Address() == common.HexToAddress("0x0000000000000000000000000000000000000011") {
		containFunction, addr, err := in.preCheckFunction(contract, input, abi)
		if err != nil {
			return nil, err
		}

		if !containFunction {
			contract.self = ContractRef(AccountRef(*addr))
			contract.SetCallCode(addr, in.evm.StateDB.GetCodeHash(*addr), in.evm.StateDB.GetCode(*addr))

			_, abi, code, er = parseRlpData(contract.Code)
			if er != nil {
				return nil, er
			}
		}
	}
	context := &exec.VMContext{
		Config:   DEFAULT_VM_CONFIG,
		Addr:     contract.Address(),
		GasLimit: contract.Gas,
		StateDB:  NewWasmStateDB(in.wasmStateDB, contract),
		Log:      in.WasmLogger,
	}

	var lvm *exec.VirtualMachine
	var module *lru.WasmModule
	module, ok := lru.WasmCache().Get(contract.Address())

	if !ok {
		module = &lru.WasmModule{}
		module.Module, module.FunctionCode, err = exec.ParseModuleAndFunc(code, nil)
		if err != nil {
			return nil, err
		}
		lru.WasmCache().Add(contract.Address(), module)
	}

	lvm, err = exec.NewVirtualMachineWithModule(module.Module, module.FunctionCode, context, in.resolver, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		lvm.Stop()
	}()

	contract.Input = input
	var (
		funcName   string
		txType     int
		params     []int64
		returnType string
	)

	if input == nil {
		funcName = "init" // init function.
	} else {
		// parse input.
		txType, funcName, params, returnType, err = parseInputFromAbi(lvm, input, abi)
		if err != nil {
			if err == errReturnInsufficientParams && txType == 0 { // transfer to contract address.
				return nil, nil
			}
			return nil, err
		}
		if txType == 0 {
			return nil, nil
		}
		if returnType == "float128" || returnType == "uint128" || returnType == "int128" {
			params = append([]int64{resolver.Malloc(lvm, 16)}, params...)
		}
	}
	entryID, ok := lvm.GetFunctionExport(funcName)
	if !ok {
		return nil, fmt.Errorf("entryId not found.")
	}
	if funcName == "init" {
		in.evm.InitEntryID = entryID
	}
	lvm.InitEntryID = in.evm.InitEntryID

	res, err := lvm.RunWithGasLimit(entryID, int(context.GasLimit), params...)
	if err != nil {
		log.Error("RunWithGasLimit error", "err", err.Error())
		return nil, err
	}
	if contract.Gas >= context.GasUsed {
		contract.Gas = contract.Gas - context.GasUsed
	} else {
		return nil, fmt.Errorf("out of gas.")
	}

	if input == nil {
		return contract.Code, nil
	}
	// todo: more type need to be completed
	switch returnType {
	case "void", "int8", "int", "int32", "int64":
		if txType == common.CallContractFlag {
			return utils.Int64ToBytes(res), nil
		}
		bigRes := new(big.Int)
		bigRes.SetInt64(res)
		finalRes := utils.Align32Bytes(math.U256(bigRes).Bytes())
		return finalRes, nil
	case "uint8", "uint16", "uint32", "uint64":
		if txType == common.CallContractFlag {
			return utils.Uint64ToBytes(uint64(res)), nil
		}
		finalRes := utils.Align32Bytes(utils.Uint64ToBytes((uint64(res))))
		return finalRes, nil
	case "float32", "float64":
		bytes := make([]byte, 8)
		binary.BigEndian.PutUint64(bytes, uint64(res))
		if txType == common.CallContractFlag {
			return bytes, nil
		}
		finalRes := utils.Align32Bytes(bytes)
		return finalRes, nil
	case "float128", "uint128", "int128":
		// float128 satisfy IEEE 754 Quadruple precision
		// wo should revert bytes from little edian to big edian
		returnBytes := lvm.Memory.Memory[params[0] : params[0]+16]
		common.RevertBytes(returnBytes)
		if txType == common.CallContractFlag {
			return returnBytes, nil
		}
		returnBytes = utils.Align32Bytes(returnBytes)
		return returnBytes, nil
	case "string", "int128_s", "uint128_s", "int256_s", "uint256_s":
		returnBytes := make([]byte, 0)
		copyData := lvm.Memory.Memory[res:]
		for _, v := range copyData {
			if v == 0 {
				break
			}
			returnBytes = append(returnBytes, v)
		}
		if txType == common.CallContractFlag || txType == common.TxTypeCallSollCompatibleWasm {
			return returnBytes, nil
		}
		strHash := common.BytesToHash(common.Int32ToBytes(32))
		sizeHash := common.BytesToHash(common.Int64ToBytes(int64((len(returnBytes)))))
		var dataRealSize = len(returnBytes)
		if (dataRealSize % 32) != 0 {
			dataRealSize = dataRealSize + (32 - (dataRealSize % 32))
		}
		dataByt := make([]byte, dataRealSize)
		copy(dataByt[0:], returnBytes)

		finalData := make([]byte, 0)
		finalData = append(finalData, strHash.Bytes()...)
		finalData = append(finalData, sizeHash.Bytes()...)
		finalData = append(finalData, dataByt...)

		return finalData, nil
	}
	return nil, nil
}

// CanRun tells if the contract, passed as an argument, can be run
// by the current interpreter
func (in *WASMInterpreter) CanRun(code, input []byte, contract *Contract) (bool, []byte) {
	if !strings.EqualFold(common.GetCurrentInterpreterType(), "all") {
		return true, input
	}

	// Handling internal calls
	addr := common.Address{}
	if contract.Caller() == addr {
		return true, input
	}

	// Handling non-wasm contracts, unable to execute.
	if ok, _, _, _ := common.IsWasmContractCode(code); !ok {
		return false, input
	}

	// Extra processing delegate call
	if contract.DelegateCall {
		return true, input
	}

	// Handling user calls
	callerCode := in.wasmStateDB.StateDB.GetCode(contract.Caller())
	if callerCode == nil {
		return true, input
	}

	// Handling caller is wasm contracts
	if ok, _, _, _ := common.IsWasmContractCode(callerCode); ok {
		return true, input
	}
	// Handling the sol contract call wasm contract
	var (
		wasmInput []byte
		err       error
	)
	if wasmInput, err = abi.GenerateInputData(&abi.WasmInput{}, input); err != nil {
		return false, input
	}
	return true, wasmInput
}

// parse input(payload)
func parseInputFromAbi(vm *exec.VirtualMachine, input []byte, abi []byte) (txType int, funcName string, params []int64, returnType string, err error) {
	if input == nil || len(input) <= 1 {
		return -1, "", nil, "", fmt.Errorf("invalid input.")
	}
	// [txType][funcName][args1][args2]
	// rlp decode
	ptr := new(interface{})
	err = rlp.Decode(bytes.NewReader(input), &ptr)
	if err != nil {
		return -1, "", nil, "", err
	}
	rlpList := reflect.ValueOf(ptr).Elem().Interface()

	if _, ok := rlpList.([]interface{}); !ok {
		return -1, "", nil, "", errReturnInvalidRlpFormat
	}

	iRlpList := rlpList.([]interface{})
	if len(iRlpList) < 2 {
		if len(iRlpList) != 0 {
			if v, ok := iRlpList[0].([]byte); ok {
				txType = int(common.BytesToInt64(v))
			}
		} else {
			txType = -1
		}
		return txType, "", nil, "", errReturnInsufficientParams
	}

	wasmabi := new(utils.WasmAbi)
	err = wasmabi.FromJson(abi)
	if err != nil {
		return -1, "", nil, "", errReturnInvalidAbi
	}

	params = make([]int64, 0)
	if v, ok := iRlpList[0].([]byte); ok {
		txType = int(common.BytesToInt64(v))
	}
	if v, ok := iRlpList[1].([]byte); ok {
		funcName = string(v)
	}
	isFuncNameInTheAbis := false
	var args []utils.InputParam
	for _, v := range wasmabi.AbiArr {
		if strings.EqualFold(funcName, v.Name) && strings.EqualFold(v.Type, "function") {
			args = v.Inputs
			if len(v.Outputs) != 0 {
				returnType = v.Outputs[0].Type
			} else {
				returnType = "void"
			}
			isFuncNameInTheAbis = true
			break
		}
	}

	if !isFuncNameInTheAbis {
		return -1, "", nil, "", errFuncNameNotInTheAbis
	}
	argsRlp := iRlpList[2:]
	if len(args) != len(argsRlp) {
		return -1, "", nil, returnType, fmt.Errorf("invalid input or invalid abi.")
	}
	// uint64 uint32  uint16 uint8 int64 int32  int16 int8 float32 float64 string void
	for i, v := range args {
		bts := argsRlp[i].([]byte)
		switch v.Type {
		case "string", "int128_s", "uint128_s", "int256_s", "uint256_s":
			pos := resolver.MallocString(vm, string(bts))
			params = append(params, pos)
		case "int8":
			if len(bts) > 1 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 1 byte but got %d bytes", len(bts))
			}
			params = append(params, int64(bts[0]))
		case "int16":
			if len(bts) > 2 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 2 byte2 but got %d bytes", len(bts))
			}
			params = append(params, int64(binary.BigEndian.Uint16(bts)))
		case "int32", "int":
			if len(bts) > 4 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 4 bytes but got %d bytes", len(bts))
			}
			params = append(params, int64(binary.BigEndian.Uint32(bts)))
		case "int64":
			if len(bts) > 8 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 8 bytes but got %d bytes", len(bts))
			}
			params = append(params, int64(binary.BigEndian.Uint64(bts)))
		case "uint8":
			if len(bts) > 1 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 1 byte but got %d bytes", len(bts))
			}
			params = append(params, int64(bts[0]))
		case "uint32", "uint":
			if len(bts) > 4 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 4 bytes but got %d bytes", len(bts))
			}
			params = append(params, int64(binary.BigEndian.Uint32(bts)))
		case "uint64":
			if len(bts) > 8 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 8 bytes but got %d bytes", len(bts))
			}
			params = append(params, int64(binary.BigEndian.Uint64(bts)))
		case "float32":
			if len(bts) > 4 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 4 bytes but got %d bytes", len(bts))
			}
			//bits bits is the floating-point number corresponding to the IEEE 754 binary representation bts
			bits := binary.BigEndian.Uint32(bts)
			params = append(params, int64(bits))
		case "float64":
			if len(bts) > 8 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 8 bytes but got %d bytes", len(bts))
			}
			bits := binary.BigEndian.Uint64(bts)
			params = append(params, int64(bits))
		case "float128", "int128", "uint128":
			if len(bts) != 16 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 16 bytes but got %d bytes", len(bts))
			}
			// wasm is little edian
			params = append(params, int64(binary.BigEndian.Uint64(bts[8:])), int64(binary.BigEndian.Uint64(bts[:8])))
		case "bool":
			if len(bts) > 1 {
				return -1, "", nil, returnType, fmt.Errorf("invalid parameter: want 1 byte but got %d bytes", len(bts))
			}
			params = append(params, int64(bts[0]))
		default:
			return -1, "", nil, returnType, fmt.Errorf("unexpected parameter type: %s", v.Type)
		}
	}
	return txType, funcName, params, returnType, nil
}

// rlpData=RLP([txType][code][abi])
func parseRlpData(rlpData []byte) (int64, []byte, []byte, error) {
	ptr := new(interface{})
	err := rlp.Decode(bytes.NewReader(rlpData), &ptr)
	if err != nil {
		return -1, nil, nil, err
	}
	rlpList := reflect.ValueOf(ptr).Elem().Interface()

	if _, ok := rlpList.([]interface{}); !ok {
		return -1, nil, nil, fmt.Errorf("invalid rlp format.")
	}

	iRlpList := rlpList.([]interface{})
	if len(iRlpList) <= 2 {
		return -1, nil, nil, fmt.Errorf("invalid input. ele must greater than 2")
	}
	var (
		txType int64
		code   []byte
		abi    []byte
	)
	if v, ok := iRlpList[0].([]byte); ok {
		txType = utils.BytesToInt64(v)
	}
	if v, ok := iRlpList[1].([]byte); ok {
		code = v
	}
	if v, ok := iRlpList[2].([]byte); ok {
		abi = v
	}
	return txType, abi, code, nil
}
