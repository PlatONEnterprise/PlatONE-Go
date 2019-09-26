package backend

import (
	"encoding/json"
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus"
	"github.com/PlatONEnetwork/PlatONE-Go/core"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/p2p/discover"
	"math"
	"math/big"
)

var (
	ErrContractNotFound = errors.New("contract not found")
)

type ChainContext struct {
	// Engine retrieves the chain's consensus engine.
	chain *consensus.ChainReader

	engine consensus.Engine
}

func (cc *ChainContext) GetHeader(hash common.Hash, number uint64) *types.Header {
	return cc.GetHeader(hash, number)
}

func (cc *ChainContext) Engine() consensus.Engine {
	return cc.engine
}

type CBFTProduceBlockCfg struct {
	ProduceDuration int32 `json:"ProduceDuration"`
	BlockInterval   int32 `json:"BlockInterval"`
}

type commonResult struct {
	RetCode int32      `json:"code"`
	RetMsg  string     `json:"msg"`
	Data    []nodeInfo `json:"data"`
}

type nodeInfo struct {
	Name       string `json:"name,omitempty"`
	Owner      string `json:"owner,omitempty"`
	Desc       string `json:"desc,omitempty"`
	Types      int32  `json:"type,omitempty"`
	Status     int32  `json:"status,omitempty"`
	ExternalIP string `json:"externalIP,omitempty"`
	InternalIP string `json:"internalIP,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
	RpcPort    int32  `json:"rpcPort,omitempty"`
	P2pPort    int32  `json:"p2pPort,omitempty"`
}

// getInitialNodesList catch initial nodes List from paramManager contract when
// new a dpos and miner a new block
func getConsensusNodesList(chain consensus.ChainReader, sb *backend, headers []*types.Header, number uint64) ([]discover.NodeID, error) {
	var tmp []common.NodeInfo
	if common.SysCfg == nil {
		loadLastConsensusNodesList(chain, sb, headers)

		common.SysCfg.UpdateSystemConfig()
		log.Info("UpdateSystemConfig successful in getConsensusNodesList function")
	}

	tmp = common.SysCfg.GetConsensusNodesFilterDelay(number)
	nodeIDs := make([]discover.NodeID, 0, len(tmp))
	for _, dataObj := range tmp {
		if pubKey := dataObj.PublicKey; len(pubKey) > 0 {
			log.Debug("Consensus node", "PublicKey", pubKey)
			if nodeID, err := discover.HexID(pubKey); err == nil {
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
	}

	return nodeIDs, nil
}

func loadLastConsensusNodesList(chain consensus.ChainReader, sb *backend, headers []*types.Header) {

	innerCall := func(conAddr common.Address, data []byte) ([]byte, error) {
		//ctx := context.Background()
		if sb == nil {
			log.Info("backend is nil")
		}

		log.Info("this is for test loadLastConsensusNodesList", chain.CurrentHeader().Root)
		// Get the state
		state, err := state.New(chain.CurrentHeader().Root, state.NewDatabase(sb.db))
		if state == nil {
			return nil,err
		}

		from := common.Address{}
		to := &conAddr
		gas := uint64(0x999999999)
		gasPrice := (hexutil.Big)(*big.NewInt(0x333333))
		nonce := uint64(0)
		value := (hexutil.Big)(*big.NewInt(0))

		// Create new call message
		msg := types.NewMessage(from, to, nonce, value.ToInt(), gas, gasPrice.ToInt(), data, false, types.NormalTxType)
		cc := ChainContext{&chain, sb}
		context := core.NewEVMContext(msg, chain.CurrentHeader(), &cc, nil)

		evm := vm.NewEVM(context, state, chain.Config(), vm.Config{})

		// Get a new instance of the EVM.
		//evm, vmError, err := ethPtr.APIBackend.GetEVM(ctx, msg, state, header, vm.Config{})
		//if err != nil {
		//	return nil
		//}

		// Setup the gas pool (also for unmetered requests)
		// and apply the message.
		gp := new(core.GasPool).AddGas(math.MaxUint64)
		res, _, _, err := core.ApplyMessage(evm, msg, gp)
		if err != nil {
			return nil, err
		}

		return res, err
	}

	sysContractCall := func(sc *common.SystemConfig) {
		//ctx := context.Background()

		state, _ := state.New(chain.CurrentHeader().Root, state.NewDatabase(sb.db))
		// Get the state
		if state == nil {
			return
		}

		// Create new call message
		msg := types.NewMessage(common.Address{}, nil, 1, big.NewInt(1), 0x1, big.NewInt(1), nil, false, types.NormalTxType)

		// Get a new instance of the EVM.

		cc := ChainContext{&chain, sb}
		context := core.NewEVMContext(msg, chain.CurrentHeader(), &cc, nil)

		evm := vm.NewEVM(context, state, chain.Config(), vm.Config{})

		//evm, vmError, err := ethPtr.APIBackend.GetEVM(ctx, msg, state, header, vm.Config{})
		//if err != nil {
		//	return
		//}

		// clusure method for call Contract
		callContract := func(conAddr common.Address, data []byte) []byte {
			res, _, err := evm.Call(vm.AccountRef(common.Address{}), conAddr, data, uint64(0xffffffffff), big.NewInt(0))
			if err != nil {
				return nil
			}
			return res
		}

		// Get all system contracts' address
		var fh string = "getContractAddress"

		systemContractList := []string{"__sys_NodeManager",
			"__sys_NodeRegister",
			"__sys_UserRegister",
			"__sys_UserManager",
			"__sys_ParamManager",
			"__sys_RoleManager",
			"__sys_RoleRegister"}

		// Update system contract address
		for _, contractName := range systemContractList {
			callParams := []interface{}{contractName, "latest"}
			btsRes := callContract(common.HexToAddress(core.CnsManagerAddr), common.GenCallData(fh, callParams))
			strRes := common.CallResAsString(btsRes)
			if !(len(strRes) == 0 || common.IsHexZeroAddress(strRes)) {
				sc.ContractAddress[contractName] = common.HexToAddress(strRes)
			}
		}

		// Get contract parameters from contract
		paramAddr := sc.ContractAddress["__sys_ParamManager"]
		if paramAddr != (common.Address{}) {
			funcName := "getTxGasLimit"
			funcParams := []interface{}{}
			res := callContract(paramAddr, common.GenCallData(funcName, funcParams))
			if res != nil {
				ret := common.CallResAsInt64(res)
				if ret > 0 {
					sc.SysParam.TxGasLimit = ret
				}
			}
			funcName = "getBlockGasLimit"
			funcParams = []interface{}{}
			res = callContract(paramAddr, common.GenCallData(funcName, funcParams))
			if res != nil {
				ret := common.CallResAsInt64(res)
				if ret > 0 {
					sc.SysParam.BlockGasLimit = ret
				}
			}
			funcName = "getCBFTTimeParam"
			funcParams = []interface{}{}
			res = callContract(paramAddr, common.GenCallData(funcName, funcParams))
			if res != nil {
				strRes := common.CallResAsString(res)

				var cbftCfgTime common.CBFTProduceBlockCfg
				if err := json.Unmarshal([]byte(strRes), &cbftCfgTime); err != nil {
					log.Error("contract return invalid data", "result", strRes, "err", err.Error())
				} else {
					sc.SysParam.CBFTTime = cbftCfgTime
				}
			}
			funcName = "getGasContractName"
			funcParams = []interface{}{}
			res = callContract(paramAddr, common.GenCallData(funcName, funcParams))
			if res != nil {
				sc.SysParam.GasContractName = common.CallResAsString(res)
			}
		}

		if sc.SysParam.GasContractName != "" {
			cnsAddr := common.HexToAddress(core.CnsManagerAddr)
			funcName := "getContractAddress"
			funcParams := []interface{}{sc.SysParam.GasContractName, "latest"}
			res := callContract(cnsAddr, common.GenCallData(funcName, funcParams))
			if res != nil {
				sc.SysParam.GasContractAddr = common.HexToAddress(common.CallResAsString(res))
			}
		}

		// Get nodes from contract
		nodeManagerAddr := sc.ContractAddress["__sys_NodeManager"]
		if nodeManagerAddr != (common.Address{}) {
			funcName := "getAllNodes"
			funcParams := []interface{}{}
			res := callContract(nodeManagerAddr, common.GenCallData(funcName, funcParams))
			if res != nil {
				sc.SysParam.GasContractAddr = common.HexToAddress(common.CallResAsString(res))
			}

			strRes := common.CallResAsString(res)

			var tmp common.CommonResult
			if err := json.Unmarshal(utils.String2bytes(strRes), &tmp); err != nil {
				log.Warn("unmarshal consensus node list failed", "result", strRes, "err", err.Error())
			} else if tmp.RetCode != 0 {
				log.Info("contract inner error", "code", tmp.RetCode, "msg", tmp.RetMsg)
			} else {
				sc.Nodes = tmp.Data
			}
		}
		return
	}

	common.InitSystemconfig(common.NodeInfo{})
	common.SetSysContractCallFunc(sysContractCall)
	common.SetInnerCallFunc(innerCall)

}
