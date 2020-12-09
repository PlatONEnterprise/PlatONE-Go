package common

import (
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"math/big"
	"sync"
)

var (
	sysContractCall func(sc *SystemConfig) = nil

	SystemContractList = []string{
		"__sys_NodeManager",
		"__sys_UserManager",
		"__sys_ParamManager",
		"__sys_CAManager"}
)

func SetSysContractCallFunc(f func(*SystemConfig)) {
	sysContractCall = f
}

type CommonResult struct {
	RetCode int32      `json:"code"`
	RetMsg  string     `json:"msg"`
	Data    []NodeInfo `json:"data"`
}

type CAResult struct {
	RetCode int32      `json:"code"`
	RetMsg  string     `json:"msg"`
	Data    []string `json:"data"`
}

type NodeInfo struct {
	Name  string `json:"name,omitempty"`
	Owner string `json:"owner,omitempty"`
	Desc  string `json:"desc,omitempty"`
	Types int32  `json:"type,omitempty"`
	// status 1为正常节点, 2为删除节点
	Status     int32  `json:"status,omitempty"`
	ExternalIP string `json:"externalIP,omitempty"`
	InternalIP string `json:"internalIP,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
	RpcPort    int32  `json:"rpcPort,omitempty"`
	P2pPort    int32  `json:"p2pPort,omitempty"`
	// delay set validatorSet
	DelayNum uint64 `json:"delayNum,omitempty"`
}

type SystemParameter struct {
	BlockGasLimit                 int64
	TxGasLimit                    int64
	GasContractName               string
	GasContractAddr               Address
	CheckContractDeployPermission int64
	IsTxUseGas                    bool
	IsProduceEmptyBlock           bool
}

type SystemConfig struct {
	SystemConfigMu  *sync.RWMutex
	SysParam        *SystemParameter
	Nodes           []NodeInfo
	nodeMap         map[string]*NodeInfo
	ConsensusNodes  []*NodeInfo
	DeleteNodes     []*NodeInfo
	HighsetNumber   *big.Int
	ContractAddress map[string]Address
	CaMap           map[string]*gmssl.Certificate
}

var SysCfg = &SystemConfig{
	SystemConfigMu: &sync.RWMutex{},
	Nodes:          make([]NodeInfo, 0),
	nodeMap:        make(map[string]*NodeInfo),
	ConsensusNodes: make([]*NodeInfo, 0),
	DeleteNodes:    make([]*NodeInfo, 0),
	HighsetNumber:  new(big.Int).SetInt64(0),
	SysParam: &SystemParameter{
		BlockGasLimit: 0xffffffffffff,
		TxGasLimit:    100000000000000,
	},
	ContractAddress: make(map[string]Address),
	CaMap:         make(map[string]*gmssl.Certificate),

}

func InitSystemconfig(root NodeInfo) {
	SysCfg = &SystemConfig{
		SystemConfigMu: &sync.RWMutex{},
		Nodes:          make([]NodeInfo, 0),
		HighsetNumber:  new(big.Int).SetInt64(0),
		SysParam: &SystemParameter{
			BlockGasLimit: 0xffffffffffff,
			TxGasLimit:    10000000000000,
		},
		ContractAddress: make(map[string]Address),
		CaMap:         make(map[string]*gmssl.Certificate),
	}
	if root.Types == 1 {
		SysCfg.Nodes = append(SysCfg.Nodes, root)
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

func (sc *SystemConfig) IsProduceEmptyBlock() bool {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.SysParam.IsProduceEmptyBlock
}

func (sc *SystemConfig) IfCheckContractDeployPermission() int64 {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.SysParam.CheckContractDeployPermission
}

func (sc *SystemConfig) GetIsTxUseGas() bool {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.SysParam.IsTxUseGas
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
	return sc.SysParam.BlockGasLimit
}

func (sc *SystemConfig) GetTxGasLimit() int64 {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	if sc.SysParam.TxGasLimit > sc.SysParam.BlockGasLimit {
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

func (sc *SystemConfig) GetNormalNodes() []NodeInfo {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	var normalNodes = make([]NodeInfo, 0)

	for _, node := range sc.Nodes {
		if node.Status == 1 {
			normalNodes = append(normalNodes, node)
		}
	}
	return normalNodes
}

func (sc *SystemConfig) IsCaExistBySubject(subject string) bool{
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	if _, ok := sc.CaMap[subject]; ok {
		return true
	}
	return false
}

func (sc *SystemConfig) GetCaBySubject(subject string) *gmssl.Certificate{
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	if ca, ok := sc.CaMap[subject]; ok {
		return ca
	}
	return nil
}

func (sc *SystemConfig) IsValidJoinNode(publicKey string) bool {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	if node, ok := sc.nodeMap[publicKey]; ok {
		return node.Status == 1
	}
	return false
}

func (sc *SystemConfig) GetConsensusNodes() []*NodeInfo {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	return sc.ConsensusNodes
}

func (sc *SystemConfig) GetConsensusNodesFilterDelay(number uint64, nodes []NodeInfo, isOldBlock bool) []NodeInfo {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	var nodesInfos []NodeInfo
	if isOldBlock {
		nodesInfos = nodes
	} else {
		nodesInfos = sc.Nodes
	}

	consensusNodes := make([]NodeInfo, 0)
	for _, node := range nodesInfos {
		if node.Status == 1 && node.Types == 1 && node.DelayNum <= number {
			consensusNodes = append(consensusNodes, node)
		}
	}

	return consensusNodes
}

func (sc *SystemConfig) GetDeletedNodes() []*NodeInfo {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	return sc.DeleteNodes
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

func (sc *SystemConfig) GetGasContractAddress() Address {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()
	return sc.SysParam.GasContractAddr
}

func (sc *SystemConfig) GenerateNodeData() {
	sc.nodeMap = make(map[string]*NodeInfo)
	sc.ConsensusNodes = make([]*NodeInfo, 0)
	sc.DeleteNodes = make([]*NodeInfo, 0)
	for i, node := range sc.Nodes {
		sc.nodeMap[node.PublicKey] = &sc.Nodes[i]
		if node.Status != 1 {
			sc.DeleteNodes = append(sc.DeleteNodes, &sc.Nodes[i])
		} else if node.Types == 1 {
			sc.ConsensusNodes = append(sc.ConsensusNodes, &sc.Nodes[i])
		}
	}
}

func (sc *SystemConfig) GetNodeTypes(publicKey string) int32 {
	sc.SystemConfigMu.RLock()
	defer sc.SystemConfigMu.RUnlock()

	if node, ok := sc.nodeMap[publicKey]; ok {
		return node.Types
	}
	return 0
}
