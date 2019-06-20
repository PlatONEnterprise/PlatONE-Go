package exec
//
//import (
//	"github.com/PlatONEnetwork/PlatONE-Go/life/compiler/opcodes"
//	"testing"
//)
//
//func TestCfcSet(t *testing.T) {
//	_2ParamTest()
//	//_loadTest()
//	//_storeTest()
//}
//
//// _2ParamTest(){}
//func _2ParamTest() {
//
//	// locals
//	locals := []int64{0, 32, 97,128}
//	//ops:=[]opcodes.Opcode{opcodes.Phi,opcodes.I32Const, opcodes.I32PopCnt, opcodes.I32EqZ, opcodes.GetGlobal,
//	//	opcodes.GetLocal, opcodes.I32Clz,
//	//	opcodes.I32Ctz, opcodes.ReturnValue, opcodes.I64Clz, opcodes.I64Ctz, opcodes.I64PopCnt, opcodes.I64EqZ,
//	//	opcodes.F64Sqrt, opcodes.F32Sqrt, opcodes.F32Ceil, opcodes.F32Floor, opcodes.F32Trunc, opcodes.F32Nearest,
//	//	opcodes.F32Abs, opcodes.F32Neg, opcodes.F64Ceil, opcodes.F64Floor, opcodes.F64Trunc, opcodes.F64Nearest,
//	//	opcodes.F64Abs, opcodes.F64Neg, opcodes.I32WrapI64, opcodes.I32TruncUF32, opcodes.I32TruncSF32,
//	//	opcodes.I32TruncSF64, opcodes.I32TruncUF64, opcodes.I64TruncUF32, opcodes.I64TruncUF64, opcodes.I64TruncSF32,
//	//	opcodes.I64TruncSF64, opcodes.F32DemoteF64, opcodes.F64PromoteF32, opcodes.F32ConvertSI32, opcodes.F32ConvertUI32,
//	//	opcodes.F32ConvertSI64, opcodes.F32ConvertUI64, opcodes.F64ConvertSI32, opcodes.F64ConvertSI64,
//	//	opcodes.F64ConvertUI32, opcodes.F64ConvertUI64, opcodes.I64ExtendUI32, opcodes.I64ExtendSI32,
//	//	opcodes.I32Add, opcodes.I32Sub, opcodes.I32Mul, opcodes.I32DivS, opcodes.I32DivU, opcodes.I32RemS,
//	//	opcodes.I32RemU, opcodes.I32And, opcodes.I32Or, opcodes.I32Xor, opcodes.I32Shl, opcodes.I32ShrS,
//	//	opcodes.I32ShrU, opcodes.I32Rotl, opcodes.I32Rotr, opcodes.I32Eq, opcodes.I32Ne, opcodes.I32LtS,
//	//	opcodes.I32LtU, opcodes.I32LeS, opcodes.I32LeU, opcodes.I32GtS, opcodes.I32GtU, opcodes.I32GeS,
//	//	opcodes.I32GeU, opcodes.I64Add, opcodes.I64Sub, opcodes.I64Mul, opcodes.I64DivS, opcodes.I64DivU,
//	//	opcodes.I64RemS, opcodes.I64RemU, opcodes.I64And, opcodes.I64Or, opcodes.I64Xor, opcodes.I64Shl,
//	//	opcodes.I64ShrS, opcodes.I64ShrU, opcodes.I64Rotl, opcodes.I64Rotr, opcodes.I64Eq, opcodes.I64Ne,
//	//	opcodes.I64LtS, opcodes.I64LtU, opcodes.I64LeS, opcodes.I64LeU, opcodes.I64GtS, opcodes.I64GtU, opcodes.I64GeS,
//	//	opcodes.I64GeU, opcodes.F32Add, opcodes.F32Sub, opcodes.F32Mul, opcodes.F32Div, opcodes.F32Min, opcodes.F32Max,
//	//	opcodes.F32CopySign, opcodes.F32Eq, opcodes.F32Ne, opcodes.F32Lt, opcodes.F32Le, opcodes.F32Gt, opcodes.F32Ge,
//	//	opcodes.F64Add, opcodes.F64Sub, opcodes.F64Mul, opcodes.F64Div, opcodes.F64Min, opcodes.F64Max, opcodes.F64CopySign,
//	//	opcodes.F64Eq, opcodes.F64Ne, opcodes.F64Lt, opcodes.F64Le, opcodes.F64Gt, opcodes.F64Ge, opcodes.Jmp,
//	//	opcodes.SetLocal, opcodes.SetGlobal}
//
//	ops := []opcodes.Opcode{opcodes.GrowMemory}
//
//	for _, op := range ops {
//		for i := 0; i < 10; i++ {
//			lvm := NewMockedVm()
//
//			ip := 0
//			code := make([]byte, 0)
//			code = append(code, []byte{0, 0, 0, 0}...) // entryID
//			code = append(code, byte(op))              // opcode
//			code = append(code, []byte{1, 0, 0, 0}...) // a's position
//			code = append(code, []byte{1, 0, 0, 0}...) // b's position
//			code = append(code, []byte{1, 0, 0, 0}...) // c's position
//			code = append(code, []byte{4, 0, 0, 0}...) // d's position
//			code = append(code, []byte{5, 0, 0, 0}...) // e's position
//
//			code = append(code, []byte{0, 0, 0, 3}...)    // entryID
//			code = append(code, byte(opcodes.ReturnVoid)) // opcode
//
//			regs := make([]int64, 10)
//			regs[0] = 6 //a
//			regs[1] = 1 // b
//			regs[2] = 8 // c
//			regs[3] = 9 // d
//
//
//			lvm.SetMockedFrame(code, regs, locals, ip, 0, 0)
//
//			// excute
//			memory := make([]byte,65537)
//			for i:= range memory{
//				memory[i] = 2
//			}
//			m := Memory{
//				Memory:memory,
//			}
//			lvm.SetMockerMemory(&m)
//
//			lvm.Execute()
//		}
//	}
//
//
//}
//
//// _loadTest
//func _loadTest() {
//
//	// locals
//	locals := []int64{0, 32, 97}
//
//	//ops := []opcodes.Opcode{opcodes.I32Load, opcodes.I64Load, opcodes.I32Store,	opcodes.I64Store,
//	//	opcodes.I32Load8S,	opcodes.I32Load16S,	opcodes.I64Load8S,	opcodes.I64Load16S,
//	//	opcodes.I64Load32S,	opcodes.I32Load8U,	opcodes.I32Load16U,	opcodes.I64Load8U,
//	//	opcodes.I64Load16U,	opcodes.I64Load32U}
//	ops :=[]opcodes.Opcode{opcodes.I32Store8,	opcodes.I32Store16,	opcodes.I64Store8,	opcodes.I64Store16,	opcodes.I64Store32}
//
//	for _, op := range ops {
//		for i := 0; i < 10; i++ {
//			lvm := NewMockedVm()
//
//			ip := 0
//			code := make([]byte, 0)
//
//			// populate opcode
//			code = append(code, []byte{0, 0, 0, 0}...) // entryID
//			code = append(code, byte(op))              // opcode
//			code = append(code, []byte{0, 0, 0, 0}...) // not used
//			code = append(code, []byte{0, 0, 0, 0}...) // offset
//			code = append(code, []byte{2, 0, 0, 0}...) // base position
//
//			// return
//			code = append(code, []byte{0, 0, 0, 3}...)    // entryID
//			code = append(code, byte(opcodes.ReturnVoid)) // opcode
//
//			regs := make([]int64, 4)
//
//			regs[2] = 0 // base value
//
//			lvm.SetMockedFrame(code, regs, locals, ip, 0, 0)
//
//			// memory
//			memory := []byte{1,1,0,0,0,0,0,0}
//
//			m := Memory{
//				Memory:memory,
//			}
//			lvm.SetMockerMemory(&m)
//
//			// excute
//			lvm.Execute()
//		}
//	}
//
//
//}
//
//// _storeTest
//func _storeTest() {
//
//	// locals
//	locals := []int64{0, 32, 97}
//
//	//ops := []opcodes.Opcode{opcodes.I32Store8,	opcodes.I32Store16,	opcodes.I64Store8,	opcodes.I64Store16,	opcodes.I64Store32}
//	ops := []opcodes.Opcode{opcodes.I64Store}
//	for _, op := range ops {
//		for i := 0; i < 10; i++ {
//			lvm := NewMockedVm()
//
//			ip := 0
//			code := make([]byte, 0)
//
//			// populate opcode
//			code = append(code, []byte{0, 0, 0, 0}...) // entryID
//			code = append(code, byte(op))              // opcode
//			code = append(code, []byte{0, 0, 0, 0}...) // not used
//			code = append(code, []byte{0, 0, 0, 0}...) // offset
//			code = append(code, []byte{2, 0, 0, 0}...) // base position
//			code = append(code, []byte{3, 0, 0, 0}...) // value position
//
//			// return
//			code = append(code, []byte{0, 0, 0, 3}...)    // entryID
//			code = append(code, byte(opcodes.ReturnVoid)) // opcode
//
//			regs := make([]int64, 4)
//
//			regs[2] = 0 // base value
//			regs[3] = 257 // base value
//
//			lvm.SetMockedFrame(code, regs, locals, ip, 0, 0)
//
//			// memory
//			memory := []byte{0,0,0,0,0,0,0,0}
//
//			m := Memory{
//				Memory:memory,
//			}
//			lvm.SetMockerMemory(&m)
//
//			// excute
//			lvm.Execute()
//		}
//	}
//
//
//}
//
