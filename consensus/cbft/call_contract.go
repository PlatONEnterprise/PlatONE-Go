package cbft

import (
	//"encoding/json"
	"errors"
	"github.com/BCOSnetwork/BCOS-Go/common"
	"github.com/BCOSnetwork/BCOS-Go/log"
	"github.com/BCOSnetwork/BCOS-Go/p2p/discover"
	"github.com/BCOSnetwork/BCOS-Go/params"
)

var (
	ErrContractNotFound = errors.New("contract not found")
)

type CBFTProduceBlockCfg struct {
	ProduceDuration int32    `json:"ProduceDuration"`
	BlockInterval   int32    `json:"BlockInterval"`
}

type commonResult struct {
	RetCode int32             `json:"code"`
	RetMsg  string            `json:"msg"`
	Data    []nodeInfo        `json:"data"`
}

type nodeInfo struct {
	Name string          `json:"name,omitempty"`
	Owner string         `json:"owner,omitempty"`
	Desc string          `json:"desc,omitempty"`
	Types int32          `json:"type,omitempty"`
	Status int32         `json:"status,omitempty"`
	ExternalIP string    `json:"externalIP,omitempty"`
	InternalIP string    `json:"internalIP,omitempty"`
	PublicKey string     `json:"publicKey,omitempty"`
	RpcPort int32        `json:"rpcPort,omitempty"`
	P2pPort int32        `json:"p2pPort,omitempty"`
}

// getInitialNodesList catch initial nodes List from paramManager contract when
// new a dpos and miner a new block
func getConsensusNodesList() ([]discover.NodeID, error) {
	/*
	// get paramMgr contract address first
	callParams := []interface{}{"__sys_NodeManager", "latest"}
	cnsContractAddr := common.HexToAddress("0x0000000000000000000000000000000000000011")
	btsRes := common.InnerCall(cnsContractAddr, "getContractAddress", callParams)
	strRes := common.CallResAsString(btsRes)
	if len(strRes) == 0 || common.IsHexZeroAddress(strRes) {
		// log.Debug("system contract not found", "name", "__sys_NodeManager")
		return nil ,ErrContractNotFound
	}

	// get consensus node list
	nodeMgrContractAddr := common.HexToAddress(strRes)
	callParams = []interface{}{"{\"type\":1,\"status\":1}"}
	btsRes = common.InnerCall(nodeMgrContractAddr, "getNodes", callParams)
	strRes = common.CallResAsString(btsRes)
	log.Debug("get nodes info", "node", strRes)

	var tmp commonResult
	if err := json.Unmarshal(utils.String2bytes(strRes), &tmp); err != nil {
		log.Warn("unmarshal consensus node list failed", "result", strRes, "err", err.Error())
		return nil, err
	}
	if tmp.RetCode != 0 {
		log.Debug("contract inner error", "code", tmp.RetCode, "msg", tmp.RetMsg)
		return nil, errors.New(tmp.RetMsg)
	}
    */

	var tmp []common.NodeInfo
    if common.SysCfg != nil{
		tmp = common.SysCfg.GetConsensusNodes()
	}

	nodeIDs := make([]discover.NodeID, 0, len(tmp))
	for _, dataObj := range tmp {
		if pubKey := dataObj.PublicKey; len(pubKey) > 0 {
			log.Debug("consensus node id", "pubkey", pubKey)
			if nodeID, err := discover.HexID(pubKey); err == nil {
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
	}

	return nodeIDs, nil
}


// getCBFTParams catch cbft params config when miner a new block
func getCBFTConfigParams(cfg *params.CbftConfig) error {
	// get paramMgr contract address first
	/*
	callParams := []interface{}{"__sys_ParamManager", "latest"}
	cnsAddr := common.HexToAddress("0x0000000000000000000000000000000000000011")
	btsRes := common.InnerCall(cnsAddr, "getContractAddress", callParams)
	strRes := common.CallResAsString(btsRes)
	if common.IsHexZeroAddress(strRes) {
		// log.Debug("system contract not found", "name", "__sys_ParamManager")
		return ErrContractNotFound
	}

	// get cbft params
	paramContractAddr := common.HexToAddress(strRes)
	btsRes = common.InnerCall(paramContractAddr, "getCBFTTimeParam", []interface{}{})
	strRes = common.CallResAsString(btsRes)

	var tmp CBFTProduceBlockCfg
	if err := json.Unmarshal([]byte(strRes), &tmp); err != nil {
		log.Error("contract return invalid data", "result", strRes, "err", err.Error())
		return err
	}
*/

	tmp := common.SysCfg.GetCBFTTime()

	cfg.Duration = int64(tmp.ProduceDuration)
	cfg.Period = uint64(tmp.BlockInterval)

	// log.Debug("get cbft config params", "dur", cfg.Duration, "per", cfg.Period)
	return nil
}