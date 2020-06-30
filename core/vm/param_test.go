package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"math/big"
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
	if nil != err{
		t.Error(err)
		return
	}
	db := newMockDB()
	addr := syscontracts.PARAMETER_MANAGEMENT_ADDRESS
	db.SetState(addr, bin, bin)

	res := db.GetState(addr, bin)
	t.Logf("%b",res)
}

func TestParamManager_getTxLimit(t *testing.T) {
	db := newMockDB()
	addr := syscontracts.PARAMETER_MANAGEMENT_ADDRESS
	p := ParamManager{CodeAddr:&addr, StateDB: db}
	p.setBlockGasLimit(12771599*100)
	ret, err := p.getBlockGasLimit()
	if nil != err{
			t.Error(err)
			return
		}
	t.Logf("%d",ret)
}

//func TestParamManager_getFn(t *testing.T) {
//	db := newMockDB()
//	addr := syscontracts.PARAMETER_MANAGEMENT_ADDRESS
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
	a := "0"
	bin, err := encode(a)
	if nil != err{
		t.Error(err)
		return
	}
	t.Logf("%b",bin)
	t.Logf("%v",string(bin))
}

func Test_decode(t *testing.T) {
	a := "abc"
	bin, err := encode(a)
	if nil != err{
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
	t.Logf("%v",bin)
}


type mockDB struct{
	mockDB1	map[string]interface{}
}

func newMockDB() *mockDB {
	return &mockDB{mockDB1: make(map[string]interface{})}
}

func (m *mockDB) CreateAccount(common.Address) {
	panic("implement me")
}

func (m *mockDB) SubBalance(common.Address, *big.Int) {
	panic("implement me")
}

func (m *mockDB) AddBalance(common.Address, *big.Int) {
	panic("implement me")
}

func (m *mockDB) GetBalance(common.Address) *big.Int {
	panic("implement me")
}

func (m *mockDB) GetNonce(common.Address) uint64 {
	panic("implement me")
}

func (m *mockDB) SetNonce(common.Address, uint64) {
	panic("implement me")
}

func (m *mockDB) GetCodeHash(common.Address) common.Hash {
	panic("implement me")
}

func (m *mockDB) GetCode(common.Address) []byte {
	panic("implement me")
}

func (m *mockDB) SetCode(common.Address, []byte) {
	panic("implement me")
}

func (m *mockDB) GetCodeSize(common.Address) int {
	panic("implement me")
}

func (m *mockDB) GetAbiHash(common.Address) common.Hash {
	panic("implement me")
}

func (m *mockDB) GetAbi(common.Address) []byte {
	panic("implement me")
}

func (m *mockDB) SetAbi(common.Address, []byte) {
	panic("implement me")
}

func (m *mockDB) AddRefund(uint64) {
	panic("implement me")
}

func (m *mockDB) SubRefund(uint64) {
	panic("implement me")
}

func (m *mockDB) GetRefund() uint64 {
	panic("implement me")
}

func (m *mockDB) GetCommittedState(common.Address, []byte) []byte {
	panic("implement me")
}

func (m *mockDB) GetState(address common.Address, key []byte) []byte {
	//panic("implement me")
	if bin, ok := m.mockDB1[string(key)]; ok {
		return bin.([]byte)
	}

	return nil
}

func (m *mockDB) SetState(address common.Address, key []byte, val []byte) {
	//panic("implement me")
	m.mockDB1[string(key)] = val
}

func (m *mockDB) Suicide(common.Address) bool {
	panic("implement me")
}

func (m *mockDB) HasSuicided(common.Address) bool {
	panic("implement me")
}

func (m *mockDB) Exist(common.Address) bool {
	panic("implement me")
}

func (m *mockDB) Empty(common.Address) bool {
	panic("implement me")
}

func (m *mockDB) RevertToSnapshot(int) {
	panic("implement me")
}

func (m *mockDB) Snapshot() int {
	panic("implement me")
}

func (m *mockDB) AddLog(*types.Log) {
	panic("implement me")
}

func (m *mockDB) AddPreimage(common.Hash, []byte) {
	panic("implement me")
}

func (m *mockDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {
	panic("implement me")
}

func (m *mockDB) FwAdd(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockDB) FwClear(contractAddr common.Address, action state.Action) {
	panic("implement me")
}

func (m *mockDB) FwDel(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockDB) FwSet(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockDB) SetFwStatus(contractAddr common.Address, status state.FwStatus) {
	panic("implement me")
}

func (m *mockDB) GetFwStatus(contractAddr common.Address) state.FwStatus {
	panic("implement me")
}

func (m *mockDB) SetContractCreator(contractAddr common.Address, creator common.Address) {
	panic("implement me")
}

func (m *mockDB) GetContractCreator(contractAddr common.Address) common.Address {
	panic("implement me")
}

func (m *mockDB) OpenFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m *mockDB) CloseFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m *mockDB) IsFwOpened(contractAddr common.Address) bool {
	panic("implement me")
}

func (m *mockDB) FwImport(contractAddr common.Address, data []byte) error {
	panic("implement me")
}
