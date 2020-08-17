package vm

import (
	"encoding/json"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/stretchr/testify/assert"
)

func Test_scNodeWrapper_add(t *testing.T) {
	addNodeInfoForTest(t)

	mockDB := newMockStateDB()
	ni := randFakeNodeInfo()
	addRandNodeInfoForTest(t, ni, mockDB)

	ni = randFakeNodeInfo()
	_, node := addRandNodeInfoForTest(t, ni, mockDB)

	fnNameInput := "getAllNodes"
	var input = MakeInput(fnNameInput)
	ret, err := node.Run(input)
	assert.NoError(t, err)
	t.Log(string(ret))
}

func addRandNodeInfoForTest(t *testing.T, ni *syscontracts.NodeInfo, mockDB *mockStateDB) (*syscontracts.NodeInfo, *scNodeWrapper) {
	fnNameInput := "add"

	params, err := json.Marshal(ni)

	assert.NoError(t, err)
	var input = MakeInput(fnNameInput, string(params))

	node := &scNodeWrapper{NewSCNode(mockDB)}

	ret, err := node.Run(input)
	assert.NoError(t, err)
	assert.Equal(t, int64(addNodeSuccess), byteutil.BytesToInt64(ret))

	return ni, node
}

func addNodeInfoForTest(t *testing.T) (*syscontracts.NodeInfo, *scNodeWrapper) {
	fnNameInput := "add"
	ni := fakeNodeInfo()

	//d := genPublicKeyInHex()
	//ni.PublicKey = d
	//ni.Name = "wxblockchian"
	params, err := json.Marshal(ni)

	assert.NoError(t, err)
	var input = MakeInput(fnNameInput, string(params))
	mockDB := newMockStateDB()
	node := &scNodeWrapper{NewSCNode(mockDB)}

	ret, err := node.Run(input)
	assert.NoError(t, err)
	assert.Equal(t, int64(addNodeSuccess), byteutil.BytesToInt64(ret))

	return ni, node
}

func Test_scNodeWrapper_getAllNodes(t *testing.T) {
	ni, node := addNodeInfoForTest(t)

	fnNameInput := "getAllNodes"
	var input = MakeInput(fnNameInput)
	ret, err := node.Run(input)
	assert.NoError(t, err)

	result := newSuccessResult([]*syscontracts.NodeInfo{ni})
	bin, err := json.Marshal(result)
	assert.NoError(t, err)

	assert.Equal(t, toContractReturnValueStringType(E_INVOKE_CONTRACT, bin), ret)
}

func Test_scNodeWrapper_getENodesOfAllDeletedNodes(t *testing.T) {
	_, node := addNodeInfoForTest(t)

	fnNameInput := "getDeletedEnodeNodes"
	var input = MakeInput(fnNameInput)
	ret, err := node.Run(input)
	assert.NoError(t, err)

	assert.Equal(t, toContractReturnValueStringType(E_INVOKE_CONTRACT, []byte(`{"code":1,"msg":"node not found","data":[]}`)), ret)
}

func Test_scNodeWrapper_getENodesOfAllNormalNodes(t *testing.T) {
	ni, node := addNodeInfoForTest(t)

	fnNameInput := "getNormalEnodeNodes"
	var input = MakeInput(fnNameInput)
	ret, err := node.Run(input)
	assert.NoError(t, err)

	enode := &eNode{}
	enode.Port = ni.P2pPort
	enode.IP = ni.InternalIP
	enode.PublicKey = ni.PublicKey
	expected := newSuccessResult([]*eNode{enode}).String()
	assert.Equal(t, toContractReturnValueStringType(E_INVOKE_CONTRACT, []byte(expected)), ret)
}

func Test_scNodeWrapper_getNodes(t *testing.T) {
	ni, node := addNodeInfoForTest(t)

	query := syscontracts.NodeInfo{}
	query.PublicKey = ni.PublicKey

	query2, err := json.Marshal(query)
	assert.NoError(t, err)

	fnNameInput := "getNodes"
	var input = MakeInput(fnNameInput, string(query2))
	ret, err := node.Run(input)
	assert.NoError(t, err)

	result := newSuccessResult([]*syscontracts.NodeInfo{ni})
	bin, err := json.Marshal(result)
	assert.NoError(t, err)

	assert.Equal(t, toContractReturnValueStringType(E_INVOKE_CONTRACT, bin), ret)
}

func Test_scNodeWrapper_isPublicKeyExist(t *testing.T) {
	ni, node := addNodeInfoForTest(t)

	fnNameInput := "validJoinNode"
	var input = MakeInput(fnNameInput, ni.PublicKey)
	ret, err := node.Run(input)
	assert.NoError(t, err)

	assert.Equal(t, toContractReturnValueIntType(E_INVOKE_CONTRACT, int64(publicKeyExist)), ret)
}

func Test_scNodeWrapper_nodesNum(t *testing.T) {
	ni, node := addNodeInfoForTest(t)

	query := syscontracts.NodeInfo{}
	query.PublicKey = ni.PublicKey

	query2, err := json.Marshal(query)
	assert.NoError(t, err)

	fnNameInput := "nodesNum"
	var input = MakeInput(fnNameInput, string(query2))
	ret, err := node.Run(input)
	assert.NoError(t, err)

	assert.Equal(t, toContractReturnValueIntType(E_INVOKE_CONTRACT, int64(1)), ret)
}

func Test_scNodeWrapper_update(t *testing.T) {
	ni, node := addNodeInfoForTest(t)

	update := syscontracts.UpdateNode{}
	update.SetStatus(NodeStatusDeleted)
	bin, err := json.Marshal(update)
	assert.NoError(t, err)

	fnNameInput := "update"
	var input = MakeInput(fnNameInput, ni.Name, string(bin))
	ret, err := node.Run(input)
	assert.NoError(t, err)
	assert.Equal(t, toContractReturnValueIntType(E_INVOKE_CONTRACT, int64(updateNodeSuccess)), ret)
}
