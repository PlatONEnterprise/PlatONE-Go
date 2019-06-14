//package resolver
//
//import (
//	"fmt"
//	"github.com/PlatONEnetwork/PlatONE-Go/life/exec"
//	"github.com/PlatONEnetwork/PlatONE-Go/log"
//	"testing"
//	"time"
//)
//
//func TestCfcSet(t *testing.T) {
//	//cfgSet := newCfcSet()
//	//for k, v := range cfgSet {
//	//	fmt.Println("key:", k)
//	//	for k1, v1 := range v {
//	//		fmt.Printf("key1: %v, v1 type: %T \n", k1, v1)
//	//	}
//	//}
//	ecrecoverTest()
//}
//
//// ecrecover test
//// the hash of msg: 1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8
//// signature: da9b32d0d0d919272f2b1f2306bd20c51f0b5cdd6c9cadafb76fda887889a70125f54bb64d3104a888f36b00e944dcccc445e128f90247a36df3e666fcca93f901
//// address: 0x9ca9eeE03a26820940E0414072a8b5e2Ca662798
//func ecrecoverTest(){
//
//
//	//r:=NewResolver(0x01)
//	//lvm := exec.NewMockedVm("ecrecover", r)
//	for j :=0; j<10;j++ {
//		lvm := exec.NewMockedVm()
//
//		//locals := []int64{3, 32,64, 97,129}
//		locals := []int64{0,288,288,32,320,32,352,32,384,65,449,65,514,65}
//		lvm.SetMockedFrame(nil, nil, locals, 0, 0, 0)
//
//		// hash of message
//		//h, _ := hex.DecodeString("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
//		//sig, _ := hex.DecodeString("da9b32d0d0d919272f2b1f2306bd20c51f0b5cdd6c9cadafb76fda887889a70125f54bb64d3104a888f36b00e944dcccc445e128f90247a36df3e666fcca93f901")
//		//addr := make([]byte, 20)
//		////memory := make([]byte,0)
//		//memory := append(h[:],sig...)
//		//memory = append(memory, addr...)
//		//memory := make([]byte, 10240)
//
//		//random number
//		//r := []byte{77}
//		//g := []byte{55}
//		//n := []byte{6,227}
//		//m1 := []byte{2}
//		//c := []byte{143,126}
//		//memory := make([]byte,0)
//		//memory = append(memory,r...)
//		//memory = append(memory,g...)
//		//memory = append(memory,n...)
//		//memory = append(memory,m1...)
//		//memory = append(memory,c...)
//
//		memory := make([]byte,1024)
//		for i := range memory{
//			memory[i] = 1
//		}
//
//		m := exec.Memory{
//			Memory: memory,
//		}
//		lvm.SetMockerMemory(&m)
//
//		// excute
//		start := time.Now()
//		for i := 0; i < 100000; i++ {
//			envNizkVerifyProof(lvm)
//		}
//		fmt.Println("envNizkVerifyProof",time.Since(start))
//	}
//
//	//addr = lvm.Memory.Memory[97:117]
//	//fmt.Println(hex.EncodeToString(addr))
//}
//
//
////funcTest := []interface{}{envEcrecover}
////for index, funcname := range funcTest{
////p := funcname.(func(*exec.VirtualMachine)int64)
////start := time.Now()
////p(lvm)
////fmt.Println(funcTest[index],time.Since(start))
////}
//
//
////type serviceHandle func(*exec.VirtualMachine) int64
////names := []serviceHandle{envMemcpy}
////for index,funcname := range names{
////	start := time.Now()
////	funcname(lvm)
////	fmt.Println(names[index],time.Since(start))
////}
//
//
//func mallocTest(){
//	for j :=0; j<10;j++ {
//		lvm := exec.NewMockedVm()
//
//
//		locals := []int64{0,5}
//		lvm.SetMockedFrame(nil, nil, locals, 0, 0, 0)
//
//
//		start := time.Now()
//		for i :=0; i < 100000; i++ {
//			lvm.SetMockerMemory(exec.NewMockedMemory())
//
//			// excute
//
//			envRealloc(lvm)
//		}
//		fmt.Println("envRealloc",time.Since(start))
//
//	}
//
//}
//
//func printsTest(){
//	lvm := exec.NewMockedVm()
//
//	locals := []int64{128}
//	lvm.SetMockedFrame(nil,nil, locals, 0,0,0)
//
//
//	memory := make([]byte,128)
//	for i:= range memory{
//		memory[i]=31
//	}
//
//	m := exec.Memory{
//		Memory:memory,
//	}
//	lvm.SetMockerMemory(&m)
//
//	lvm.Context = &exec.VMContext{
//		Log: log.Logger.New(log.New("wasmlog")),
//	}
//
//	// excute
//	start := time.Now()
//	for i:=0; i<10000; i++{
//		envPrints(lvm)
//	}
//	fmt.Println("envPrints",time.Since(start))
//}
//
//
