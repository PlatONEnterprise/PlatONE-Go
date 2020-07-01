package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

func emitEvent(address common.Address, stateDB StateDB, bn uint64, topic string, params ...interface{}) error {
	eLog := types.Log{}
	eLog.Address = address
	eLog.Topics = []common.Hash{common.BytesToHash(crypto.Keccak256([]byte(topic)))}
	eLog.BlockNumber = bn

	bin, err := rlp.EncodeToBytes(params)
	if nil != err {
		return err
	}

	eLog.Data = bin

	stateDB.AddLog(&eLog)

	return nil
}
