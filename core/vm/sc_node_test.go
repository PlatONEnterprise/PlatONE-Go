package vm

import (
	"encoding/hex"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/stretchr/testify/assert"
	"math/big"
	"reflect"
	"testing"
)

type stateDBMock struct {
	database map[string]interface{}
}

func newStateDBMock() *stateDBMock {
	return &stateDBMock{database: make(map[string]interface{})}
}

func (m stateDBMock) CreateAccount(address common.Address) {
	panic("implement me")
}

func (m stateDBMock) SubBalance(address common.Address, b *big.Int) {
	panic("implement me")
}

func (m stateDBMock) AddBalance(address common.Address, b *big.Int) {
	panic("implement me")
}

func (m stateDBMock) GetBalance(address common.Address) *big.Int {
	panic("implement me")
}

func (m stateDBMock) GetNonce(address common.Address) uint64 {
	panic("implement me")
}

func (m stateDBMock) SetNonce(address common.Address, u uint64) {
	panic("implement me")
}

func (m stateDBMock) GetCodeHash(address common.Address) common.Hash {
	panic("implement me")
}

func (m stateDBMock) GetCode(address common.Address) []byte {
	panic("implement me")
}

func (m stateDBMock) SetCode(address common.Address, bytes []byte) {
	panic("implement me")
}

func (m stateDBMock) GetCodeSize(address common.Address) int {
	panic("implement me")
}

func (m stateDBMock) GetAbiHash(address common.Address) common.Hash {
	panic("implement me")
}

func (m stateDBMock) GetAbi(address common.Address) []byte {
	panic("implement me")
}

func (m stateDBMock) SetAbi(address common.Address, bytes []byte) {
	panic("implement me")
}

func (m stateDBMock) AddRefund(u uint64) {
	panic("implement me")
}

func (m stateDBMock) SubRefund(u uint64) {
	panic("implement me")
}

func (m stateDBMock) GetRefund() uint64 {
	panic("implement me")
}

func (m stateDBMock) GetCommittedState(address common.Address, bytes []byte) []byte {
	panic("implement me")
}

func (m stateDBMock) GetState(address common.Address, key []byte) []byte {
	if bin, ok := m.database[string(key)]; ok {
		return bin.([]byte)
	}

	return nil
}

func (m stateDBMock) SetState(address common.Address, key []byte, val []byte) {
	m.database[string(key)] = val
}

func (m stateDBMock) Suicide(address common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) HasSuicided(address common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) Exist(address common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) Empty(address common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) RevertToSnapshot(i int) {
	panic("implement me")
}

func (m stateDBMock) Snapshot() int {
	panic("implement me")
}

func (m stateDBMock) AddLog(log *types.Log) {
	panic("implement me")
}

func (m stateDBMock) AddPreimage(hash common.Hash, bytes []byte) {
	panic("implement me")
}

func (m stateDBMock) ForEachStorage(address common.Address, f func(common.Hash, common.Hash) bool) {
	panic("implement me")
}

func (m stateDBMock) FwAdd(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m stateDBMock) FwClear(contractAddr common.Address, action state.Action) {
	panic("implement me")
}

func (m stateDBMock) FwDel(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m stateDBMock) FwSet(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m stateDBMock) SetFwStatus(contractAddr common.Address, status state.FwStatus) {
	panic("implement me")
}

func (m stateDBMock) GetFwStatus(contractAddr common.Address) state.FwStatus {
	panic("implement me")
}

func (m stateDBMock) SetContractCreator(contractAddr common.Address, creator common.Address) {
	panic("implement me")
}

func (m stateDBMock) GetContractCreator(contractAddr common.Address) common.Address {
	panic("implement me")
}

func (m stateDBMock) OpenFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m stateDBMock) CloseFirewall(contractAddr common.Address) {

}

func (m stateDBMock) IsFwOpened(contractAddr common.Address) bool {
	panic("implement me")
}

func (m stateDBMock) FwImport(contractAddr common.Address, data []byte) error {
	panic("implement me")
}

