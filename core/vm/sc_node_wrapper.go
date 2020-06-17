package vm

import (
	"encoding/json"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"strings"
)

const (
	ADD_NODE_SUCCESS       = 0
	ADD_NODE_BAD_PARAMETER = 1
	ADD_NODE_NO_PERMISSION = 2
	//new
	//todo 确认：新增了一个返回码
	ADD_NODE_INTERNAL_SERVER_ERROR = 3
)

const (
	UPDATE_NODE_SUCCESS = 0
	UPDATE_NODE_FAILED  = 1
)

const (
	PUBLIC_KEY_NOT_EXIST      = 0
	PUBLIC_KEY_EXIST          = 1
	PUBLIC_KEY_INTERNAL_ERROR = 2
)

const (
	RESULT_CODE_SUCCESS        = 0
	RESULT_CODE_INTERNAL_ERROR = 1
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewResult(code int, msg string, data interface{}) *Result {
	return &Result{Code: code, Msg: msg, Data: data}
}

func NewSuccessResult(data interface{}) *Result {
	return NewResult(RESULT_CODE_SUCCESS, "success", data)
}

func (res *Result) String() string {
	b, err := json.Marshal(res)
	if err != nil {
		return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, RESULT_CODE_INTERNAL_ERROR, err.Error())
	}

	return string(b)
}

type SCNodeWrapper struct {
	base *SCNode
}

func NewSCNodeWrapper() *SCNodeWrapper {
	return &SCNodeWrapper{NewSCNode()}
}

func (n *SCNodeWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.SCNodeGas
}

func (n *SCNodeWrapper) Run(input []byte) ([]byte, error) {
	return execSC(input, n.AllExportFns())
}

func (n *SCNodeWrapper) Add(node *NodeInfo) int {
	if err := n.base.Add(node); nil != err {
		switch err {
		case ErrNoPermissionManageSCNode:
			return ADD_NODE_NO_PERMISSION
		case ErrParamsInvalid:
			return ADD_NODE_BAD_PARAMETER
		default:
			return ADD_NODE_INTERNAL_SERVER_ERROR
		}
	}

	return ADD_NODE_SUCCESS
}

//todo 确认： 旧的接口返回值表示更新字段个个数（感觉有点奇怪！！），现在返回值表示更新操作是否成功
func (n *SCNodeWrapper) Update(name string, node *UpdateNode) int {
	err := n.base.Update(name, node)
	if err != nil {
		return UPDATE_NODE_FAILED
	}

	return UPDATE_NODE_SUCCESS
}

//todo 确认：旧接口一律返回成功，没有失败。data值为空时返回的时[]
func (n *SCNodeWrapper) GetAllNodes() string {
	nodes, err := n.base.GetAllNodes()
	if err != nil {
		return NewResult(RESULT_CODE_INTERNAL_ERROR, err.Error(), nil).String()
	}

	return NewSuccessResult(nodes).String()
}

//todo 确认: 旧的接口只返回了pub公钥是否存在，存在返回1，不存在返回0
//for c++ interface: validJoinNode
func (n *SCNodeWrapper) IsPublicKeyExist(pub string) int {
	err := n.base.CheckPublicKeyExist(pub)
	if err != nil {
		if ErrPublicKeyExist == err {
			return PUBLIC_KEY_EXIST
		}

		return PUBLIC_KEY_INTERNAL_ERROR
	}

	return PUBLIC_KEY_NOT_EXIST
}

//enode format: "enode://" + publicKey + "@" + ip:port
//for c++ sc interface: getNormalEnodeNodes
//todo C++这个接口返回的值是自定义格式，是不是应该采用通用的格式？都采用json会比较好？
func (n *SCNodeWrapper) GetENodesOfAllNormalNodes() string {
	enodes, err := n.base.GetENodesOfAllNormalNodes()
	if err != nil {
		return ""
	}

	return ENodesToString(enodes)
}

func ENodesToString(enodes []*ENode) string {
	ret := make([]string, 0, len(enodes))
	for _, v := range enodes {
		ret = append(ret, v.String())
	}

	return strings.Join(ret, "|")
}

func (n *SCNodeWrapper) GetENodesOfAllDeletedNodes() string {
	enodes, err := n.base.GetENodesOfAllDeletedNodes()
	if err != nil {
		return ""
	}

	return ENodesToString(enodes)
}

//TODO 需要支持像C++合约一样组合搜索吗
func (n *SCNodeWrapper) GetNodes(query *NodeInfo) string {
	nodes, err := n.base.GetNodes(query)
	if err != nil {
		return NewResult(RESULT_CODE_INTERNAL_ERROR, err.Error(), nil).String()
	}

	return NewSuccessResult(nodes).String()
}

//todo 旧接口是定义查询错误时的返回值。 返回值0有多种含义。
func (n *SCNodeWrapper) NodesNum(query *NodeInfo) int {
	num, err := n.base.NodesNum(query)
	if err != nil {
		return 0
	}

	return num
}

//for access control
func (n *SCNodeWrapper) AllExportFns() SCExportFns {
	return SCExportFns{
		"add":                  n.Add,
		"update":               n.Update,
		"getAllNodes":          n.GetAllNodes,
		"getNodes":             n.GetNodes,
		"getNormalEnodeNodes":  n.GetENodesOfAllNormalNodes(),
		"getDeletedEnodeNodes": n.GetENodesOfAllDeletedNodes,
		"validJoinNode":        n.IsPublicKeyExist,
		"nodesNum":             n.NodesNum,
	}
}
