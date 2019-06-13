package exec

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/log"

	"github.com/PlatONEnetwork/PlatONE-Go/life/compiler"
	"github.com/PlatONEnetwork/PlatONE-Go/life/compiler/opcodes"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"

	"github.com/go-interpreter/wagon/wasm"
)

type (
	Execute func(vm *VirtualMachine) int64
	GasCost func(vm *VirtualMachine) (uint64, error)
)

// FunctionImport represents the function import type. If len(sig.ReturnTypes) == 0, the return value will be ignored.
type FunctionImport struct {
	Execute Execute
	GasCost GasCost
}

const (
	// DefaultCallStackSize is the default call stack size.
	DefaultCallStackSize = 512

	// DefaultPageSize is the linear memory page size.  65536
	DefaultPageSize = 65536

	// JITCodeSizeThreshold is the lower-bound code size threshold for the JIT compiler.
	JITCodeSizeThreshold = 30

	DefaultMemoryPages = 16
	DynamicMemoryPages = 16

	DefaultMemPoolCount   = 5
	DefaultMemBlockSize   = 5
	DefaultMemTreeMaxPage = 8
)

// LE is a simple alias to `binary.LittleEndian`.
var LE = binary.LittleEndian
var memPool = NewMemPool(DefaultMemPoolCount, DefaultMemBlockSize)
var treePool = NewTreePool(DefaultMemPoolCount, DefaultMemBlockSize)

//var pageMemPool = NewPageMemPool()

// VirtualMachine is a WebAssembly execution environment.
type VirtualMachine struct {
	Context         *VMContext
	Module          *compiler.Module
	FunctionCode    []compiler.InterpreterCode
	FunctionImports []*FunctionImport
	JumpTable       [256]Instruction
	CallStack       []Frame
	CurrentFrame    int
	Table           []uint32
	Globals         []int64
	//Memory          *VMMemory
	Memory         *Memory
	NumValueSlots  int
	Yielded        int64
	InsideExecute  bool
	Delegate       func()
	Exited         bool
	ExitError      interface{}
	ReturnValue    int64
	Gas            uint64
	ExternalParams []int64
	InitEntryID    int
}

// VMConfig denotes a set of options passed to a single VirtualMachine insta.ce
type VMConfig struct {
	EnableJIT          bool
	DynamicMemoryPages int
	MaxMemoryPages     int
	MaxTableSize       int
	MaxValueSlots      int
	MaxCallStackDepth  int
	DefaultMemoryPages int
	DefaultTableSize   int
	GasLimit           uint64
	DisableFree        bool
}

type VMContext struct {
	Config   VMConfig
	Addr     [20]byte
	GasUsed  uint64
	GasLimit uint64

	StateDB StateDB
	Log     log.Logger
}

type VMMemory struct {
	Memory    []byte
	Start     int
	Current   int
	MemPoints map[int]int
}

// Frame represents a call frame.
type Frame struct {
	FunctionID   int
	Code         []byte
	JITInfo      interface{}
	Regs         []int64
	Locals       []int64
	IP           int
	ReturnReg    int
	Continuation int32
}

// ImportResolver is an interface for allowing one to define imports to WebAssembly modules
// ran under a single VirtualMachine instance.
type ImportResolver interface {
	ResolveFunc(module, field string) *FunctionImport
	ResolveGlobal(module, field string) int64
}

func ParseModuleAndFunc(code []byte, gasPolicy compiler.GasPolicy) (*compiler.Module, []compiler.InterpreterCode, error) {
	m, err := compiler.LoadModule(code)
	if err != nil {
		return nil, nil, err
	}

	functionCode, err := m.CompileForInterpreter(nil)
	if err != nil {
		return nil, nil, err
	}
	return m, functionCode, nil
}

func NewMockedVm() *VirtualMachine {

	funcImports := make([]*FunctionImport, 0)

	memory := Memory{}


	return &VirtualMachine{
		Context: &VMContext{
			GasLimit: 0xffffffff,
		},
		FunctionImports: funcImports,
		CallStack:       make([]Frame, DefaultCallStackSize),
		CurrentFrame:    -1,
		Memory:          &memory,
		Exited:          false,
		ExternalParams:  make([]int64, 0),
		InitEntryID:     -1,
		Globals: make([]int64,1024),
		JumpTable:       GasTable,
		FunctionCode:    []compiler.InterpreterCode{},
	}
}

// GetCurrentFrame returns the current frame.
func (vm *VirtualMachine) SetMockedFrame(code []byte, regs []int64, locals []int64, ip int, returnReg int, contiueaton int32) *Frame {

	f := Frame{
		FunctionID:   0,
		Code:         code,
		JITInfo:      nil,
		Regs:         regs,
		Locals:       locals,
		IP:           ip,
		ReturnReg:    returnReg,
		Continuation: contiueaton,
	}
	vm.CurrentFrame++

	vm.CallStack[vm.CurrentFrame] = f

	return &f
}

func (vm *VirtualMachine) SetMockerMemory(m *Memory) {
	vm.Memory = m
}

