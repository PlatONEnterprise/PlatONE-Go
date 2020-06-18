package vm

import (
	"encoding/json"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"strings"
)

const (
	addNodeSuccess      = 0
	addNodeBadParameter = 1
	addNodeNoPermission = 2
	//new
	//todo 确认：新增了一个返回码
	addNodeInternalServerError = 3
)

const (
	updateNodeSuccess = 0
	updateNodeFailed  = 1
)

const (
	publicKeyNotExist      = 0
	publicKeyExist         = 1
	publicKeyInternalError = 2
)

const (
	resultCodeSuccess       = 0
	resultCodeInternalError = 1
)

func eNodesToString(enodes []*eNode) string {
	ret := make([]string, 0, len(enodes))
	for _, v := range enodes {
		ret = append(ret, v.String())
	}

	return strings.Join(ret, "|")
}

type result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func newResult(code int, msg string, data interface{}) *result {
	return &result{Code: code, Msg: msg, Data: data}
}

func newSuccessResult(data interface{}) *result {
	return newResult(resultCodeSuccess, "success", data)
}

func (res *result) String() string {
	b, err := json.Marshal(res)
	if err != nil {
		return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, resultCodeInternalError, err.Error())
	}

	return string(b)
}

type scNodeWrapper struct {
	base *SCNode
}

func newSCNodeWrapper() *scNodeWrapper {
	return &scNodeWrapper{NewSCNode()}
}

func (n *scNodeWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.SCNodeGas
}

func (n *scNodeWrapper) Run(input []byte) ([]byte, error) {
	return execSC(input, n.allExportFns())
}

func (n *scNodeWrapper) add(node *NodeInfo) int {
	if err := n.base.add(node); nil != err {
		switch err {
		case errNoPermissionManageSCNode:
			return addNodeNoPermission
		case errParamsInvalid:
			return addNodeBadParameter
		default:
			return addNodeInternalServerError
		}
	}

	return addNodeSuccess
}

//todo 确认： 旧的接口返回值表示更新字段个个数（感觉有点奇怪！！），现在返回值表示更新操作是否成功
func (n *scNodeWrapper) update(name string, node *updateNode) int {
	err := n.base.update(name, node)
	if err != nil {
		return updateNodeFailed
	}

	return updateNodeSuccess
}

//todo 确认：旧接口一律返回成功，没有失败。data值为空时返回的时[]
func (n *scNodeWrapper) fetAllNodesa() string {
	nodes, err := n.base.GetAllNodes()
	if err != nil {
		return newResult(resultCodeInternalError, err.Error(), nil).String()
	}

	return newSuccessResult(nodes).String()
}

//todo 确认: 旧的接口只返回了pub公钥是否存在，存在返回1，不存在返回0
//for c++ interface: validJoinNode
func (n *scNodeWrapper) isPublicKeyExist(pub string) int {
	err := n.base.checkPublicKeyExist(pub)
	if err != nil {
		if errPublicKeyExist == err {
			return publicKeyExist
		}

		return publicKeyInternalError
	}

	return publicKeyNotExist
}

//enode format: "enode://" + publicKey + "@" + ip:port
//for c++ sc interface: getNormalEnodeNodes
//todo C++这个接口返回的值是自定义格式，是不是应该采用通用的格式？都采用json会比较好？
func (n *scNodeWrapper) getENodesOfAllNormalNodes() string {
	enodes, err := n.base.getENodesOfAllNormalNodes()
	if err != nil {
		return ""
	}

	return eNodesToString(enodes)
}

func (n *scNodeWrapper) getENodesOfAllDeletedNodes() string {
	enodes, err := n.base.getENodesOfAllDeletedNodes()
	if err != nil {
		return ""
	}

	return eNodesToString(enodes)
}

//TODO 需要支持像C++合约一样组合搜索吗
func (n *scNodeWrapper) getNodes(query *NodeInfo) string {
	nodes, err := n.base.GetNodes(query)
	if err != nil {
		return newResult(resultCodeInternalError, err.Error(), nil).String()
	}

	return newSuccessResult(nodes).String()
}

//todo 旧接口是定义查询错误时的返回值。 返回值0有多种含义。
func (n *scNodeWrapper) nodesNum(query *NodeInfo) int {
	num, err := n.base.nodesNum(query)
	if err != nil {
		return 0
	}

	return num
}

//for access control
func (n *scNodeWrapper) allExportFns() SCExportFns {
	return SCExportFns{
		"add":                  n.add,
		"update":               n.update,
		"getAllNodes":          n.fetAllNodesa,
		"getNodes":             n.getNodes,
		"getNormalEnodeNodes":  n.getENodesOfAllNormalNodes(),
		"getDeletedEnodeNodes": n.getENodesOfAllDeletedNodes,
		"validJoinNode":        n.isPublicKeyExist,
		"nodesNum":             n.nodesNum,
	}
}
