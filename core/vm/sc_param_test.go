package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"testing"
)

//func TestParamManager_get(t *testing.T) {
//	type paramManagerTest struct {
//		Contract *Contract
//		Evm      *EVM
//	}
//	tests := []struct {
//		name    string
//		param paramManagerTest
//	}{
//
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p := &ParamManager{
//				Contract: tt.param.Contract,
//				Evm:      tt.param.Evm,
//			}
//			ls := "abc"
//			got, err := p. getGasContractName()
//			t.Logf("%s\n", got)
//			t.Logf("%s\n", ls)
//			if err != nil {
//				t.Errorf("can't find")
//				return
//			}else {
//				t.Logf("%b\n", got)
//				t.Logf("%s\n", ls)
//			}
//		})
//	}
//}

//func TestParamManager_get(t *testing.T) {
//	type paramManagerTest struct {
//		Contract *Contract
//		Evm      *EVM
//	}
//	param := paramManagerTest{}
//	p := &ParamManager{
//		Contract: param.Contract,
//		Evm:      param.Evm,
//	}
//	set, err := p.setGasContractName("abc")
//	if err != nil {
//		t.Errorf("can't find")
//		return
//	}else {
//		t.Logf("%b\n", set)
//
//	}
//
//}

//func TestParamManager_set(t *testing.T) {
//	type paramManagerTest struct {
//		Contract *Contract
//		Evm      *EVM
//	}
//	tests := []struct {
//		name    string
//		param paramManagerTest
//	}{
//		{
//			"abc",
//			paramManagerTest{},
//		},
//	}
//	//ls1 := "abc"
//	//t.Logf("%v\n", ls1)
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p := &ParamManager{
//				Contract: tt.param.Contract,
//				Evm:      tt.param.Evm,
//			}
//			//ls1 := "abc"
//			//t.Logf("%v\n", ls1)
//			ls := "abc"
//			got, err := p.getGasContractName()
//			t.Logf("%b\n", got)
//			t.Logf("%v\n", ls)
//			if err != nil {
//				t.Errorf("can't find")
//				return
//			}else {
//				t.Logf("%b\n", got)
//				t.Logf("%s\n", ls)
//			}
//		})
//	}
//}
func TestParamManager_stateDB(t *testing.T) {
	a := "0123"
	bin, err := encode(a)
	if nil != err {
		t.Error(err)
		return
	}
	db := newMockStateDB()
	addr := syscontracts.ParameterManagementAddress
	db.SetState(addr, bin, bin)

	res := db.GetState(addr, bin)
	t.Logf("%b", res)
}

func TestParamManager_getTxLimit(t *testing.T) {
	db := newMockStateDB()
	addr := syscontracts.ParameterManagementAddress
	addr1 := syscontracts.UserManagementAddress
	caller := common.HexToAddress("0x62fb664c49cfa4fa35931760c704f9b3ab664666")
	um := UserManagement{state: db, caller: caller, address: addr1}
	um.setSuperAdmin()
	um.addChainAdminByAddress(caller)
	p := ParamManager{codeAddr: &addr, state: db, callerAddr: caller}
	p.setIsApproveDeployedContract(1 )
	ret, err := p.getIsApproveDeployedContract()
	if nil != err {
		t.Error(err)
		return
	}
	t.Logf("%d", ret)
}

//func TestParamManager_getFn(t *testing.T) {
//	db := newMockDB()
//	addr := syscontracts.ParameterManagementAddress
//	p := ParamManager{CodeAddr:&addr, StateDB: db}
//	set := "abc"
//	res, err := p.setGasContractName(set)
//	if nil != err{
//		t.Error(err)
//		return
//	}
//	t.Logf("%b",res)
//
//	res, err = p.getGasContractName()
//	if nil != err{
//		t.Error(err)
//		return
//	}
//	var ci string
//	if err := rlp.DecodeBytes(res, &ci); nil != err {
//		return
//	}
//
//	//var rea string
//	//if err = rlp.DecodeBytes(res, rea); nil != err {
//	//	//t.Logf("abc")
//	//	//t.Logf("%v",rea)
//	//	t.Logf("err")
//	//	return
//	//}
//	t.Logf("%v",ci)
//
//}

func Test_encode(t *testing.T) {
	a := uint32(8)
	bin, err := encode(a)
	if nil != err {
		t.Error(err)
		return
	}
	t.Logf("%b", bin)
	t.Logf("%v", string(bin))
}

func Test_decode(t *testing.T) {
	a := "abc"
	bin, err := encode(a)
	if nil != err {
		t.Error(err)
		return
	}
	//t.Logf("%b",bin)
	//t.Logf("%v",string(bin))
	//var bin1 string
	t.Logf("abc")
	//if err := rlp.DecodeBytes(bin, bin1); nil != err {
	//	t.Logf("abc")
	//	t.Logf("%v",bin1)
	//	return
	//}
	//bin1
	t.Logf("%v", bin)
}
