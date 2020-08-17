package cbfttypes

import (
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
)

// Block's Signature info
type BlockSignature struct {
	SignHash  common.Hash // Signature hash，header[0:32]
	Hash      common.Hash // Block hash，header[:]
	Number    *big.Int
	Signature *common.BlockConfirmSign
}

type BlockSynced struct {
}

type CbftResult struct {
	Block *types.Block
	//Receipts          types.Receipts
	//State             *state.StateDB
	BlockConfirmSigns []*common.BlockConfirmSign
}
