package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"reflect"
	"testing"
)

type stateDBMock struct {
	database map[string]interface{}
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
	return m.database[string(key)].([]byte)
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

func TestCheckRequiredFieldsIsEmpty(t *testing.T) {
	type args struct {
		node *NodeInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkRequiredFieldsIsEmpty(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("checkRequiredFieldsIsEmpty() error = %v, wantErr %v", err, tt.wantErr)
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
		// TODO: add test cases.
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
	type args struct {
		nodes []*NodeInfo
	}
	tests := []struct {
		name string
		args args
		want []*eNode
	}{
		// TODO: add test cases.
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
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: add test cases.
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

func TestNewSCNode(t *testing.T) {
	tests := []struct {
		name string
		want *SCNode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSCNode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSCNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_Add(t *testing.T) {
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
		// TODO: Add test cases.
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
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		name   string
		update *updateNode
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *NodeInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			got, err := n.checkParamsOfUpdateNodeAndReturnUpdatedNode(tt.args.name, tt.args.update)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkParamsOfUpdateNodeAndReturnUpdatedNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checkParamsOfUpdateNodeAndReturnUpdatedNode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_CheckPermissionForAdd(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if err := n.checkPermissionForAdd(); (err != nil) != tt.wantErr {
				t.Errorf("checkPermissionForAdd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSCNode_CheckPublicKeyExist(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		pub string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if err := n.checkPublicKeyExist(tt.args.pub); (err != nil) != tt.wantErr {
				t.Errorf("checkPublicKeyExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSCNode_GetAllNodes(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*NodeInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			got, err := n.GetAllNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_GetENodesOfAllDeletedNodes(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*eNode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			got, err := n.getENodesOfAllDeletedNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("getENodesOfAllDeletedNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getENodesOfAllDeletedNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_GetENodesOfAllNormalNodes(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*eNode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			got, err := n.getENodesOfAllNormalNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("getENodesOfAllNormalNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getENodesOfAllNormalNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_GetNames(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			got, err := n.getNames()
			if (err != nil) != tt.wantErr {
				t.Errorf("getNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNames() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_GetNodeByName(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *NodeInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			got, err := n.getNodeByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodeByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNodeByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_GetNodes(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		query *NodeInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*NodeInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			got, err := n.GetNodes(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_GetState(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if got := n.getState(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_IsNameExist(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		names []string
		name  string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if got := n.isNameExist(tt.args.names, tt.args.name); got != tt.want {
				t.Errorf("isNameExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_NodesNum(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		query *NodeInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			got, err := n.nodesNum(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("nodesNum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("nodesNum() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCNode_Update(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		name   string
		update *updateNode
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if err := n.update(tt.args.name, tt.args.update); (err != nil) != tt.wantErr {
				t.Errorf("update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSCNode_isMatch1(t *testing.T) {
	type fields struct {
		stateDB StateDB
		address common.Address
		caller  common.Address
	}
	type args struct {
		node  *NodeInfo
		query *NodeInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB: tt.fields.stateDB,
				address: tt.fields.address,
				caller:  tt.fields.caller,
			}
			if got := n.isMatch(tt.args.node, tt.args.query); got != tt.want {
				t.Errorf("isMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