func NewVirtualMachineWithModule(m *compiler.Module, functionCode []compiler.InterpreterCode, context *VMContext, impResolver ImportResolver, gasPolicy compiler.GasPolicy) (_retVM *VirtualMachine, retErr error) {
	defer utils.CatchPanic(&retErr)

	table := make([]uint32, 0)
	globals := make([]int64, 0)
	funcImports := make([]*FunctionImport, 0)

	if m.Base.Import != nil && impResolver != nil {
		for _, imp := range m.Base.Import.Entries {
			switch imp.Type.Kind() {
			case wasm.ExternalFunction:
				funcImports = append(funcImports, impResolver.ResolveFunc(imp.ModuleName, imp.FieldName))
			case wasm.ExternalGlobal:
				globals = append(globals, impResolver.ResolveGlobal(imp.ModuleName, imp.FieldName))
			case wasm.ExternalMemory:
				// TODO: Do we want a real import?
				if m.Base.Memory != nil && len(m.Base.Memory.Entries) > 0 {
					panic("cannot import another memory while we already have one")
				}
				m.Base.Memory = &wasm.SectionMemories{
					Entries: []wasm.Memory{
						wasm.Memory{
							Limits: wasm.ResizableLimits{
								Initial: uint32(context.Config.DefaultMemoryPages),
							},
						},
					},
				}
			case wasm.ExternalTable:
				// TODO: Do we want a real import?
				if m.Base.Table != nil && len(m.Base.Table.Entries) > 0 {
					panic("cannot import another table while we already have one")
				}
				m.Base.Table = &wasm.SectionTables{
					Entries: []wasm.Table{
						wasm.Table{
							Limits: wasm.ResizableLimits{
								Initial: uint32(context.Config.DefaultTableSize),
							},
						},
					},
				}
			default:
				panic(fmt.Errorf("import kind not supported: %d", imp.Type.Kind()))
			}
		}
	}

	// Load global entries.
	for _, entry := range m.Base.GlobalIndexSpace {
		globals = append(globals, execInitExpr(entry.Init, globals))
	}

	// Populate table elements.
	if m.Base.Table != nil && len(m.Base.Table.Entries) > 0 {
		t := &m.Base.Table.Entries[0]

		if context.Config.MaxTableSize != 0 && int(t.Limits.Initial) > context.Config.MaxTableSize {
			panic("max table size exceeded")
		}

		table = make([]uint32, int(t.Limits.Initial))
		for i := 0; i < int(t.Limits.Initial); i++ {
			table[i] = 0xffffffff
		}
		if m.Base.Elements != nil && len(m.Base.Elements.Entries) > 0 {
			for _, e := range m.Base.Elements.Entries {
				offset := int(execInitExpr(e.Offset, globals))
				copy(table[offset:], e.Elems)
			}
		}
	}

	// Load linear memory.
	//memory := make([]byte, 0)

	//memory := &VMMemory{
	//	Memory:    make([]byte, 0),
	//	Start:     0,
	//	Current:   0,
	//	MemPoints: make(map[int]int),
	//}
	//
	//if m.Base.Memory != nil && len(m.Base.Memory.Entries) > 0 {
	//	initialLimit := int(m.Base.Memory.Entries[0].Limits.Initial)
	//	if context.Config.MaxMemoryPages != 0 && initialLimit > context.Config.MaxMemoryPages {
	//		panic("max memory exceeded")
	//	}
	//
	//	capacity := initialLimit + context.Config.DynamicMemoryPages
	//
	//	memory.Start = initialLimit * DefaultPageSize
	//	memory.Current = initialLimit * DefaultPageSize
	//	// Initialize empty memory.
	//	//buffer := bytes.NewBuffer(make([]byte, capacity))
	//	memory.Memory = memPool.Get(capacity)
	//
	//	if m.Base.Data != nil && len(m.Base.Data.Entries) > 0 {
	//		for _, e := range m.Base.Data.Entries {
	//			offset := int(execInitExpr(e.Offset, globals))
	//			copy(memory.Memory[int(offset):], e.Data)
	//		}
	//	}
	//}

	memory := &Memory{}
	if m.Base.Memory != nil && len(m.Base.Memory.Entries) > 0 {
		initialLimit := int(m.Base.Memory.Entries[0].Limits.Initial)
		if context.Config.MaxMemoryPages != 0 && initialLimit > context.Config.MaxMemoryPages {
			panic("max memory exceeded")
		}

		capacity := initialLimit + context.Config.DynamicMemoryPages
		// Initialize empty memory.
		memory.Memory = memPool.Get(capacity)
		memory.Start = initialLimit * DefaultPageSize
		memory.tree = treePool.GetTree(capacity - initialLimit)
		memory.Size = (len(memory.tree) + 1) / 2

		if m.Base.Data != nil && len(m.Base.Data.Entries) > 0 {
			for _, e := range m.Base.Data.Entries {
				offset := int(execInitExpr(e.Offset, globals))
				copy(memory.Memory[int(offset):], e.Data)
			}
		}
	}

	return &VirtualMachine{
		Module:          m,
		Context:         context,
		FunctionCode:    functionCode,
		FunctionImports: funcImports,
		JumpTable:       GasTable,
		CallStack:       make([]Frame, DefaultCallStackSize),
		CurrentFrame:    -1,
		Table:           table,
		Globals:         globals,
		Memory:          memory,
		Exited:          true,
		ExternalParams:  make([]int64, 0),
		InitEntryID:     -1,
	}, nil
}

// NewVirtualMachine instantiates a virtual machine for a given WebAssembly module, with
// specific execution options specified under a VMConfig, and a WebAssembly module import
// resolver.
func NewVirtualMachine(code []byte, context *VMContext, impResolver ImportResolver, gasPolicy compiler.GasPolicy) (_retVM *VirtualMachine, retErr error) {
	if context.Config.EnableJIT {
		fmt.Println("Warning: JIT support is incomplete and the internals are likely to change in the future.")
	}

	m, functionCode, err := ParseModuleAndFunc(code, gasPolicy)
	if err != nil {
		return nil, err
	}

	return NewVirtualMachineWithModule(m, functionCode, context, impResolver, gasPolicy)

}

func ImportGasFunc(vm *VirtualMachine, frame *Frame) (uint64, error) {
	importID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
	return vm.FunctionImports[importID].GasCost(vm)
}

