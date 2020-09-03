package packet

import (
	"errors"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
)

// DeployCall, used for combining the data of contract deployment
type DeployDataGen struct {
	codeBytes         []byte
	abiBytes          []byte
	TxType            uint64
	ConstructorParams []string
	Interpreter       deployInter
}

// NewDeployCall new a DeployCall object
func NewDeployDataGen(codeBytes, abiBytes []byte, consParams []string, vm string, txType uint64) *DeployDataGen {

	dataGen := &DeployDataGen{
		codeBytes:         codeBytes,
		abiBytes:          abiBytes,
		TxType:            txType,
		ConstructorParams: consParams,
	}

	// set the virtual machine interpreter
	err := dataGen.SetInterpreter(vm)
	if err != nil {
		// todo: handle error
	}

	return dataGen
}

func parseAbiConstructor(abiBytes []byte, funcParams []string) ([]byte, error) {
	var abiFunc = new(FuncDesc)

	funcs, err := ParseAbiFromJson(abiBytes)
	if err != nil {
		return nil, err
	}

	for _, value := range funcs {
		if value.Type == "constructor" {
			abiFunc = &value
		}
	}

	// todo: better solution?
	if abiFunc == nil {
		return nil, nil
	}

	conBytes, _, err := EvmStringToEncodeByte(abiFunc, funcParams)
	return conBytes, err
}

// SetInterpreter set the interpreter of DeployCall object
func (dataGen *DeployDataGen) SetInterpreter(vm string) error {
	switch vm {
	case "evm":
		if IsWasmContract(dataGen.codeBytes) {
			// utils.Fatalf("the input  is not evm byte code")
			return errors.New("the input is not evm byte code")
		}

		// todo: code refactory
		consInput, _ := parseAbiConstructor(dataGen.abiBytes, dataGen.ConstructorParams)

		dataGen.Interpreter = &EvmDeployInterpreter{
			codeBytes:        dataGen.codeBytes,
			constructorInput: consInput,
		}

	default:
		if !IsWasmContract(dataGen.codeBytes) {
			// utils.Fatalf("the input  is not wasm byte code")
			return errors.New("the input is not wasm byte code")
		}
		dataGen.Interpreter = &WasmDeployInterpreter{
			codeBytes: dataGen.codeBytes,
			abiBytes:  dataGen.abiBytes,
		}
	}

	return nil
}

func (dataGen *DeployDataGen) ReceiptParsing(receipt *Receipt) *ReceiptParsingReturn {
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

func (dataGen *DeployDataGen) ParseNonConstantResponse(respStr string, outputType []abi.ArgumentMarshaling) []interface{} {
	return nil
}
