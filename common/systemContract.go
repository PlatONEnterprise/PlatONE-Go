package common

import (
	"math/big"
	"sync"
)

var (
	sysContractCall func(sc *SystemConfig) = nil
)

func SetSysContractCallFunc(f func(*SystemConfig)) {
	sysContractCall = f
}

type CBFTProduceBlockCfg struct {
	ProduceDuration int32 `json:"ProduceDuration"`
	BlockInterval   int32 `json:"BlockInterval"`
}

type CommonResult struct {
	RetCode int32      `json:"code"`
	RetMsg  string     `json:"msg"`
	Data    []NodeInfo `json:"data"`
}

type NodeInfo struct {
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

type SystemParameter struct {
	BlockGasLimit   int64
	TxGasLimit      int64
	CBFTTime        CBFTProduceBlockCfg
	GasContractName string
	GasContractAddr Address
}

type SystemConfig struct {
	SystemConfigMu  *sync.RWMutex
	SysParam        *SystemParameter
	Nodes           []NodeInfo
	HighsetNumber   *big.Int
	ContractAddress map[string]Address
}

var SysCfg *SystemConfig


func InitSystemconfig() {
	SysCfg = &SystemConfig{
		SystemConfigMu: &sync.RWMutex{},
		Nodes:         make([]NodeInfo, 0),
		HighsetNumber: new(big.Int).SetInt64(0),
		SysParam: &SystemParameter{
			BlockGasLimit:0xffffffffffff,
			TxGasLimit: 10000000000000,
			CBFTTime: CBFTProduceBlockCfg{
				ProduceDuration: int32(10),
				BlockInterval:   int32(1),
			},
		},
		ContractAddress: make(map[string]Address),
	}
}

func (sc *SystemConfig) UpdateSystemConfig() {
	sc.SystemConfigMu.Lock()
	defer sc.SystemConfigMu.Unlock()

	if sysContractCall == nil {
		return
	}

	sysContractCall(sc)
}

func (sc *SystemConfig) GetBlockGasLimit() int64 {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	if sc.SysParam.BlockGasLimit < sc.SysParam.TxGasLimit {
		sc.SysParam.BlockGasLimit = sc.SysParam.TxGasLimit
	}
	if sc.SysParam.BlockGasLimit == 0 {
		return 0xffffffffffff
	}
	return sc.SysParam.TxGasLimit
}

func (sc *SystemConfig) GetTxGasLimit() int64 {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	if sc.SysParam.TxGasLimit >  sc.SysParam.BlockGasLimit{
		sc.SysParam.TxGasLimit = sc.SysParam.BlockGasLimit
	}

	if sc.SysParam.TxGasLimit == 0 {
		return 10000000000000
	}
	return sc.SysParam.TxGasLimit
}

func (sc *SystemConfig) GetHighsetNumber() *big.Int {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.HighsetNumber
}

func (sc *SystemConfig) GetCBFTTime() CBFTProduceBlockCfg {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.SysParam.CBFTTime
}

func (sc *SystemConfig) GetNormalNodes() []NodeInfo {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	var normalNodes =make([]NodeInfo,0)

	for _, node := range sc.Nodes {
		if node.Status <= 2 {
			normalNodes = append(normalNodes, node)
		}
	}
	return normalNodes
}

func (sc *SystemConfig) IsValidJoinNode(publicKey string) bool {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	var validNodes =make([]NodeInfo,0)

	for _, node := range sc.Nodes {
		if (node.Status == 1 || node.Status == 2) && node.PublicKey == publicKey {
			validNodes = append(validNodes, node)
		}
	}

	return len(validNodes) == 1
}

func (sc *SystemConfig) GetConsensusNodes() []NodeInfo {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	consensusNodes := make([]NodeInfo,0)

	for _, node := range sc.Nodes {
		if node.Status == 1 && node.Types == 1 {
			consensusNodes = append(consensusNodes, node)
		}
	}

	return consensusNodes
}

func (sc *SystemConfig) GetDeletedNodes() []NodeInfo {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	var deletedNodes  = make([]NodeInfo,0)

	for _, node := range sc.Nodes {
		if node.Status == 3 {
			deletedNodes = append(deletedNodes, node)
		}
	}
	return deletedNodes
}

func (sc *SystemConfig) GetContractAddress(name string) Address {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.ContractAddress[name]
}

func (sc *SystemConfig) GetGasContractName() string {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.SysParam.GasContractName
}

func (sc *SystemConfig) GetGasContractGas() Address {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.SysParam.GasContractAddr
}
