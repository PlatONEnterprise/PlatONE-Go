package syscontracts

import (
	"encoding/json"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

// the system contract addr  table
var (
	UserManagementAddress        = common.HexToAddress("0x1000000000000000000000000000000000000001") // The PlatONE Precompiled contract addr for user management
	NodeManagementAddress        = common.HexToAddress("0x1000000000000000000000000000000000000002") // The PlatONE Precompiled contract addr for node management
	CnsManagementAddress         = common.HexToAddress("0x0000000000000000000000000000000000000011") // The PlatONE Precompiled contract addr for CNS
	ParameterManagementAddress   = common.HexToAddress("0x1000000000000000000000000000000000000004") // The PlatONE Precompiled contract addr for parameter management
	FirewallManagementAddress    = common.HexToAddress("0x1000000000000000000000000000000000000005") // The PlatONE Precompiled contract addr for fire wall management
	GroupManagementAddress       = common.HexToAddress("0x1000000000000000000000000000000000000006") // The PlatONE Precompiled contract addr for group management
	ContractDataProcessorAddress = common.HexToAddress("0x1000000000000000000000000000000000000007") // The PlatONE Precompiled contract addr for group management
	CnsInvokeAddress             = common.HexToAddress("0x0000000000000000000000000000000000000000") // The PlatONE Precompiled contract addr for group management
)

type UpdateNode struct {
	Desc *string `json:"desc,omitempty"`
	Typ  *uint32 `json:"type,omitempty"` // 0:观察者节点；1:共识节点
	// status 1为正常节点, 2为删除节点
	Status *uint32 `json:"status,omitempty,required"`
	// delay set validatorSet
	DelayNum *uint64 `json:"delayNum,omitempty"` //共识节点延迟设置的区块高度 (可选, 默认实时设置)
}

func (un *UpdateNode) SetStatus(status uint32) {
	un.Status = &status
}

func (un *UpdateNode) SetTyp(typ uint32) {
	un.Typ = &typ
}

func (un *UpdateNode) String() string {
	str, _ := json.Marshal(un)
	return string(str)
}

type NodeInfo struct {
	Name  string `json:"name,omitempty,required"` //全网唯一，不能重复。所有接口均以此为主键。 这个名称意义是？
	Owner string `json:"owner"`
	Desc  string `json:"desc"`
	Typ   uint32 `json:"type"` // 0:观察者节点；1:共识节点
	// status 1为正常节点, 2为删除节点
	Status     uint32 `json:"status,required"`
	ExternalIP string `json:"externalIP,required"`
	InternalIP string `json:"internalIP,omitempty,required"`
	PublicKey  string `json:"publicKey,required"` //节点公钥，全网唯一，不能重复
	RpcPort    uint32 `json:"rpcPort"`
	P2pPort    uint32 `json:"p2pPort,required"`
	// delay set validatorSet
	DelayNum uint64 `json:"delayNum,omitempty"` //共识节点延迟设置的区块高度 (可选, 默认实时设置)
}

func (node *NodeInfo) String() string {
	str, _ := json.Marshal(node)
	return string(str)
}

type UserInfo struct {
	Address    common.Address `json:"address,string,omitempty,required"` // 地址，不可变更
	Authorizer common.Address `json:"authorizer,string,omitempty"`       // 授权者，不可变更
	Name       string         `json:"name,omitempty"`                    // 用户名，不可变更

	DescInfo string `json:"descInfo,omitempty"` // 描述信息，可变更
	Version  uint32 `json:"version,omitempty"`  // 可变更
}

type UserDescInfo struct {
	Email        string `json:"email,omitempty"`
	Organization string `json:"organization,omitempty"`
	Phone        string `json:"phone,omitempty"`
}