// Init initializes a frame. Must be called on `call` and `call_indirect`.
func (f *Frame) Init(vm *VirtualMachine, functionID int, code compiler.InterpreterCode) {
	numValueSlots := code.NumRegs + code.NumParams + code.NumLocals
	if vm.Context.Config.MaxValueSlots != 0 && vm.NumValueSlots+numValueSlots > vm.Context.Config.MaxValueSlots {
		panic("max value slot count exceeded")
	}
	vm.NumValueSlots += numValueSlots

	values := make([]int64, numValueSlots)

	f.FunctionID = functionID
	f.Regs = values[:code.NumRegs]
	f.Locals = values[code.NumRegs:]
	f.Code = code.Bytes
	f.IP = 0
	f.Continuation = 0

	//fmt.Printf("Enter function %d (%s)\n", functionID, vm.Module.FunctionNames[functionID])
	if vm.Context.Config.EnableJIT {
		code := &vm.FunctionCode[functionID]
		if !code.JITDone {
			if len(code.Bytes) > JITCodeSizeThreshold {
				if !vm.GenerateCodeForFunction(functionID) {
					fmt.Printf("codegen for function %d failed\n", functionID)
				} else {
					fmt.Printf("codegen for function %d succeeded\n", functionID)
				}
			}
			code.JITDone = true
		}
		f.JITInfo = code.JITInfo
	}
}

// Destroy destroys a frame. Must be called on return.
func (f *Frame) Destroy(vm *VirtualMachine) {
	//numValueSlots := len(f.Regs) + len(f.Locals)
	//vm.NumValueSlots -= numValueSlots
	_ = len(f.Regs) + len(f.Locals)
	vm.NumValueSlots -= 0

	//fmt.Printf("Leave function %d (%s)\n", f.FunctionID, vm.Module.FunctionNames[f.FunctionID])
}

// GetCurrentFrame returns the current frame.
func (vm *VirtualMachine) GetCurrentFrame() *Frame {
	if vm.Context.Config.MaxCallStackDepth != 0 && vm.CurrentFrame >= vm.Context.Config.MaxCallStackDepth {
		panic("max call stack depth exceeded")
	}

	if vm.CurrentFrame >= len(vm.CallStack) {
		panic("call stack overflow")
		//vm.CallStack = append(vm.CallStack, make([]Frame, DefaultCallStackSize / 2)...)
	}
	return &vm.CallStack[vm.CurrentFrame]
}

func (vm *VirtualMachine) getExport(key string, kind wasm.External) (int, bool) {
	if vm.Module.Base.Export == nil {
		return -1, false
	}

	entry, ok := vm.Module.Base.Export.Entries[key]
	if !ok {
		return -1, false
	}

	if entry.Kind != kind {
		return -1, false
	}

	return int(entry.Index), true
}

// GetGlobalExport returns the global export with the given name.
func (vm *VirtualMachine) GetGlobalExport(key string) (int, bool) {
	return vm.getExport(key, wasm.ExternalGlobal)
}

// GetFunctionExport returns the function export with the given name.
func (vm *VirtualMachine) GetFunctionExport(key string) (int, bool) {
	return vm.getExport(key, wasm.ExternalFunction)
}

// PrintStackTrace prints the entire VM stack trace for debugging.
func (vm *VirtualMachine) PrintStackTrace() {
	fmt.Println("--- Begin stack trace ---")
	for i := vm.CurrentFrame; i >= 0; i-- {
		functionID := vm.CallStack[i].FunctionID
		fmt.Printf("<%d> [%d] %s\n", i, functionID, vm.Module.FunctionNames[functionID])
	}
	fmt.Println("--- End stack trace ---")
}

// Ignite initializes the first call frame.
func (vm *VirtualMachine) Ignite(functionID int, params ...int64) {
	if vm.ExitError != nil {
		panic("last execution exited with error; cannot ignite.")
	}

	if vm.CurrentFrame != -1 {
		panic("call stack not empty; cannot ignite.")
	}

	code := vm.FunctionCode[functionID]
	if code.NumParams != len(params) {
		panic("param count mismatch")
	}

	vm.Exited = false

	vm.CurrentFrame++
	frame := vm.GetCurrentFrame()
	frame.Init(
		vm,
		functionID,
		code,
	)
	copy(frame.Locals, params)
}

func (vm *VirtualMachine) AddAndCheckGas(delta uint64) {
	newGas := vm.Gas + delta
	if newGas < vm.Gas {
		panic("gas overflow")
	}
	if vm.Context.Config.GasLimit != 0 && newGas > vm.Context.Config.GasLimit {
		panic("gas limit exceeded")
	}
	vm.Gas = newGas
}

