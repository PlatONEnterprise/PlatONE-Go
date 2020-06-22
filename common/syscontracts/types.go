package syscontracts

import "github.com/PlatONEnetwork/PlatONE-Go/common"

// the system contract addr  table
var (
	USER_MANAGEMENT_ADDRESS      = common.HexToAddress("0x1000000000000000000000000000000000000001") // The PlatONE Precompiled contract addr for user management
	NODE_MANAGEMENT_ADDRESS      = common.HexToAddress("0x1000000000000000000000000000000000000002") // The PlatONE Precompiled contract addr for node management
	CNS_MANAGEMENT_ADDRESS       = common.HexToAddress("0x1000000000000000000000000000000000000003") // The PlatONE Precompiled contract addr for CNS
	PARAMETER_MANAGEMENT_ADDRESS = common.HexToAddress("0x1000000000000000000000000000000000000004") // The PlatONE Precompiled contract addr for parameter management
)

type UpdateNode struct {
	Desc *string `json:"desc,omitempty"` //没有用到？删除？
	Typ  *int32  `json:"type,omitempty"` // 0:观察者节点；1:共识节点
	// status 1为正常节点, 2为删除节点
	Status *int32 `json:"status,omitempty,required"`
	// delay set validatorSet
	DelayNum *uint64 `json:"delayNum,omitempty"` //共识节点延迟设置的区块高度 (可选, 默认实时设置)
}

func (un *UpdateNode) SetStatus(status int32) {
	un.Status = &status
}

func (un *UpdateNode) SetTyp(typ int32) {
	un.Typ = &typ
}

type NodeInfo struct {
	Name  string `json:"name,omitempty,required"` //全网唯一，不能重复。所有接口均以此为主键。 这个名称意义是？
	Owner string `json:"owner,omitempty"`         //todo 没有用到？删除？
	Desc  string `json:"desc,omitempty"`          //todo 没有用到？删除？
	Typ   int32  `json:"type,omitempty"`          // 0:观察者节点；1:共识节点
	// status 1为正常节点, 2为删除节点
	Status     int32  `json:"status,omitempty,required"`
	ExternalIP string `json:"externalIP,omitempty"` //todo 没有用到？删除？
	InternalIP string `json:"internalIP,omitempty,required"`
	PublicKey  string `json:"publicKey,omitempty,required"` //节点公钥，全网唯一，不能重复
	RpcPort    int32  `json:"rpcPort,omitempty"`            //todo 没有用到？删除？
	P2pPort    int32  `json:"p2pPort,omitempty,required"`
	// delay set validatorSet
	DelayNum uint64 `json:"delayNum,omitempty"` //共识节点延迟设置的区块高度 (可选, 默认实时设置)
}
