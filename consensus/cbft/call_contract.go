package cbft

import (
	"encoding/json"
	"errors"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/life/utils"
	"github.com/PlatONnetwork/PlatON-Go/log"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/PlatON-Go/params"
)

var (
	ErrContractNotFound = errors.New("contract not found")
)

type CBFTProduceBlockCfg struct {
	ProduceDuration int32    `json:"ProduceDuration"`
	BlockInterval   int32    `json:"BlockInterval"`
}

type CommonResult struct {
	retCode int32        `json:"code"`
	retMsg  string       `json:"msg"`
	data    []string     `json:"data"`
}


// getInitialNodesList catch initial nodes List from paramManager contract when
// new a dpos and miner a new block
func getInitialNodesList() ([]discover.NodeID, error) {
	// get paramMgr contract address first
	callParams := []interface{}{"__sys_ParamManager", "latest"}
	cnsContractAddr := common.HexToAddress("0x0000000000000000000000000000000000000011")
	btsRes := common.InnerCall(cnsContractAddr, "getContractAddress", callParams)
	strRes := common.CallResAsString(btsRes)
	if common.IsHexZeroAddress(strRes) {
		log.Trace("system contract not found", "name", "__sys_ParamManager")
		return nil ,ErrContractNotFound
	}

	// get node list
	paramContractAddr := common.HexToAddress(strRes)
	btsRes = common.InnerCall(paramContractAddr, "getAllNodes", []interface{}{})
	strRes = common.CallResAsString(btsRes)

	var tmp *CommonResult
	if err := json.Unmarshal(utils.String2bytes(strRes), tmp); err != nil {
		//log.Error("", "result", strRes, "err", err.Error())
		return nil, err
	}

	if tmp.retCode != 0 {
		log.Debug("contract inner error", "code", tmp.retCode, "msg", tmp.retMsg)
		return nil, errors.New(tmp.retMsg)
	}

	nodeIDs := make([]discover.NodeID, 0, len(tmp.data))
	for _, nodeStr := range tmp.data {
		nodeID, err := discover.HexID(nodeStr)
		if err != nil {
			log.Debug("node string from data is valid", "info", nodeID)
			continue
		}
		nodeIDs = append(nodeIDs, nodeID)
	}

	return nodeIDs, nil
}


// getCBFTParams catch cbft params config when miner a new block
func getCBFTConfigParams(cfg *params.CbftConfig) error {
	// get paramMgr contract address first
	callParams := []interface{}{"__sys_ParamManager", "latest"}
	cnsAddr := common.HexToAddress("0x0000000000000000000000000000000000000011")
	btsRes := common.InnerCall(cnsAddr, "getContractAddress", callParams)
	strRes := common.CallResAsString(btsRes)
	if common.IsHexZeroAddress(strRes) {
		log.Trace("system contract not found", "name", "__sys_ParamManager")
		return ErrContractNotFound
	}

	// get cbft params
	paramContractAddr := common.HexToAddress(strRes)
	btsRes = common.InnerCall(paramContractAddr, "getCBFTTimeParam", []interface{}{})
	strRes = common.CallResAsString(btsRes)

	var tmp *CBFTProduceBlockCfg
	if err := json.Unmarshal(utils.String2bytes(strRes), tmp); err != nil {
		//log.Error("", "result", strRes, "err", err.Error())
		return err
	}

	cfg.Duration = int64(tmp.ProduceDuration)
	cfg.Period = uint64(tmp.BlockInterval)

	return nil
}