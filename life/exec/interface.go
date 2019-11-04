package exec

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"math/big"
)

type StateDB interface {
	GasPrice() int64
	BlockHash(num uint64) common.Hash
	BlockNumber() *big.Int
	GasLimimt() uint64
	Time() *big.Int
	Coinbase() common.Address
	GetBalance(addr common.Address) *big.Int
	Origin() common.Address
	Caller() common.Address
	GetCode(addr common.Address) []byte
	Address() common.Address
	CallValue() *big.Int
	IsOwner(contractAddress common.Address, accountAddress common.Address) int64
	AddLog(address common.Address, topics []common.Hash, data []byte, bn uint64)
	SetState(key []byte, value []byte)
	GetState(key []byte) []byte

	GetCallerNonce() int64
	Transfer(addr common.Address, value *big.Int) (ret []byte, leftOverGas uint64, err error)
	DelegateCall(addr, params []byte) ([]byte, error)
	Call(addr, params []byte) ([]byte, error)
}
