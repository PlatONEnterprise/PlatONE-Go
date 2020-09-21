package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	math2 "github.com/PlatONEnetwork/PlatONE-Go/common/math"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/ethdb"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

var abi_ = `{
	"version": "0.01",
	"abi": [{
			"method": "transfer",
			"args": [{
					"name": "from",
					"typeName": "Address",
					"realTypeName": "string"
				}, {
					"name": "to",
					"typeName": "address",
					"realTypeName": "string"
				}, {
					"name": "asset",
					"typeName": "",
					"realTypeName": "int64"
				}
			]
		}
	]
}`

func TestAddressUtil(t *testing.T) {
	ref := ContractRefSelf{}
	addr := ref.Address()
	fmt.Println(addr.Hex())
}

func TestWasmInterpreter(t *testing.T) {

	evm := &EVM{
		StateDB: stateDB{},
		Context: Context{
			GasLimit:    1000000,
			BlockNumber: big.NewInt(10),
		},
	}
	cfg := Config{}

	wasmInterpreter := NewWASMInterpreter(evm, cfg)

	code, _ := ioutil.ReadFile("..\\..\\life\\contract\\hello.wasm")

	contract := &Contract{
		CallerAddress: common.BigToAddress(big.NewInt(88888)),
		caller:        ContractRefCaller{},
		self:          ContractRefSelf{},
		Code:          code,
		Gas:           1000000,
		ABI:           []byte(abi_),
	}
	// build input, {1}{transfer}{from}{to}{asset}
	input := genInput()
	wasmInterpreter.Run(contract, input, true)

}

