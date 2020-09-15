package util

import (
	"data-manager/config"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"math/big"
)

func Sender(tx *types.Transaction) (common.Address, error) {
	//first try Frontier
	signer := types.FrontierSigner{}
	addr, err := signer.Sender(tx)
	if nil == err {
		return addr, nil
	}

	addr, err = types.NewEIP155Signer(big.NewInt(0).SetUint64(config.Config.ChainConf.ID)).Sender(tx)
	if nil == err {
		return addr, nil
	}

	return types.HomesteadSigner{}.Sender(tx)
}

