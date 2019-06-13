//package main
//
//import (
//	"fmt"
//	"github.com/PlatONEnetwork/PlatONE-Go/common"
//	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
//	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
//	"github.com/PlatONEnetwork/PlatONE-Go/ethdb"
//	"github.com/PlatONEnetwork/PlatONE-Go/life/exec"
//	"github.com/PlatONEnetwork/PlatONE-Go/life/resolver"
//	"github.com/PlatONEnetwork/PlatONE-Go/params"
//	"math/big"
//	"time"
//)
//
////func main() {
////	//entryFunctionFlag := flag.String("entry", "app_main", "entry function id")
////	//dynamicPages := flag.Int("dynamicPages", 1, "dynamic memory pages")
////
////	//jitFlag := flag.Bool("jit", false, "enable jit")
////	//flag.Parse()
////
////	// mocking test
////	flag := false
////	pages := 64
////	functionFlag := "transfer"
////	jitFlag := &flag
////	dynamicPages := &pages
////	entryFunctionFlag := &functionFlag
////
////	rl := resolver.NewResolver(0x01)
////	// Read WebAssembly *.wasm file.
////	//input, err := ioutil.ReadFile(flag.Arg(0))
////	input, err := ioutil.ReadFile("D:\\repos\\PlatONE-contract\\build\\hello\\hello.wasm")
////	//fmt.Println(common.ToHex(input))
////	if err != nil {
////		panic(err)
////	}
////
////	rootLog := log.New()
////	rootLog.SetHandler(log.StderrHandler)
////
////	// Instantiate a new WebAssembly VM with a few resolved imports.
////	vm, err := exec.NewVirtualMachine(input, &exec.VMContext{
////		Config: exec.VMConfig{
////			EnableJIT:          *jitFlag,
////			DefaultMemoryPages: 128,
////			DefaultTableSize:   65536,
////			DynamicMemoryPages: *dynamicPages,
////		},
////		Addr:     [20]byte{},
////		GasUsed:  0,
////		GasLimit: 20000000,
////		Log: rootLog,
////	}, rl, nil)
////
////	if err != nil {
////		panic(err)
////	}
////
////	*entryFunctionFlag = "transfer"
////	// Get the function ID of the entry function to be executed.
////	entryID, ok := vm.GetFunctionExport(*entryFunctionFlag)
////	if !ok {
////		fmt.Printf("Entry function %s not found; starting from 0.\n", *entryFunctionFlag)
////		entryID = 0
////	}
////
////	start := time.Now()
////
////	// If any function prior to the entry function was declared to be
////	// called by the module, run it first.
////	if vm.Module.Base.Start != nil {
////		startID := int(vm.Module.Base.Start.Index)
////		_, err := vm.Run(startID)
////		if err != nil {
////			vm.PrintStackTrace()
////			panic(err)
////		}
////	}
////
////	// Run the WebAssembly module's entry function.
////	ret, err := vm.Run(entryID, resolver.MallocString(vm, "hello"), resolver.MallocString(vm, "world"), 45)
////	if err != nil {
////		vm.PrintStackTrace()
////		panic(err)
////	}
////	end := time.Now()
////
////	fmt.Printf("return value = %d, duration = %v  gasUsed:%d \n", ret, end.Sub(start), vm.Context.GasUsed)
////}
//
//func main() {
//
//
//	lvm := NewMockedLVM()
//
//	locals := []int64{0, 65, 65, 65}
//	lvm.SetMockedFrame(nil, nil, locals, 0, 0, 0)
//
//	memory := make([]byte, 256)
//	m := exec.Memory{
//		Memory: memory,
//	}
//	lvm.SetMockerMemory(&m)
//
//	funcString := []string{"gasPrice", "blockHash", "number", "gasLimit", "timestamp", "coinbase", "balance", "origin", "caller", "isOwner", "isFromInit", "callValue", "address", "sha3", "emitEvent", "setState", "getState", "getStateSize"}
//
//	for i := 0; i < 10; i++ {
//		for _, str := range funcString {
//			start := time.Now()
//			for j := 0; j < 100000; j++ {
//				importFuncs := resolver.Cfc["env"][str]
//				importFuncs.Execute(lvm)
//			}
//			fmt.Println(str, time.Since(start))
//		}
//
//	}
//
//}
//
//func NewMockedLVM() *exec.VirtualMachine {
//
//	// new db: LDB, CacheDB
//	ldb, err := ethdb.NewLDBDatabase("db_for_test", 1, 1)
//	if err != nil {
//		panic("NewLDBDatabase()  failed")
//	}
//
//	cacheDB := state.NewDatabase(ldb)
//
//	// new stateDB
//	statedb, err := state.New(common.Hash{}, cacheDB)
//	if err != nil {
//		panic("state.New(common.Hash{},cacheDB)")
//	}
//
//	// new evm
//	context := vm.Context{
//		GetHash: func(uint64) common.Hash{return common.Hash{}},
//		CanTransfer:func(vm.StateDB, common.Address, *big.Int) bool{return true},
//		Transfer:func(vm.StateDB, common.Address, common.Address, *big.Int){},
//		GasPrice:big.NewInt(321),
//		BlockNumber:big.NewInt(321),
//		Time       :big.NewInt(321),
//		Difficulty :big.NewInt(321),
//	}
//	Evm := vm.NewEVM(context, statedb, params.MainnetChainConfig, vm.Config{})
//
//	// new wasmDB
//	c := vm.AccountRef(common.HexToAddress("1337"))
//	d := vm.AccountRef(common.HexToAddress("1337"))
//
//	contract := vm.NewContract(c, d, big.NewInt(321), 321)
//	wasmStateDB := vm.NewMockedWasmStateDB(statedb, Evm, &vm.Config{}, contract)
//
//	// set memory, locals, etc.
//	lvm := exec.NewMockedVm()
//
//	lvm.Context.StateDB = wasmStateDB
//
//	return lvm
//}