func TestSCNode_isMatch(t *testing.T) {
	scNode := NewSCNode()
	node := &NodeInfo{}
	query := &NodeInfo{}
	assert.Equal(t, true, scNode.isMatch(node, query))
	node.Name = "elvin"
	assert.Equal(t, true, scNode.isMatch(node, query))
	query.PublicKey = "aaaaaa"
	assert.Equal(t, false, scNode.isMatch(node, query))

	node.PublicKey = "aaaaaa"
	assert.Equal(t, true, scNode.isMatch(node, query))
}

func TestCheckNodeDescLen(t *testing.T) {
	type args struct {
		desc string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"t1", args{`上海万向区块链有限公司`}, false},
		{"t2", args{`shanghai wanxiang`}, false},
		{"t3", args{`01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890`}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkNodeDescLen(tt.args.desc); (err != nil) != tt.wantErr {
				t.Errorf("checkNodeDescLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckNodeNameLen(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"t1", args{`上海万向区块链有限公司`}, false},
		{"t1", args{`上海万向区块链有限公司-上海万向区块链有限公司-上海万向区块链有限公司-上海万向区块链有限公司-`}, false},
		{"t2", args{`shanghai wanxiang`}, false},
		{"t3", args{`012345678901234567890123456789012345678901234567890`}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkNodeNameLen(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("checkNodeNameLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestENode_String(t *testing.T) {
	type fields struct {
		PublicKey string
		IP        string
		Port      int32
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"t1", fields{"123", "127.0.0.1", 8987}, "enode://123@127.0.0.1:8987"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			en := &eNode{
				PublicKey: tt.fields.PublicKey,
				IP:        tt.fields.IP,
				Port:      tt.fields.Port,
			}
			if got := en.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromNodes(t *testing.T) {
	input := []*NodeInfo{&NodeInfo{PublicKey: "123", InternalIP: "127.0.0.1", P2pPort: 8888}}
	want := []*eNode{&eNode{"123", "127.0.0.1", 8888}}
	type args struct {
		nodes []*NodeInfo
	}
	tests := []struct {
		name string
		args args
		want []*eNode
	}{
		{"t1", args{input}, want},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromNodes(tt.args.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenNodeName(t *testing.T) {
	name := "万向区块链"
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"t1", args{name}, prefixNodeName + "-" + name},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genNodeName(tt.args.name); got != tt.want {
				t.Errorf("genNodeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAddNodePermission(t *testing.T) {
	type args struct {
		caller common.Address
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasAddNodePermission(tt.args.caller); got != tt.want {
				t.Errorf("hasAddNodePermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidUser(t *testing.T) {
	type args struct {
		caller common.Address
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidUser(tt.args.caller); got != tt.want {
				t.Errorf("isValidUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fakeNodeInfo() *NodeInfo {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	//ni.PublicKey = "0x294866ff9693257147c7AE69293609F4b6E59Aa1"
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"
	return ni
}

func genPublicKeyInHex() string {
	prk, _ := crypto.GenerateKey()
	pub := crypto.FromECDSAPub(&prk.PublicKey)
	//fmt.Println(hex.EncodeToString(pub))

	return hex.EncodeToString(pub)
}

func TestSCNode_Add(t *testing.T) {
	errNi := &NodeInfo{}
	errNi.P2pPort = 8888
	errNi.InternalIP = "127.0.0.1"
	errNi.Name = "万向区块链"
	errNi.Typ = NodeTypeObserver
	errNi.Status = NodeStatusNormal
	errNi.PublicKey = "0x294866ff9693257147c7AE69293609F4b6E59Aa1"

	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		node *NodeInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"t1", fields{stateDB: newStateDBMock()}, args{fakeNodeInfo()}, false},
		{"t1", fields{stateDB: newStateDBMock()}, args{errNi}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if err := n.add(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSCNode_AddName(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		name string
	}
	stateDB := newStateDBMock()
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"t1", fields{stateDB: stateDB}, args{"上海万向区块链"}, false},
		{"t2", fields{stateDB: stateDB}, args{"上海万向区块链"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if err := n.addName(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("addName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSCNode_CheckParamsOfAddNode(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		node *NodeInfo
	}
	db := newMockStateDB()
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"t1", fields{stateDB: db}, args{ni}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if err := n.checkParamsOfAddNode(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("checkParamsOfAddNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSCNode_CheckParamsOfUpdateNodeAndReturnUpdatedNode(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	update := &updateNode{}
	update.setStatus(5)
	update.setTyp(NodeTypeValidator)

	_, err = n.checkParamsOfUpdateNodeAndReturnUpdatedNode(ni.Name, update)
	assert.Error(t, err)

	update.setStatus(NodeStatusDeleted)
	update.setTyp(NodeTypeValidator)
	desc := "上海万向区块链是一个伟大的企业"
	update.Desc = &desc
	updatedNI, err := n.checkParamsOfUpdateNodeAndReturnUpdatedNode(ni.Name, update)
	assert.NoError(t, err)
	ni.Desc = desc
	ni.Status = NodeStatusDeleted
	ni.Typ = NodeTypeValidator
	assert.Equal(t, ni, updatedNI)
}

func TestSCNode_CheckPublicKeyExist(t *testing.T) {
	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.checkPublicKeyExist("044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc")
	assert.NoError(t, err)

	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	err = n.add(ni)
	assert.NoError(t, err)

	err = n.checkPublicKeyExist("044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc")
	assert.Error(t, err)
}

func TestSCNode_GetAllNodes(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	allNodes, err := n.GetAllNodes()
	assert.NoError(t, err)
	assert.Equal(t, []*NodeInfo{ni}, allNodes)
}

func TestSCNode_GetENodesOfAllDeletedNodes(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	enodes, err := n.getENodesOfAllDeletedNodes()
	assert.Error(t, err)

	update := &updateNode{}
	update.setStatus(NodeStatusDeleted)
	err = n.update(ni.Name, update)
	assert.NoError(t, err)

	enodes, err = n.getENodesOfAllDeletedNodes()
	assert.NoError(t, err)
	ni.Status = NodeStatusDeleted
	assert.Equal(t, fromNodes([]*NodeInfo{ni}), enodes)
}

func TestSCNode_GetENodesOfAllNormalNodes(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	enodes, err := n.getENodesOfAllNormalNodes()
	assert.NoError(t, err)
	assert.Equal(t, fromNodes([]*NodeInfo{ni}), enodes)
}

func TestSCNode_GetNames(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	names, err := n.getNames()
	assert.NoError(t, err)
	assert.Equal(t, []string{ni.Name}, names)
}

func TestSCNode_GetNodeByName(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	node, err := n.getNodeByName(ni.Name)
	assert.NoError(t, err)
	assert.Equal(t, ni, node)
}

func TestSCNode_GetNodes(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	query := &NodeInfo{}
	query.Typ = NodeTypeObserver
	query.Status = NodeStatusDeleted

	_, err = n.GetNodes(query)
	assert.Error(t, err)

	query.Status = NodeStatusNormal
	node, err := n.GetNodes(query)
	assert.NoError(t, err)
	assert.Equal(t, []*NodeInfo{ni}, node)
}

func TestSCNode_IsNameExist(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	ni2 := &NodeInfo{}
	ni2.P2pPort = 8888
	ni2.InternalIP = "127.0.0.1"
	ni2.Name = "通联支付"
	ni2.Typ = NodeTypeObserver
	ni2.Status = NodeStatusNormal
	ni2.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	err = n.add(ni2)
	assert.Error(t, err)

	ni2.PublicKey = genPublicKeyInHex()
	err = n.add(ni2)
	assert.NoError(t, err)

	names, err := n.getNames()
	assert.NoError(t, err)

	exist := n.isNameExist(names, ni.Name)
	assert.Equal(t, true, exist)
}

func TestSCNode_NodesNum(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	query := &NodeInfo{}
	query.Name = "万向"
	num, err := n.nodesNum(query)
	assert.NoError(t, err)
	assert.Equal(t, 0, num)

	query.Name = ni.Name
	num, err = n.nodesNum(query)
	assert.NoError(t, err)
	assert.Equal(t, 1, num)
}

func TestSCNode_Update(t *testing.T) {
	ni := &NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode()
	n.stateDB = db

	err := n.add(ni)
	assert.NoError(t, err)

	update := &updateNode{}
	update.setStatus(NodeStatusDeleted)
	err = n.update(ni.Name, update)
	assert.NoError(t, err)

	node, err := n.getNodeByName(ni.Name)
	assert.NoError(t, err)
	ni.Status = NodeStatusDeleted
	assert.Equal(t, ni, node)
}
