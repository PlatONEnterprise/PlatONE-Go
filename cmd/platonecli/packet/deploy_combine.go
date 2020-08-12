package packet

import (
	"errors"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
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

	dataGen := &DeployDataGen{
		codeBytes: codeBytes,
		abiBytes:  abiBytes,
		TxType:    txType,
	}

	// set the virtual machine interpreter
	dataGen.SetInterpreter(vm)

	return dataGen
}

// SetInterpreter set the interpreter of DeployCall object
func (dataGen *DeployDataGen) SetInterpreter(vm string) error {
	switch vm {
	case "evm":
		if IsWasmContract(dataGen.codeBytes) {
			// utils.Fatalf("the input  is not evm byte code")
			return errors.New("the input  is not evm byte code")
		}
		dataGen.Interpreter = &EvmDeployInterpreter{codeBytes: dataGen.codeBytes}
	default:
		if !IsWasmContract(dataGen.codeBytes) {
			// utils.Fatalf("the input  is not wasm byte code")
			return errors.New("the input  is not wasm byte code")
		}
		dataGen.Interpreter = &WasmDeployInterpreter{
			codeBytes: dataGen.codeBytes,
			abiBytes:  dataGen.abiBytes,
		}
	}

	return nil
}

func (dataGen *DeployDataGen) ReceiptParsing(receipt *Receipt) string {
	return dataGen.Interpreter.ReceiptParsing(receipt, dataGen.abiBytes)
}

// CombineData of DeployCall data struct is used for packeting the data of wasm or evm contracts deployment
// Implement the MessageCallDemo interface
func (dataGen DeployDataGen) CombineData() (string, []abi.ArgumentMarshaling, bool, error) {
	if dataGen.Interpreter == nil {
		return "", nil, false, errors.New("interpreter is not provided")
	}

	data, err := dataGen.Interpreter.combineData()
	return data, nil, true, err
}

func (dataGen *DeployDataGen) GetAbiBytes() []byte {
	return dataGen.abiBytes
}

func (dataGen *DeployDataGen) ParseNonConstantResponse(respStr string, outputType []abi.ArgumentMarshaling) []interface{} {
	return nil
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
