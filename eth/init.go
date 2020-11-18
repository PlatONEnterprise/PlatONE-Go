package eth

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core"
)

func InitInnerCallFunc(bc *core.BlockChain) {
	updateParamSCConfFunc := func(sysContractConf *common.SystemConfig) {
		core.UpdateParamSysContractConfig(bc, sysContractConf)
	}

	updateNodeSCConfFunc := func(sysContractConf *common.SystemConfig) {
		core.UpdateNodeSysContractConfig(bc, sysContractConf)
	}

	common.SetParamSysContractUpdateFunc(updateParamSCConfFunc)
	common.SetNodeSysContractUpdateFunc(updateNodeSCConfFunc)
	common.InitSystemconfig(common.NodeInfo{})
}
