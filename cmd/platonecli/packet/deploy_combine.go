package packet

import (
	"errors"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

// DeployCall, used for combining the data of contract deployment
type DeployDataGen struct {
	codeBytes   []byte
	abiBytes    []byte
	TxType      uint64
	Interpreter deployInter
}

// NewDeployCall new a DeployCall object
func NewDeployDataGen(codeBytes, abiBytes []byte, vm string, txType uint64) *DeployDataGen {

	call := &DeployDataGen{
		codeBytes: codeBytes,
		abiBytes:  abiBytes,
		TxType:    txType,
	}

	// set the virtual machine interpreter
	call.SetInterpreter(vm)

	return call
}

// SetInterpreter set the interpreter of DeployCall object
func (call *DeployDataGen) SetInterpreter(vm string) error {
	switch vm {
	case "evm":
		if IsWasmContract(call.codeBytes) {
			// utils.Fatalf("the input  is not evm byte code")
			return errors.New("the input  is not evm byte code")
		}
		call.Interpreter = &EvmDeployInterpreter{codeBytes: call.codeBytes}
	default:
		if !IsWasmContract(call.codeBytes) {
			// utils.Fatalf("the input  is not wasm byte code")
			return errors.New("the input  is not wasm byte code")
		}
		call.Interpreter = &WasmDeployInterpreter{
			codeBytes: call.codeBytes,
			abiBytes:  call.abiBytes,
		}
	}

	return nil
}

// CombineData of DeployCall data struct is used for packeting the data of wasm or evm contracts deployment
// Implement the MessageCallDemo interface
func (call DeployDataGen) CombineData() (string, []string, bool, error) {
	if call.Interpreter == nil {
		return "", nil, false, errors.New("interpreter is not provided")
	}

	data, err := call.Interpreter.combineData()
	return data, nil, true, err
}

func (call *DeployDataGen) GetAbiBytes() []byte {
	return call.abiBytes
}

// combineDeployData packet the data in the way defined by the evm virtual mechine
// Implement the Interpreter interface
func (i *EvmDeployInterpreter) combineData() (string, error) {
	return "0x" + string(i.codeBytes), nil
}

// combineDeployData packet the data in the way defined by the wasm virtual mechine
// Implement the Interpreter interface
func (i *WasmDeployInterpreter) combineData() (string, error) {
	/// utl.Logger.Printf("int wasm combineDeployData()")

	dataParams := make([][]byte, 0)
	dataParams = append(dataParams, common.Int64ToBytes(int64(i.txType)))
	dataParams = append(dataParams, i.codeBytes)
	dataParams = append(dataParams, i.abiBytes)

	return rlpEncode(dataParams)
}
