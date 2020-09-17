package packet

import (
	"errors"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
)

// DeployCall, used for combining the data of contract deployment
type DeployDataGen struct {
	/// codeBytes         []byte
	/// abiBytes          []byte
	TxType uint64
	/// ConstructorParams []string
	Interpreter deployInter

	conAbi ContractAbi
}

// NewDeployCall new a DeployCall object
func NewDeployDataGen(conAbi ContractAbi, txType uint64) *DeployDataGen {
	var dataGen = new(DeployDataGen)
	dataGen.conAbi = conAbi

	return dataGen
}

func parseAbiConstructor(abiBytes []byte, funcParams []string) ([]byte, error) {
	var abiFunc *FuncDesc

	funcs, err := ParseAbiFromJson(abiBytes)
	if err != nil {
		return nil, err
	}

	for _, value := range funcs {
		if value.Type == "constructor" {
			abiFunc = value
			break
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
func (dataGen *DeployDataGen) SetInterpreter(vm string, abiBytes, codeBytes []byte, consParams []interface{}, methodAbi *FuncDesc) {

	if IsWasmContract(codeBytes) {
		// utils.Fatalf("the input  is not evm byte code")
		// return errors.New("the input is not evm byte code")
		vm = "wasm"
	}

	switch vm {
	case "evm":
		dataGen.Interpreter = &EvmDeployInterpreter{
			codeBytes:        codeBytes,
			constructorInput: consParams,
			constructorAbi:   methodAbi,
		}
	// the default interpreter is wasm
	default:
		dataGen.Interpreter = &WasmDeployInterpreter{
			codeBytes: codeBytes,
			abiBytes:  abiBytes,
		}
	}
}

func (dataGen *DeployDataGen) ReceiptParsing(receipt *Receipt) *ReceiptParsingReturn {
	return dataGen.Interpreter.ReceiptParsingV2(receipt, dataGen.conAbi)
}

// CombineData of DeployCall data struct is used for packeting the data of wasm or evm contracts deployment
// Implement the MessageCallDemo interface
func (dataGen DeployDataGen) CombineData() (string, error) {
	if dataGen.Interpreter == nil {
		return "", errors.New("interpreter is not provided")
	}

	return dataGen.Interpreter.combineData()
}

func (dataGen *DeployDataGen) GetIsWrite() bool {
	return true
}

func (dataGen *DeployDataGen) GetContractDataDen() *ContractDataGen {
	return nil
}

func (dataGen *DeployDataGen) ParseNonConstantResponse(respStr string, outputType []abi.ArgumentMarshaling) []interface{} {
	return nil
}