// Execute starts the virtual machines main instruction processing loop.
// This function may return at any point and is guaranteed to return
// at least once every 10000 instructions. Caller is responsible for
// detecting VM status in a loop.
func (vm *VirtualMachine) Execute() {

	if vm.Exited == true {
		panic("attempting to execute an exited vm")
	}

	if vm.Delegate != nil {
		panic("delegate not cleared")
	}

	if vm.InsideExecute {
		panic("vm execution is not re-entrant")
	}
	vm.InsideExecute = true

	defer func() {
		vm.InsideExecute = false
		if err := recover(); err != nil {
			vm.Exited = true
			vm.ExitError = err
		}
	}()

	frame := vm.GetCurrentFrame()

	for {
		if frame.JITInfo != nil {
			dm := frame.JITInfo.(*DynamicModule)
			var fRetVal int64
			status := dm.Run(vm, &fRetVal)
			if status < 0 {
				panic(fmt.Errorf("status = %d", status))
			}
			//fmt.Printf("JIT: continuation = %d, ip = %d\n", status, int(fRetVal))
			frame.Continuation = status
			frame.IP = int(fRetVal)
		}

		valueID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
		ins := opcodes.Opcode(frame.Code[frame.IP+4])
		frame.IP += 5

		//cost, err := vm.JumpTable[ins].GasCost(vm, frame)
		//if err != nil || (cost+vm.Context.GasUsed) > vm.Context.GasLimit {
		//	//panic(fmt.Sprintf("out of gas  cost:%d GasUsed:%d GasLimit:%d", cost, vm.Context.GasUsed, vm.Context.GasLimit))
		//}
		//vm.Context.GasUsed += cost

		//fmt.Printf("INS: [%d] %s\n", valueID, ins.String())

		switch ins {
		case opcodes.Nop:
		case opcodes.Unreachable:
			panic("wasm: unreachable executed")
		case opcodes.Select:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				c := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])
				//frame.IP += 12
				if c != 0 {
					frame.Regs[valueID] = a
				} else {
					frame.Regs[valueID] = b
				}
			}
			fmt.Println("Select", time.Since(start))
		case opcodes.I32Const:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				//frame.IP += 4
				frame.Regs[valueID] = int64(val)
			}
			fmt.Println("I32Const", time.Since(start))
		case opcodes.I32Add:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				frame.Regs[valueID] = int64(a + b)
			}
			fmt.Println("I32Add", time.Since(start))

		case opcodes.I32Sub:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				frame.Regs[valueID] = int64(a - b)
			}
			fmt.Println("I32Sub", time.Since(start))
		case opcodes.I32Mul:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				frame.Regs[valueID] = int64(a * b)
			}
			fmt.Println("I32Mul", time.Since(start))
		case opcodes.I32DivS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				if b == 0 {
					panic("integer division by zero")
				}

				if a == math.MinInt32 && b == -1 {
					panic("signed integer overflow")
				}

				//frame.IP += 8
				frame.Regs[valueID] = int64(a / b)
			}
			fmt.Println("I32DivS", time.Since(start))
		case opcodes.I32DivU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				if b == 0 {
					panic("integer division by zero")
				}

				//frame.IP += 8
				frame.Regs[valueID] = int64(a / b)
			}
			fmt.Println("I32DivU", time.Since(start))
		case opcodes.I32RemS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				if b == 0 {
					panic("integer division by zero")
				}

				//frame.IP += 8
				frame.Regs[valueID] = int64(a % b)
			}
			fmt.Println("I32RemS", time.Since(start))
		case opcodes.I32RemU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				if b == 0 {
					panic("integer division by zero")
				}

				//frame.IP += 8
				frame.Regs[valueID] = int64(a % b)
			}
			fmt.Println("I32RemU", time.Since(start))
		case opcodes.I32And:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(a & b)
			}
			fmt.Println("I32And", time.Since(start))
		case opcodes.I32Or:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(a | b)
			}
			fmt.Println("I32Or", time.Since(start))
		case opcodes.I32Xor:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(a ^ b)
			}
			fmt.Println("I32Xor", time.Since(start))
		case opcodes.I32Shl:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(a << (b % 32))
			}
			fmt.Println("I32Shl", time.Since(start))
		case opcodes.I32ShrS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(a >> (b % 32))
			}
			fmt.Println("I32ShrS", time.Since(start))
		case opcodes.I32ShrU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(a >> (b % 32))
			}
			fmt.Println("I32ShrU", time.Since(start))
		case opcodes.I32Rotl:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(bits.RotateLeft32(a, int(b)))
			}
			fmt.Println("I32Rotl", time.Since(start))
		case opcodes.I32Rotr:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(bits.RotateLeft32(a, -int(b)))
			}
			fmt.Println("I32Rotr", time.Since(start))
		case opcodes.I32Clz:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

				//frame.IP += 4
				frame.Regs[valueID] = int64(bits.LeadingZeros32(val))
			}
			fmt.Println("I32Clz", time.Since(start))
		case opcodes.I32Ctz:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

				//frame.IP += 4
				frame.Regs[valueID] = int64(bits.TrailingZeros32(val))
			}
			fmt.Println("I32Ctz", time.Since(start))
		case opcodes.I32PopCnt:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

				//frame.IP += 4
				frame.Regs[valueID] = int64(bits.OnesCount32(val))
			}
			fmt.Println("I32PopCnt", time.Since(start))
		case opcodes.I32EqZ:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

				//frame.IP += 4
				if val == 0 {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32EqZ", time.Since(start))
		case opcodes.I32Eq:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a == b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32Eq", time.Since(start))
		case opcodes.I32Ne:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a != b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32Ne", time.Since(start))
		case opcodes.I32LtS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a < b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32LtS", time.Since(start))
		case opcodes.I32LtU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a < b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32LtU", time.Since(start))
		case opcodes.I32LeS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a <= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32LeS", time.Since(start))
		case opcodes.I32LeU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a <= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32LeU", time.Since(start))
		case opcodes.I32GtS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a > b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32GtS", time.Since(start))
		case opcodes.I32GtU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a > b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32GtU", time.Since(start))
		case opcodes.I32GeS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a >= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32GeS", time.Since(start))
		case opcodes.I32GeU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a >= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I32GeU", time.Since(start))
		case opcodes.I64Const:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := LE.Uint64(frame.Code[frame.IP : frame.IP+8])
				//frame.IP += 8
				frame.Regs[valueID] = int64(val)
			}
			fmt.Println("I64Const", time.Since(start))
		case opcodes.I64Add:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				frame.Regs[valueID] = a + b
			}
			fmt.Println("I64Add", time.Since(start))
		case opcodes.I64Sub:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				frame.Regs[valueID] = a - b
			}
			fmt.Println("I64Sub", time.Since(start))
		case opcodes.I64Mul:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				frame.Regs[valueID] = a * b
			}
			fmt.Println("I64Mul", time.Since(start))
		case opcodes.I64DivS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

				if b == 0 {
					panic("integer division by zero")
				}

				if a == math.MinInt64 && b == -1 {
					panic("signed integer overflow")
				}

				//frame.IP += 8
				frame.Regs[valueID] = a / b
			}
			fmt.Println("I64DivS", time.Since(start))
		case opcodes.I64DivU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				if b == 0 {
					panic("integer division by zero")
				}

				//frame.IP += 8
				frame.Regs[valueID] = int64(a / b)
			}
			fmt.Println("I64DivU", time.Since(start))
		case opcodes.I64RemS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

				if b == 0 {
					panic("integer division by zero")
				}

				//frame.IP += 8
				frame.Regs[valueID] = a % b
			}
			fmt.Println("I64RemS", time.Since(start))
		case opcodes.I64RemU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				if b == 0 {
					panic("integer division by zero")
				}

				//frame.IP += 8
				frame.Regs[valueID] = int64(a % b)
			}
			fmt.Println("I64RemU", time.Since(start))
		case opcodes.I64And:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

				//frame.IP += 8
				frame.Regs[valueID] = a & b
			}
			fmt.Println("I64And", time.Since(start))
		case opcodes.I64Or:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

				//frame.IP += 8
				frame.Regs[valueID] = a | b
			}
			fmt.Println("I64Or", time.Since(start))
		case opcodes.I64Xor:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

				//frame.IP += 8
				frame.Regs[valueID] = a ^ b
			}
			fmt.Println("I64Xor", time.Since(start))
		case opcodes.I64Shl:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = a << (b % 64)
			}
			fmt.Println("I64Shl", time.Since(start))
		case opcodes.I64ShrS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = a >> (b % 64)
			}
			fmt.Println("I64ShrS", time.Since(start))
		case opcodes.I64ShrU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(a >> (b % 64))
			}
			fmt.Println("I64ShrU", time.Since(start))
		case opcodes.I64Rotl:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(bits.RotateLeft64(a, int(b)))
			}
			fmt.Println("I64Rotl", time.Since(start))
		case opcodes.I64Rotr:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

				//frame.IP += 8
				frame.Regs[valueID] = int64(bits.RotateLeft64(a, -int(b)))
			}
			fmt.Println("I64Rotr", time.Since(start))
		case opcodes.I64Clz:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

				//frame.IP += 4
				frame.Regs[valueID] = int64(bits.LeadingZeros64(val))
			}
			fmt.Println("I64Clz", time.Since(start))
		case opcodes.I64Ctz:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

				//frame.IP += 4
				frame.Regs[valueID] = int64(bits.TrailingZeros64(val))
			}
			fmt.Println("I64Ctz", time.Since(start))
		case opcodes.I64PopCnt:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

				//frame.IP += 4
				frame.Regs[valueID] = int64(bits.OnesCount64(val))
			}
			fmt.Println("I64PopCnt", time.Since(start))
		case opcodes.I64EqZ:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

				//frame.IP += 4
				if val == 0 {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64EqZ", time.Since(start))
		case opcodes.I64Eq:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				if a == b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64Eq", time.Since(start))
		case opcodes.I64Ne:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				if a != b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64Ne", time.Since(start))
		case opcodes.I64LtS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				if a < b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64LtS", time.Since(start))
		case opcodes.I64LtU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a < b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64LtU", time.Since(start))
		case opcodes.I64LeS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				if a <= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64LeS", time.Since(start))
		case opcodes.I64LeU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a <= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64LeU", time.Since(start))
		case opcodes.I64GtS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				if a > b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64GtS", time.Since(start))
		case opcodes.I64GtU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a > b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64GtU", time.Since(start))
		case opcodes.I64GeS:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				if a >= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64GeS", time.Since(start))
		case opcodes.I64GeU:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
				//frame.IP += 8
				if a >= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("I64GeU", time.Since(start))
		case opcodes.F32Add:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float32bits(a + b))
			}
			fmt.Println("F32Add", time.Since(start))
		case opcodes.F32Sub:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float32bits(a - b))
			}
			fmt.Println("F32Sub", time.Since(start))
		case opcodes.F32Mul:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float32bits(a * b))
			}
			fmt.Println("F32Mul", time.Since(start))
		case opcodes.F32Div:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float32bits(a / b))
			}
			fmt.Println("F32Div", time.Since(start))
		case opcodes.F32Sqrt:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.Sqrt(float64(val)))))
			}
			fmt.Println("F32Sqrt", time.Since(start))
		case opcodes.F32Min:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.Min(float64(a), float64(b)))))
			}
			fmt.Println("F32Min", time.Since(start))
		case opcodes.F32Max:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.Max(float64(a), float64(b)))))
			}
			fmt.Println("F32Max", time.Since(start))
		case opcodes.F32Ceil:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.Ceil(float64(val)))))
			}
			fmt.Println("F32Ceil", time.Since(start))
		case opcodes.F32Floor:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.Floor(float64(val)))))
			}
			fmt.Println("F32Floor", time.Since(start))
		case opcodes.F32Trunc:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.Trunc(float64(val)))))
			}
			fmt.Println("F32Trunc", time.Since(start))
		case opcodes.F32Nearest:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.RoundToEven(float64(val)))))
			}
			fmt.Println("F32Nearest", time.Since(start))
		case opcodes.F32Abs:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.Abs(float64(val)))))
			}
			fmt.Println("F32Abs", time.Since(start))
		case opcodes.F32Neg:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(-val))
			}
			fmt.Println("F32Neg", time.Since(start))
		case opcodes.F32CopySign:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float32bits(float32(math.Copysign(float64(a), float64(b)))))
			}
			fmt.Println("F32CopySign", time.Since(start))
		case opcodes.F32Eq:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a == b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F32Eq", time.Since(start))
		case opcodes.F32Ne:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a != b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F32Ne", time.Since(start))
		case opcodes.F32Lt:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a < b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F32Lt", time.Since(start))
		case opcodes.F32Le:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a <= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F32Le", time.Since(start))
		case opcodes.F32Gt:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a > b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F32Gt", time.Since(start))
		case opcodes.F32Ge:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a >= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F32Ge", time.Since(start))
		case opcodes.F64Add:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float64bits(a + b))
			}
			fmt.Println("F64Add ", time.Since(start))
		case opcodes.F64Sub:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float64bits(a - b))
			}
			fmt.Println("F64Sub ", time.Since(start))
		case opcodes.F64Mul:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float64bits(a * b))
			}
			fmt.Println("F64Mul ", time.Since(start))
		case opcodes.F64Div:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float64bits(a / b))
			}
			fmt.Println("F64Div ", time.Since(start))
		case opcodes.F64Sqrt:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//	frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(math.Sqrt(val)))
			}
			fmt.Println("F64Sqrt ", time.Since(start))
		case opcodes.F64Min:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float64bits(math.Min(a, b)))
			}
			fmt.Println("F64Min ", time.Since(start))
		case opcodes.F64Max:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float64bits(math.Max(a, b)))
			}
			fmt.Println("F64Max ", time.Since(start))
		case opcodes.F64Ceil:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(math.Ceil(val)))
			}
			fmt.Println("F64Ceil ", time.Since(start))
		case opcodes.F64Floor:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(math.Floor(val)))
			}
			fmt.Println("F64Floor ", time.Since(start))
		case opcodes.F64Trunc:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(math.Trunc(val)))
			}
			fmt.Println("F64Trunc ", time.Since(start))
		case opcodes.F64Nearest:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(math.RoundToEven(val)))
			}
			fmt.Println("F64Nearest ", time.Since(start))
		case opcodes.F64Abs:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(math.Abs(val)))
			}
			fmt.Println("F64Abs ", time.Since(start))
		case opcodes.F64Neg:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//	frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(-val))
			}
			fmt.Println("F64Neg ", time.Since(start))
		case opcodes.F64CopySign:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				frame.Regs[valueID] = int64(math.Float64bits(math.Copysign(a, b)))
			}
			fmt.Println("F64CopySign ", time.Since(start))
		case opcodes.F64Eq:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a == b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F64Eq ", time.Since(start))
		case opcodes.F64Ne:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a != b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F64Ne ", time.Since(start))
		case opcodes.F64Lt:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a < b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F64Lt ", time.Since(start))
		case opcodes.F64Le:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a <= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F64Le ", time.Since(start))
		case opcodes.F64Gt:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a > b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F64Gt ", time.Since(start))
		case opcodes.F64Ge:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
				//frame.IP += 8
				if a >= b {
					frame.Regs[valueID] = 1
				} else {
					frame.Regs[valueID] = 0
				}
			}
			fmt.Println("F64Ge ", time.Since(start))
		case opcodes.I32WrapI64:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//frame.IP += 4
				frame.Regs[valueID] = int64(v)
			}
			fmt.Println("I32WrapI64 ", time.Since(start))
		case opcodes.I32TruncSF32, opcodes.I32TruncUF32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(int32(math.Trunc(float64(v))))
			}
			fmt.Println("I32TruncSF32 I32TruncUF32 ", time.Since(start))
		case opcodes.I32TruncSF64, opcodes.I32TruncUF64:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(int32(math.Trunc(v)))
			}
			fmt.Println("I32TruncSF64 I32TruncUF64 ", time.Since(start))
		case opcodes.I64TruncSF32, opcodes.I64TruncUF32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Trunc(float64(v)))
			}
			fmt.Println("I64TruncSF32 I64TruncUF32 ", time.Since(start))
		case opcodes.I64TruncSF64, opcodes.I64TruncUF64:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Trunc(v))
			}
			fmt.Println("I64TruncSF64 I64TruncUF64 ", time.Since(start))
		case opcodes.F32DemoteF64:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(v)))
			}
			fmt.Println("F32DemoteF64", time.Since(start))
		case opcodes.F64PromoteF32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(float64(v)))
			}
			fmt.Println("F64PromoteF32", time.Since(start))
		case opcodes.F32ConvertSI32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//	frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(v)))
			}
			fmt.Println("F32ConvertSI32", time.Since(start))
		case opcodes.F32ConvertUI32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(v)))
			}
			fmt.Println("F32ConvertUI32", time.Since(start))
		case opcodes.F32ConvertSI64:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := int64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(v)))
			}
			fmt.Println("F32ConvertSI64", time.Since(start))
		case opcodes.F32ConvertUI64:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//	frame.IP += 4
				frame.Regs[valueID] = int64(math.Float32bits(float32(v)))
			}
			fmt.Println("F32ConvertUI64", time.Since(start))
		case opcodes.F64ConvertSI32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//frame.IP += 4
				frame.Regs[valueID] = int64(int32(math.Float64bits(float64(v))))
			}
			fmt.Println("F64ConvertSI32", time.Since(start))
		case opcodes.F64ConvertUI32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//frame.IP += 4
				frame.Regs[valueID] = int64(int32(math.Float64bits(float64(v))))
			}
			fmt.Println("F64ConvertUI32", time.Since(start))
		case opcodes.F64ConvertSI64:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := int64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(float64(v)))
			}
			fmt.Println("F64ConvertSI64", time.Since(start))
		case opcodes.F64ConvertUI64:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//frame.IP += 4
				frame.Regs[valueID] = int64(math.Float64bits(float64(v)))
			}
			fmt.Println("F64ConvertUI64", time.Since(start))
		case opcodes.I64ExtendUI32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
				//frame.IP += 4
				frame.Regs[valueID] = int64(v)
			}
			fmt.Println("I64ExtendUI32", time.Since(start))
		case opcodes.I64ExtendSI32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				v := int32(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				frame.Regs[valueID] = int64(v)
			}
			fmt.Println("I64ExtendSI32", time.Since(start))
		case opcodes.I32Load, opcodes.I64Load32U:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				//frame.IP += 12

				effective := int(uint64(base) + uint64(offset))
				frame.Regs[valueID] = int64(uint32(LE.Uint32(vm.Memory.Memory[effective : effective+4])))
			}
			fmt.Println("I32Load I64Load32U", time.Since(start))
		case opcodes.I64Load32S:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				//frame.IP += 12

				effective := int(uint64(base) + uint64(offset))
				frame.Regs[valueID] = int64(int32(LE.Uint32(vm.Memory.Memory[effective : effective+4])))
			}
			fmt.Println("I64Load32S", time.Since(start))
		case opcodes.I64Load:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				//frame.IP += 12

				effective := int(uint64(base) + uint64(offset))
				frame.Regs[valueID] = int64(LE.Uint64(vm.Memory.Memory[effective : effective+8]))
			}
			fmt.Println("I64Load", time.Since(start))
		case opcodes.I32Load8S, opcodes.I64Load8S:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				//frame.IP += 12

				effective := int(uint64(base) + uint64(offset))
				frame.Regs[valueID] = int64(int8(vm.Memory.Memory[effective]))
			}
			fmt.Println("I32Load8S I64Load8S", time.Since(start))
		case opcodes.I32Load8U, opcodes.I64Load8U:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				//frame.IP += 12

				effective := int(uint64(base) + uint64(offset))
				frame.Regs[valueID] = int64(uint8(vm.Memory.Memory[effective]))
			}
			fmt.Println("I32Load8U I64Load8U", time.Since(start))
		case opcodes.I32Load16S, opcodes.I64Load16S:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				//frame.IP += 12

				effective := int(uint64(base) + uint64(offset))
				frame.Regs[valueID] = int64(int16(LE.Uint16(vm.Memory.Memory[effective : effective+2])))
			}
			fmt.Println("I32Load16S I64Load16S", time.Since(start))
		case opcodes.I32Load16U, opcodes.I64Load16U:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				//frame.IP += 12

				effective := int(uint64(base) + uint64(offset))
				frame.Regs[valueID] = int64(uint16(LE.Uint16(vm.Memory.Memory[effective : effective+2])))
			}
			fmt.Println("I32Load16U I64Load16U", time.Since(start))
		case opcodes.I32Store, opcodes.I64Store32:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				value := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+12:frame.IP+16]))]

				effective := int(uint64(base) + uint64(offset))
				LE.PutUint32(vm.Memory.Memory[effective:effective+4], uint32(value))

			}
			frame.IP += 16

			fmt.Println("I32Store I64Store32", time.Since(start))
		case opcodes.I64Store:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				value := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+12:frame.IP+16]))]

				effective := int(uint64(base) + uint64(offset))
				LE.PutUint64(vm.Memory.Memory[effective:effective+8], uint64(value))
			}
			frame.IP += 16

			fmt.Println("I64Store", time.Since(start))
		case opcodes.I32Store8, opcodes.I64Store8:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				value := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+12:frame.IP+16]))]

				effective := int(uint64(base) + uint64(offset))
				vm.Memory.Memory[effective] = byte(value)
			}
			frame.IP += 16
			fmt.Println("I32Store8 I64Store8", time.Since(start))
		case opcodes.I32Store16, opcodes.I64Store16:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				LE.Uint32(frame.Code[frame.IP : frame.IP+4])
				offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
				base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

				value := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+12:frame.IP+16]))]

				effective := int(uint64(base) + uint64(offset))
				LE.PutUint16(vm.Memory.Memory[effective:effective+2], uint16(value))
			}
			frame.IP += 16

			fmt.Println("I32Store16 I64Store16", time.Since(start))

		case opcodes.Jmp:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				target := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				vm.Yielded = frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP = target
				_ = target
			}
			fmt.Println("Jmp", time.Since(start))
		case opcodes.JmpEither:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				targetA := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				targetB := int(LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8]))
				cond := int(LE.Uint32(frame.Code[frame.IP+8 : frame.IP+12]))
				yieldedReg := int(LE.Uint32(frame.Code[frame.IP+12 : frame.IP+16]))
				//frame.IP += 16
				vm.Yielded = frame.Regs[yieldedReg]

				if frame.Regs[cond] != 0 {
					//frame.IP = targetA
					_ = targetA
				} else {
					//frame.IP = targetB
					_ = targetB
				}
			}
			fmt.Println("JmpEither", time.Since(start))
		case opcodes.JmpIf:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				target := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				cond := int(LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8]))
				yieldedReg := int(LE.Uint32(frame.Code[frame.IP+8 : frame.IP+12]))
				//frame.IP += 12
				if frame.Regs[cond] != 0 {
					vm.Yielded = frame.Regs[yieldedReg]
					//frame.IP = target
					_ = target
				}
			}
			fmt.Println("JmpIf", time.Since(start))
		case opcodes.JmpTable:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				targetCount := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				//frame.IP += 4

				targetsRaw := frame.Code[frame.IP : frame.IP+4*targetCount]
				//frame.IP += 4 * targetCount

				defaultTarget := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				//	frame.IP += 4

				cond := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				//frame.IP += 4

				vm.Yielded = frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				//frame.IP += 4

				val := int(frame.Regs[cond])
				if val >= 0 && val < targetCount {
					//frame.IP = int(LE.Uint32(targetsRaw[val*4 : val*4+4]))
					_ = targetsRaw
				} else {
					//frame.IP = defaultTarget
					_ = defaultTarget
				}

			}
			fmt.Println("JmpTable", time.Since(start))
		case opcodes.ReturnValue:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				val := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				//frame.Destroy(vm)
				//vm.CurrentFrame--
				if vm.CurrentFrame == -1 {
					vm.Exited = true
					vm.ReturnValue = val
					return
				} else {
					frame = vm.GetCurrentFrame()
					frame.Regs[frame.ReturnReg] = val
					//fmt.Printf("Return value %d\n", val)
				}
			}
			fmt.Println("ReturnValue", time.Since(start))
		case opcodes.ReturnVoid:
			start := time.Now()
			num := 0
			for i := 0; i < 100000000; i++ {
				frame.Destroy(vm)
				num --
				//vm.CurrentFrame--

				if vm.CurrentFrame == -1 {
					vm.Exited = true
					vm.ReturnValue = 0
					//return
				} else {
					frame = vm.GetCurrentFrame()
				}
			}
			fmt.Println("ReturnVoid", time.Since(start))
		case opcodes.GetLocal:
			start := time.Now()
			for i := 0; i < 100000000; i++ {

				id := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				val := frame.Locals[id]
				//frame.IP += 4
				frame.Regs[valueID] = val
				//fmt.Printf("GetLocal %d = %d\n", id, val)
			}
			fmt.Println("GetLocal", time.Since(start))
		case opcodes.SetLocal:
			start := time.Now()
			for i := 0; i < 100000000; i++ {

				id := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				val := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8
				frame.Locals[id] = val
				//fmt.Printf("SetLocal %d = %d\n", id, val)
			}
			fmt.Println("SetLocal", time.Since(start))
		case opcodes.GetGlobal:
			start := time.Now()
			for i := 0; i < 100000000; i++ {

				frame.Regs[valueID] = vm.Globals[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				//frame.IP += 4
			}
			fmt.Println("GetGlobal", time.Since(start))
		case opcodes.SetGlobal:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				id := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				val := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
				//frame.IP += 8

				vm.Globals[id] = val
			}
			fmt.Println("SetGlobal", time.Since(start))
		case opcodes.Call:
			start := time.Now()
			//for i := 0; i < 100000000; i++ {
			functionID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			//frame.IP += 4
			argCount := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			//frame.IP += 4
			argsRaw := frame.Code[frame.IP : frame.IP+4*argCount]
			//frame.IP += 4 * argCount

			oldRegs := frame.Regs
			frame.ReturnReg = valueID
			vm.CurrentFrame++
			frame = vm.GetCurrentFrame()

			frame.Init(vm, functionID, vm.FunctionCode[functionID])
			//fmt.Println("hello")
			_ = functionID
			for i := 0; i < argCount; i++ {
				frame.Locals[i] = oldRegs[int(LE.Uint32(argsRaw[i*4:i*4+4]))]
			}
			//}
			fmt.Println("Call", time.Since(start))
			//fmt.Println("Call params =", frame.Locals[:argCount])

		case opcodes.CallIndirect:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				typeID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				//frame.IP += 4
				argCount := int(LE.Uint32(frame.Code[frame.IP:frame.IP+4])) - 1
				//frame.IP += 4
				argsRaw := frame.Code[frame.IP : frame.IP+4*argCount]
				//frame.IP += 4 * argCount
				tableItemID := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
				//frame.IP += 4

				sig := &vm.Module.Base.Types.Entries[typeID]

				functionID := int(vm.Table[tableItemID])
				code := vm.FunctionCode[functionID]

				// TODO: We are only checking CC here; Do we want strict typeck?
				if code.NumParams != len(sig.ParamTypes) || code.NumReturns != len(sig.ReturnTypes) {
					panic("type mismatch")
				}

				oldRegs := frame.Regs
				frame.ReturnReg = valueID

				vm.CurrentFrame++
				frame = vm.GetCurrentFrame()
				frame.Init(vm, functionID, code)
				for i := 0; i < argCount; i++ {
					frame.Locals[i] = oldRegs[int(LE.Uint32(argsRaw[i*4:i*4+4]))]
				}
			}
			fmt.Println("CallIndirect", time.Since(start))

		case opcodes.InvokeImport:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				importID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
				//frame.IP += 4
				_ = importID
				vm.Delegate = func() {
					//frame.Regs[valueID] = vm.FunctionImports[importID].Execute(vm)
				}
				//return
			}
			fmt.Println("InvokeImport", time.Since(start))

		case opcodes.CurrentMemory:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				frame.Regs[valueID] = int64(len(vm.Memory.Memory) / DefaultPageSize)
			}
			fmt.Println("CurrentMemory", time.Since(start))

		case opcodes.GrowMemory:
			start := time.Now()
			for i := 0; i < 10000; i++ {
				n := int(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
				//frame.IP += 4
				//fmt.Println("n=",n)
				current := len(vm.Memory.Memory) / DefaultPageSize
				//fmt.Println(current)
				if vm.Context.Config.MaxMemoryPages == 0 || (current+n >= current && current+n <= vm.Context.Config.MaxMemoryPages) {
					frame.Regs[valueID] = int64(current)
					vm.Memory.Memory = append(vm.Memory.Memory, make([]byte, n*DefaultPageSize)...)
				} else {
					frame.Regs[valueID] = -1
				}
			}
			fmt.Println("GrowMemory", time.Since(start))

		case opcodes.Phi:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				frame.Regs[valueID] = vm.Yielded
			}
			fmt.Println("Phi", time.Since(start))

		case opcodes.AddGas:
			start := time.Now()
			for i := 0; i < 100000000; i++ {
				delta := LE.Uint64(frame.Code[frame.IP : frame.IP+8])
				//frame.IP += 8
				vm.AddAndCheckGas(delta)
			}
			fmt.Println("AddGas", time.Since(start))
		default:
			panic("unknown instruction")
		}
	}
}
