package vm

import (
	"math/big"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
)

func TestMain(m *testing.M) {
	// initial the data needed for sc_cns_test.go and sc_cns_db_test.go
	cnsTestInitial()

	m.Run()
}

func newMockStateDB() *mockStateDB {
	return &mockStateDB{
		mockDB: make(map[common.Address]map[string][]byte),
		eLogs:  make(map[string]*types.Log),
	}
}

type mockStateDB struct {
	mockDB map[common.Address]map[string][]byte
	eLogs  map[string]*types.Log
}

func (m *mockStateDB) CloneAccount(src common.Address, dest common.Address) error {
	panic("implement me")
}

func (m *mockStateDB) GetState(addr common.Address, key []byte) []byte {

	return m.mockDB[addr][string(key)]
}

func (m *mockStateDB) SetState(addr common.Address, key []byte, value []byte) {

	if m.mockDB[addr] == nil {
		m.mockDB[addr] = make(map[string][]byte)
	}

	m.mockDB[addr][string(key)] = value
}

func (m *mockStateDB) GetContractCreator(contractAddr common.Address) common.Address {
	return testOrigin
}

func (m *mockStateDB) CreateAccount(common.Address) {
	panic("implement me")
}

func (m *mockStateDB) SubBalance(common.Address, *big.Int) {
	panic("implement me")
}

func (m *mockStateDB) AddBalance(common.Address, *big.Int) {
	panic("implement me")
}

func (m *mockStateDB) GetBalance(common.Address) *big.Int {
	panic("implement me")
}

func (m *mockStateDB) GetNonce(common.Address) uint64 {
	panic("implement me")
}

func (m *mockStateDB) SetNonce(common.Address, uint64) {
	panic("implement me")
}

func (m *mockStateDB) GetCodeHash(common.Address) common.Hash {
	panic("implement me")
}

func (m *mockStateDB) GetCode(common.Address) []byte {
	panic("implement me")
}

func (m *mockStateDB) SetCode(common.Address, []byte) {
	panic("implement me")
}

func (m *mockStateDB) GetCodeSize(common.Address) int {
	panic("implement me")
}

func (m *mockStateDB) GetAbiHash(common.Address) common.Hash {
	panic("implement me")
}

func (m *mockStateDB) GetAbi(common.Address) []byte {
	panic("implement me")
}

func (m *mockStateDB) SetAbi(common.Address, []byte) {
	panic("implement me")
}

func (m *mockStateDB) AddRefund(uint64) {
	panic("implement me")
}

func (m *mockStateDB) SubRefund(uint64) {
	panic("implement me")
}

func (m *mockStateDB) GetRefund() uint64 {
	panic("implement me")
}

func (m *mockStateDB) GetCommittedState(common.Address, []byte) []byte {
	panic("implement me")
}

func (m *mockStateDB) Suicide(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) HasSuicided(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) Exist(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) Empty(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) RevertToSnapshot(int) {
	panic("implement me")
}

func (m *mockStateDB) Snapshot() int {
	panic("implement me")
}

func (m *mockStateDB) AddLog(log *types.Log) {
	m.eLogs[log.Topics[0].String()] = log
}

func (m *mockStateDB) AddPreimage(common.Hash, []byte) {
	panic("implement me")
}

func (m *mockStateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {
	panic("implement me")
}

func (m *mockStateDB) FwAdd(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) FwClear(contractAddr common.Address, action state.Action) {
	panic("implement me")
}

func (m *mockStateDB) FwDel(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) FwSet(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) SetFwStatus(contractAddr common.Address, status state.FwStatus) {
	panic("implement me")
}

func (m *mockStateDB) GetFwStatus(contractAddr common.Address) state.FwStatus {
	panic("implement me")
}

func (m *mockStateDB) SetContractCreator(contractAddr common.Address, creator common.Address) {
	panic("implement me")
}

func (m *mockStateDB) OpenFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m *mockStateDB) CloseFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m *mockStateDB) IsFwOpened(contractAddr common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) FwImport(contractAddr common.Address, data []byte) error {
	panic("implement me")
}
