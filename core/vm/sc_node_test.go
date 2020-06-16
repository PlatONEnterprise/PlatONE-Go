package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

type stateDBMock struct {
	database map[string]map[string]interface{}
}

func (m stateDBMock) CreateAccount(address common.Address) {
	panic("implement me")
}

func (m stateDBMock) SubBalance(address common.Address, b *big.Int) {
	panic("implement me")
}

func (m stateDBMock) AddBalance(address common.Address, b *big.Int) {
	panic("implement me")
}

func (m stateDBMock) GetBalance(address common.Address) *big.Int {
	panic("implement me")
}

func (m stateDBMock) GetNonce(address common.Address) uint64 {
	panic("implement me")
}

func (m stateDBMock) SetNonce(address common.Address, u uint64) {
	panic("implement me")
}

func (m stateDBMock) GetCodeHash(address common.Address) common.Hash {
	panic("implement me")
}

func (m stateDBMock) GetCode(address common.Address) []byte {
	panic("implement me")
}

func (m stateDBMock) SetCode(address common.Address, bytes []byte) {
	panic("implement me")
}

func (m stateDBMock) GetCodeSize(address common.Address) int {
	panic("implement me")
}

func (m stateDBMock) GetAbiHash(address common.Address) common.Hash {
	panic("implement me")
}

func (m stateDBMock) GetAbi(address common.Address) []byte {
	panic("implement me")
}

func (m stateDBMock) SetAbi(address common.Address, bytes []byte) {
	panic("implement me")
}

func (m stateDBMock) AddRefund(u uint64) {
	panic("implement me")
}

func (m stateDBMock) SubRefund(u uint64) {
	panic("implement me")
}

func (m stateDBMock) GetRefund() uint64 {
	panic("implement me")
}

func (m stateDBMock) GetCommittedState(address common.Address, bytes []byte) []byte {
	panic("implement me")
}

func (m stateDBMock) GetState(address common.Address, key []byte) []byte {
	panic("implement me")
}

func (m stateDBMock) SetState(address common.Address, bytes []byte, bytes2 []byte) {
	panic("implement me")
}

func (m stateDBMock) Suicide(address common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) HasSuicided(address common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) Exist(address common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) Empty(address common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) RevertToSnapshot(i int) {
	panic("implement me")
}

func (m stateDBMock) Snapshot() int {
	panic("implement me")
}

func (m stateDBMock) AddLog(log *types.Log) {
	panic("implement me")
}

func (m stateDBMock) AddPreimage(hash common.Hash, bytes []byte) {
	panic("implement me")
}

func (m stateDBMock) ForEachStorage(address common.Address, f func(common.Hash, common.Hash) bool) {
	panic("implement me")
}

func (m stateDBMock) FwAdd(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m stateDBMock) FwClear(contractAddr common.Address, action state.Action) {
	panic("implement me")
}

func (m stateDBMock) FwDel(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m stateDBMock) FwSet(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m stateDBMock) SetFwStatus(contractAddr common.Address, status state.FwStatus) {
	panic("implement me")
}

func (m stateDBMock) GetFwStatus(contractAddr common.Address) state.FwStatus {
	panic("implement me")
}

func (m stateDBMock) SetContractCreator(contractAddr common.Address, creator common.Address) {
	panic("implement me")
}

func (m stateDBMock) GetContractCreator(contractAddr common.Address) common.Address {
	panic("implement me")
}

func (m stateDBMock) OpenFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m stateDBMock) CloseFirewall(contractAddr common.Address) {

}

func (m stateDBMock) IsFwOpened(contractAddr common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) FwImport(contractAddr common.Address, data []byte) error {
	panic("implement me")
}

func Test_isMatch(t *testing.T) {
	node := &NodeInfo{}
	query := &NodeInfo{}
	assert.Equal(t, true, isMatch(node, query))
	node.Name = "elvin"
	assert.Equal(t, true, isMatch(node, query))
	query.PublicKey = "aaaaaa"
	assert.Equal(t, false, isMatch(node, query))

	node.PublicKey = "aaaaaa"
	assert.Equal(t, true, isMatch(node, query))
}
