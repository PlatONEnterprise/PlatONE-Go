package vm

import (
	"encoding/hex"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math/rand"

	"github.com/stretchr/testify/assert"
	"math/big"
	"reflect"
	"testing"
)

func TestSCNode_isMatch(t *testing.T) {
	scNode := NewSCNode(nil)
	node := &syscontracts.NodeInfo{}
	query := &syscontracts.NodeInfo{}
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
	input := []*syscontracts.NodeInfo{&syscontracts.NodeInfo{PublicKey: "123", InternalIP: "127.0.0.1", P2pPort: 8888}}
	want := []*eNode{&eNode{"123", "127.0.0.1", 8888}}
	type args struct {
		nodes []*syscontracts.NodeInfo
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

func randFakeNodeInfo() *syscontracts.NodeInfo {
	ni := &syscontracts.NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = fmt.Sprintf("name-%d", rand.Int())
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = genPublicKeyInHex()
	fmt.Printf("rand fake node info:%+v\n", ni)
	return ni
}

func fakeNodeInfo() *syscontracts.NodeInfo {
	ni := &syscontracts.NodeInfo{}
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

func TestSCNode_TxReceipt(t *testing.T) {
	ni := &syscontracts.NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	stateDB := newMockStateDB()
	n := NewSCNode(stateDB)

	err := n.add(ni)
	assert.NoError(t, err)

	topic := "Notify"
	code := addNodeSuccess

	topicH := common.BytesToHash(crypto.Keccak256([]byte(topic)))
	log := stateDB.eLogs[topicH.String()]
	assert.NotEqual(t, nil, log)

	var data []rlp.RawValue
	err = rlp.DecodeBytes(log.Data, &data)
	assert.NoError(t, err)

	var code2 uint64
	err = rlp.DecodeBytes(data[0], &code2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(code), code2)

	var msg2 string
	err = rlp.DecodeBytes(data[1], &msg2)
	t.Log(msg2)
	assert.Regexp(t, "success", msg2)
}

func TestSCNode_Add(t *testing.T) {
	errNi := &syscontracts.NodeInfo{}
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
		node *syscontracts.NodeInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"t1", fields{stateDB: newMockStateDB()}, args{fakeNodeInfo()}, false},
		{"t1", fields{stateDB: newMockStateDB()}, args{errNi}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB:         tt.fields.stateDB,
				contractAddress: tt.fields.address,
				caller:          tt.fields.caller,
				blockNumber:     big.NewInt(0),
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
	stateDB := newMockStateDB()
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
				stateDB:         tt.fields.stateDB,
				contractAddress: tt.fields.address,
				caller:          tt.fields.caller,
			}
			if err := n.addName(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("addName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSCNode_CheckParamsOfAddNode(t *testing.T) {
	ni := &syscontracts.NodeInfo{}
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
		node *syscontracts.NodeInfo
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
				stateDB:         tt.fields.stateDB,
				contractAddress: tt.fields.address,
				caller:          tt.fields.caller,
			}
			if err := n.checkParamsOfAddNode(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("checkParamsOfAddNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func addNodeInfoIntoDB() (*SCNode, *syscontracts.NodeInfo) {
	ni := &syscontracts.NodeInfo{}
	ni.P2pPort = 8888
	ni.InternalIP = "127.0.0.1"
	ni.Name = "万向区块链"
	ni.Typ = NodeTypeObserver
	ni.Status = NodeStatusNormal
	ni.PublicKey = "044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc"

	db := newMockStateDB()
	n := NewSCNode(db)

	return n, ni
}

func TestSCNode_CheckParamsOfUpdateNodeAndReturnUpdatedNode(t *testing.T) {
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	update := &syscontracts.UpdateNode{}
	update.SetStatus(5)
	update.SetTyp(NodeTypeValidator)

	_, err = n.checkParamsOfUpdateNodeAndReturnUpdatedNode(ni.Name, update)
	assert.Error(t, err)

	update.SetStatus(NodeStatusDeleted)
	update.SetTyp(NodeTypeValidator)
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
	n := NewSCNode(db)

	err := n.checkPublicKeyExist("044b5378266d543212f1ebbea753ab98c26826d0f0fae86b2a5dabce563488a6569226228840ba02a606a003b9c708562906360478803dd6f3d446c54c79987fcc")
	assert.NoError(t, err)

	ni := &syscontracts.NodeInfo{}
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
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	allNodes, err := n.GetAllNodes()
	assert.NoError(t, err)
	assert.Equal(t, []*syscontracts.NodeInfo{ni}, allNodes)
}

func TestSCNode_GetENodesOfAllDeletedNodes(t *testing.T) {
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	enodes, err := n.getENodesOfAllDeletedNodes()
	assert.Error(t, err)

	update := &syscontracts.UpdateNode{}
	update.SetStatus(NodeStatusDeleted)
	err = n.update(ni.Name, update)
	assert.NoError(t, err)

	enodes, err = n.getENodesOfAllDeletedNodes()
	assert.NoError(t, err)
	ni.Status = NodeStatusDeleted
	assert.Equal(t, fromNodes([]*syscontracts.NodeInfo{ni}), enodes)
}

func TestSCNode_GetENodesOfAllNormalNodes(t *testing.T) {
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	enodes, err := n.getENodesOfAllNormalNodes()
	assert.NoError(t, err)
	assert.Equal(t, fromNodes([]*syscontracts.NodeInfo{ni}), enodes)
}

func TestSCNode_GetNames(t *testing.T) {
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	names, err := n.getNames()
	assert.NoError(t, err)
	assert.Equal(t, []string{ni.Name}, names)
}

func TestSCNode_GetNodeByName(t *testing.T) {
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	node, err := n.getNodeByName(ni.Name)
	assert.NoError(t, err)
	assert.Equal(t, ni, node)
}

func TestSCNode_GetNodes(t *testing.T) {
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	query := &syscontracts.NodeInfo{}
	query.Typ = NodeTypeObserver
	query.Status = NodeStatusDeleted

	_, err = n.GetNodes(query)
	assert.Error(t, err)

	query.Status = NodeStatusNormal
	node, err := n.GetNodes(query)
	assert.NoError(t, err)
	assert.Equal(t, []*syscontracts.NodeInfo{ni}, node)
}

func TestSCNode_IsNameExist(t *testing.T) {
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	ni2 := &syscontracts.NodeInfo{}
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
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	query := &syscontracts.NodeInfo{}
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
	n, ni := addNodeInfoIntoDB()
	err := n.add(ni)
	assert.NoError(t, err)

	update := &syscontracts.UpdateNode{}
	update.SetStatus(NodeStatusDeleted)
	err = n.update(ni.Name, update)
	assert.NoError(t, err)

	node, err := n.getNodeByName(ni.Name)
	assert.NoError(t, err)
	ni.Status = NodeStatusDeleted
	assert.Equal(t, ni, node)
}

func TestSCNode_isNameExist(t *testing.T) {
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
		{"t1", fields{}, args{[]string{"万向区块链"}, "wxblockchain"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SCNode{
				stateDB:         tt.fields.stateDB,
				contractAddress: tt.fields.address,
				caller:          tt.fields.caller,
			}
			if got := n.isNameExist(tt.args.names, tt.args.name); got != tt.want {
				t.Errorf("isNameExist() = %v, want %v", got, tt.want)
			}
		})
	}
}
