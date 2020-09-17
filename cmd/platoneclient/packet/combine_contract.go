package packet

import "github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

type RawData struct {
	funcParams []interface{}
	funcAbi    *FuncDesc
}

// NewData new a RawData object
func NewData(funcParams []interface{}, funcAbi *FuncDesc) *RawData {
	return &RawData{
		funcParams: funcParams,
		funcAbi:    funcAbi,
	}
}

// ContractDataGen is used for combining the data of contract execution
type ContractDataGen struct {
	data   *RawData
	conAbi ContractAbi
	TxType uint64
	/// name   string
	Interp contractInter
}

func NewContractDataGen(data *RawData, conAbi ContractAbi, txType uint64) *ContractDataGen {
	dataGen := &ContractDataGen{
		data:   data,
		conAbi: conAbi,
		TxType: txType,
	}

	return dataGen
}

// SetInterpreter set the interpreter of ContractDataGen object
func (dataGen *ContractDataGen) SetInterpreter(vm, name string, txType uint64) {
	switch vm {
	case "evm":
		dataGen.Interp = &EvmContractInterpreter{}
	// the default interpreter is "wasm"
	default:
		dataGen.Interp = &WasmContractInterpreter{
			cnsName: name,
			txType:  txType,
		}
	}
}

func (dataGen *ContractDataGen) ReceiptParsing(receipt *Receipt) *ReceiptParsingReturn {
	return dataGen.Interp.ReceiptParsingV2(receipt, dataGen.conAbi)
}

// CombineData of Contractcall data struct is used for packeting the data of wasm or evm contracts execution
// Implement the MessageCallDemo interface
func (dataGen *ContractDataGen) CombineData() (string, error) {

	// packet contract method and input parameters
	funcBytes, err := dataGen.combineFunc()
	if err != nil {
		return "", err
	}

	// packet contract data
	return dataGen.combineContractData(funcBytes)
}

// combineFunc of Contractcall data struct is used for combining the
func (dataGen *ContractDataGen) combineFunc() ([][]byte, error) {

	// encode the function and get the function constant
	funcByte, err := dataGen.Interp.encodeFunctionV2(dataGen.data.funcAbi, dataGen.data.funcParams)
	if err != nil {
		return nil, err
	}

	return funcByte, nil
}

// combineContractData selects the interpreter for combining the contract call data
func (dataGen *ContractDataGen) combineContractData(funcBytes [][]byte) (string, error) {
	return dataGen.Interp.combineData(funcBytes)
}

func (dataGen *ContractDataGen) GetIsWrite() bool {
	return dataGen.Interp.setIsWrite(dataGen.data.funcAbi)
}

func (dataGen *ContractDataGen) GetContractDataDen() *ContractDataGen {
	return dataGen
}

func (dataGen *ContractDataGen) GetMethodAbi() *FuncDesc {
	return dataGen.data.funcAbi
}

func (dataGen *ContractDataGen) ParseNonConstantResponse(respStr string, outputType []abi.ArgumentMarshaling) []interface{} {
	return dataGen.Interp.ParseNonConstantResponse(respStr, outputType)
}