func Int64ToBytes(n int64) []byte {
	tmp := int64(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func BenchmarkWasmInterpreter_SetState(bench *testing.B) {
	db, _ := ethdb.NewLDBDatabase("./data.getsettest", 0, 0)

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if nil != err {
		bench.Fatal(err)
	}

	wasmRun(bench, statedb, "Set", 10000)
}

func BenchmarkWasmInterpreter_GetState(bench *testing.B) {
	db, _ := ethdb.NewLDBDatabase("./data.getsettest", 0, 0)

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if nil != err {
		bench.Fatal(err)
	}

	wasmRun(bench, statedb, "Get", 10000)
}

func BenchmarkWasmInterpreter_GetState_MockStateDB(bench *testing.B) {
	wasmRun(bench, stateDB{}, "Get", 10000)
}

func BenchmarkWasmInterpreter_SetState_MockStateDB(bench *testing.B) {
	wasmRun(bench, stateDB{}, "Set", 10000)
}

func BenchmarkWasmInterpreter_GetState_FixedInput(bench *testing.B) {
	db, _ := ethdb.NewLDBDatabase("./data.getsettest", 0, 0)

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if nil != err {
		bench.Fatal(err)
	}

	wasmRun(bench, statedb, "GetFixed", 10000)
}

func BenchmarkWasmInterpreter_SetState_FixedInput(bench *testing.B) {
	db, _ := ethdb.NewLDBDatabase("./data.getsettest", 0, 0)

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if nil != err {
		bench.Fatal(err)
	}

	wasmRun(bench, statedb, "SetFixed", 10000)
}

func BenchmarkWasmInterpreter_Deploy(bench *testing.B) {
	db, _ := ethdb.NewLDBDatabase("./data.getsettest", 0, 0)

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if nil != err {
		bench.Fatal(err)
	}

	wasmRun(bench, statedb, "Deploy", 10000)
}

func BenchmarkWasmInterpreter_GetState_FixedInput_MockStateDB(bench *testing.B) {
	wasmRun(bench, stateDB{}, "GetFixed", 10000)
}

func BenchmarkWasmInterpreter_SetState_FixedInput_MockStateDB(bench *testing.B) {
	wasmRun(bench, stateDB{}, "SetFixed", 10000)
}

func BenchmarkWasmInterpreter_Deploy_MockStateDB(bench *testing.B) {
	wasmRun(bench, stateDB{}, "Deploy", 10000)
}

func wasmRun(bench *testing.B, statedb stateDBer, inputKind string, prepareCount int) {
	address := common.HexToAddress("0x823140710bf13990e4500136726d8b55")
	codeBytes, err := ioutil.ReadFile("../../life/contract/getsettest.wasm")
	if nil != err {
		bench.Fatal(err)
	}

	abiBytes, err := ioutil.ReadFile("../../life/contract/getsettest.cpp.abi.json")
	if nil != err {
		bench.Fatal(err)
	}

	param := [3][]byte{
		Int64ToBytes(1),
		codeBytes,
		abiBytes,
	}
	code, err := rlp.EncodeToBytes(param)
	if err != nil {
		bench.Fatal(err)
	}

	statedb.SetState(address, crypto.Keccak256Hash(code).Bytes(), code)
	_, err = statedb.Commit(false)
	if nil != err {
		bench.Error(err)
	}

	evm := &EVM{
		StateDB: statedb,
		Context: Context{
			GasLimit:    1000000,
			BlockNumber: big.NewInt(10),
		},
	}
	cfg := Config{}

	wasmInterpreter := NewWASMInterpreter(evm, cfg)

	contract := &Contract{
		CallerAddress: common.BigToAddress(big.NewInt(rand.Int63())),
		caller:        ContractRefCaller{},
		self:          ContractRefSelf{},
		Code:          code,
		Gas:           99999999999999999,
		ABI:           []byte(abi_),
	}

	for i := 0; i < prepareCount; i++ {
		input := genSetInput()

		_, err := wasmInterpreter.Run(contract, input, true)
		if nil != err {
			bench.Fatal(err)
		}

		_, err = statedb.Commit(false)
		if nil != err {
			bench.Fatal(err)
		}
	}

	bench.ResetTimer()

	for i := 0; i < bench.N; i++ {
		var input []byte
		switch inputKind {
		case "SetFixed":
			input = genSetFixedInput()
		case "GetFixed":
			input = genGetFixedInput()
		case "Set":
			input = genSetInput()
		case "Get":
			input = genGetInput()
		case "Deploy":
			input = nil
		}
		_, err := wasmInterpreter.Run(contract, input, true)
		if nil != err {
			bench.Fatal(err)
		}
		//bench.Log(ret)
		_, err = statedb.Commit(false)
		if nil != err {
			bench.Fatal(err)
		}
	}
}

type stateDBer interface {
	StateDB
	Commit(deleteEmptyObjects bool) (root common.Hash, err error)
}

type stateDB struct {
	StateDB
}

func (stateDB) CreateAccount(common.Address) {}

func (stateDB) SubBalance(common.Address, *big.Int) {}
func (stateDB) AddBalance(common.Address, *big.Int) {}
func (stateDB) GetBalance(common.Address) *big.Int  { return nil }

func (stateDB) GetNonce(common.Address) uint64  { return 0 }
func (stateDB) SetNonce(common.Address, uint64) {}

func (stateDB) GetCodeHash(common.Address) common.Hash { return common.Hash{} }
func (stateDB) GetCode(common.Address) []byte          { return nil }
func (stateDB) SetCode(common.Address, []byte)         {}
func (stateDB) GetCodeSize(common.Address) int         { return 0 }

// todo: new func for abi of contract.
func (stateDB) GetAbiHash(common.Address) common.Hash { return common.Hash{} }
func (stateDB) GetAbi(common.Address) []byte          { return nil }
func (stateDB) SetAbi(common.Address, []byte)         {}

func (stateDB) AddRefund(uint64)  {}
func (stateDB) SubRefund(uint64)  {}
func (stateDB) GetRefund() uint64 { return 0 }

func (stateDB) GetCommittedState(common.Address, []byte) []byte { return nil }
func (stateDB) GetState(common.Address, []byte) []byte          { return []byte("world+++++++**") }
func (stateDB) SetState(common.Address, []byte, []byte)         {}
func (stateDB) Suicide(common.Address) bool                     { return true }
func (stateDB) HasSuicided(common.Address) bool                 { return true }

// Exist reports whether the given account exists in state.
// Notably this should also return true for suicided accounts.
func (stateDB) Exist(common.Address) bool { return true }

// Empty returns whether the given account is empty. Empty
// is defined according to EIP161 (balance = nonce = code = 0).
func (stateDB) Empty(common.Address) bool { return true }

func (stateDB) RevertToSnapshot(int) {}
func (stateDB) Snapshot() int        { return 0 }

func (stateDB) AddPreimage(common.Hash, []byte) {}

func (stateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {}

func (stateDB) AddLog(*types.Log) {
	fmt.Println("add log")
}
func (stateDB) Commit(deleteEmptyObjects bool) (root common.Hash, err error) {
	return common.Hash{}, nil
}

func RunContractAndGetRes(codePath, abiPath string, input [][]byte) ([]byte, error) {
	//gen contract codes
	//this contract accepts a float64 and return it by add 1
	codeBytes, err := ioutil.ReadFile(codePath)
	if nil != err {
		return []byte{}, err
	}

	abiBytes, err := ioutil.ReadFile(abiPath)
	if nil != err {
		return []byte{}, err
	}

	param := [3][]byte{
		Int64ToBytes(2),
		codeBytes,
		abiBytes,
	}
	code, err := rlp.EncodeToBytes(param)
	if err != nil {
		return []byte{}, err
	}
	evm := &EVM{
		StateDB: stateDB{},
		Context: Context{
			GasLimit:    1000000,
			BlockNumber: big.NewInt(10),
		},
	}
	cfg := Config{}

	wasmInterpreter := NewWASMInterpreter(evm, cfg)

	contract := &Contract{
		CallerAddress: common.BigToAddress(big.NewInt(88888)),
		caller:        ContractRefCaller{},
		self:          ContractRefSelf{},
		Code:          code,
		Gas:           math.MaxUint64,
	}

	buffer := new(bytes.Buffer)
	rlp.Encode(buffer, input)

	return wasmInterpreter.Run(contract, buffer.Bytes(), true)
}

func TestFloatInputAndOutput(t *testing.T) {
	codePath := "../../life/contract/numberstest.wasm"
	abiPath := "../../life/contract/numberstest.abi.json"

	// build input, {1}{getNum}{num}
	var input [][]byte
	input = make([][]byte, 0)
	input = append(input, common.Int64ToBytes(1))
	input = append(input, []byte("addDouble"))
	array := []float64{123456789123456789123456789.12345678901234567890123456789, 123456789123456789123456789.12345678901234567890123456789}

	for _, f := range array {
		input = append(input, common.Float64ToBytes(f))
	}

	ret, err := RunContractAndGetRes(codePath, abiPath, input)
	if err != nil {
		t.Fatal(err)
	}
	retInFloat := common.BytesToFloat64(ret[len(ret)-8:])

	if retInFloat != array[0]+array[1] {
		t.Fatal("result not correct")
	}
	fmt.Println(common.BytesToFloat64(ret))
}

func TestFloat128(t *testing.T) {
	codePath := "../../life/contract/numberstest.wasm"
	abiPath := "../../life/contract/numberstest.abi.json"

	var input [][]byte
	input = make([][]byte, 0)
	input = append(input, common.Int64ToBytes(1))
	input = append(input, []byte("addLongDouble"))

	// build input, {1}{getNum}{num,num}
	array := []string{"123456789123456789123456789.12345678901234567890123456789", "123456789123456789123456789.12345678901234567890123456789"}
	for _, f := range array {
		F, _, _ := big.ParseFloat(f, 10, 113, big.ToNearestEven)
		fmt.Println(F)
		F128, _ := math2.NewFromBig(F)

		bytes := append(common.Uint64ToBytes(F128.High()), common.Uint64ToBytes(F128.Low())...)
		input = append(input, bytes)
	}

	ret, err := RunContractAndGetRes(codePath, abiPath, input)
	if err != nil {
		t.Fatal(err)
	}
	high := binary.BigEndian.Uint64(ret[len(ret)-16 : len(ret)-8])
	low := binary.BigEndian.Uint64(ret[len(ret)-8:])

	FR := math2.NewFromBits(high, low)
	FB, _ := FR.Big()
	fmt.Println(FB)
}

func TestInt128(t *testing.T) {
	codePath := "../../life/contract/numberstest.wasm"
	abiPath := "../../life/contract/numberstest.abi.json"

	var input [][]byte
	input = make([][]byte, 0)
	input = append(input, common.Int64ToBytes(1))
	input = append(input, []byte("addLong"))

	// build input, {1}{getNum}{num,num}
	array := []string{"153265412365478951234569874125632", "153265412365478951234569874125632"}
	for _, integer := range array {
		I, _ := new(big.Int).SetString(integer, 10)
		b, _ := common.BigToByte128(I)

		input = append(input, b)
	}

	ret, err := RunContractAndGetRes(codePath, abiPath, input)
	if err != nil {
		t.Fatal(err)
	}

	r := common.CallResAsInt128(ret)
	fmt.Println(r.String())
	if r.String() != "306530824730957902469139748251264" {
		t.Fatal("result is not correct")
	}
}
