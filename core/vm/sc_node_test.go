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
			if err := CheckNodeDescLen(tt.args.desc); (err != nil) != tt.wantErr {
				t.Errorf("CheckNodeDescLen() error = %v, wantErr %v", err, tt.wantErr)
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
			if err := CheckNodeNameLen(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("CheckNodeNameLen() error = %v, wantErr %v", err, tt.wantErr)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckRequiredFieldsIsEmpty(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("CheckRequiredFieldsIsEmpty() error = %v, wantErr %v", err, tt.wantErr)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			en := &ENode{
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
		want []*ENode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromNodes(tt.args.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromNodes() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenNodeName(tt.args.name); got != tt.want {
				t.Errorf("GenNodeName() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAddNodePermission(tt.args.caller); got != tt.want {
				t.Errorf("HasAddNodePermission() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUser(tt.args.caller); got != tt.want {
				t.Errorf("IsValidUser() = %v, want %v", got, tt.want)
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
			if err := n.Add(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
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
			if err := n.AddName(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("AddName() error = %v, wantErr %v", err, tt.wantErr)
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
			if err := n.CheckParamsOfAddNode(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("CheckParamsOfAddNode() error = %v, wantErr %v", err, tt.wantErr)
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
		update *UpdateNode
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
			got, err := n.CheckParamsOfUpdateNodeAndReturnUpdatedNode(tt.args.name, tt.args.update)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckParamsOfUpdateNodeAndReturnUpdatedNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckParamsOfUpdateNodeAndReturnUpdatedNode() got = %v, want %v", got, tt.want)
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
			if err := n.CheckPermissionForAdd(); (err != nil) != tt.wantErr {
				t.Errorf("CheckPermissionForAdd() error = %v, wantErr %v", err, tt.wantErr)
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
			if err := n.CheckPublicKeyExist(tt.args.pub); (err != nil) != tt.wantErr {
				t.Errorf("CheckPublicKeyExist() error = %v, wantErr %v", err, tt.wantErr)
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
		want    []*ENode
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
			got, err := n.GetENodesOfAllDeletedNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetENodesOfAllDeletedNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetENodesOfAllDeletedNodes() got = %v, want %v", got, tt.want)
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
		want    []*ENode
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
			got, err := n.GetENodesOfAllNormalNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetENodesOfAllNormalNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetENodesOfAllNormalNodes() got = %v, want %v", got, tt.want)
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
			got, err := n.GetNames()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNames() got = %v, want %v", got, tt.want)
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
			got, err := n.GetNodeByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNodeByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodeByName() got = %v, want %v", got, tt.want)
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
			if got := n.GetState(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetState() = %v, want %v", got, tt.want)
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
			if got := n.IsNameExist(tt.args.names, tt.args.name); got != tt.want {
				t.Errorf("IsNameExist() = %v, want %v", got, tt.want)
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
			got, err := n.NodesNum(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodesNum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NodesNum() got = %v, want %v", got, tt.want)
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
		update *UpdateNode
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
			if err := n.Update(tt.args.name, tt.args.update); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
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
