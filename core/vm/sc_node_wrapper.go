package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"strings"
)

const (
	publicKeyNotExist = 0
	publicKeyExist    = 1
)

func eNodesToString(enodes []*eNode) string {
	ret := make([]string, 0, len(enodes))
	for _, v := range enodes {
		ret = append(ret, v.String())
	}

	return strings.Join(ret, "|")
}

type scNodeWrapper struct {
	base *SCNode
}

func newSCNodeWrapper(db StateDB) *scNodeWrapper {
	return &scNodeWrapper{NewSCNode(db)}
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

func (n *scNodeWrapper) add(node *syscontracts.NodeInfo) (int, error) {
	if err := n.base.add(node); nil != err {
		switch err {
		case errNoPermissionManageSCNode:
			return int(addNodeNoPermission), nil
		case errParamsInvalid:
			return int(addNodeBadParameter), nil
		default:
			return 0, err
		}
	}

	return int(addNodeSuccess), nil
}

//todo 确认： 旧的接口返回值表示更新字段个个数（感觉有点奇怪！！）
func (n *scNodeWrapper) update(name string, node *syscontracts.UpdateNode) (int, error) {
	err := n.base.update(name, node)
	if err != nil {
		return 0, err
	}

	count := 0
	if node.Status != nil {
		count++
	}
	if node.Typ != nil {
		count++
	}
	if node.Desc != nil {
		count++
	}
	if node.DelayNum != nil {
		count++
	}

	return count, nil
}

//todo 确认：旧接口一律返回成功，没有失败。失败时设置data值为空，返回的时[]
func (n *scNodeWrapper) getAllNodes() (string, error) {
	nodes, err := n.base.GetAllNodes()
	if err != nil {
		if errNodeNotFound == err {
			return newResult(resultCodeInternalError, err.Error(), []string{}).String(), nil
		}

		return "", err
	}

	return newSuccessResult(nodes).String(), nil
}

//todo 确认: 旧的接口只返回了pub公钥是否存在，存在返回1，不存在返回0
func (n *scNodeWrapper) isPublicKeyExist(pub string) (int, error) {
	err := n.base.checkPublicKeyExist(pub)
	if err != nil {
		if errPublicKeyExist == err {
			return publicKeyExist, nil
		}

		return 0, err
	}

	return publicKeyNotExist, nil
}

//enode format: "enode://" + publicKey + "@" + ip:port
//todo C++这个接口返回的值是自定义格式，是不是应该采用通用的格式？都采用json会比较好？
//todo 找不到时，直接返回了空字符串，没有按照规定的json结构。
func (n *scNodeWrapper) getENodesOfAllNormalNodes() (string, error) {
	enodes, err := n.base.getENodesOfAllNormalNodes()
	if err != nil {
		if err == errNodeNotFound {
			return "", nil
		}

		return "", err
	}

	return eNodesToString(enodes), nil
}

//todo 找不到时，直接返回了空字符串，没有按照规定的json结构。
func (n *scNodeWrapper) getENodesOfAllDeletedNodes() (string, error) {
	enodes, err := n.base.getENodesOfAllDeletedNodes()
	if err != nil {
		if errNodeNotFound == err {
			return "", nil
		}

		return "", err
	}

	return eNodesToString(enodes), nil
}

//TODO 需要支持像C++合约一样组合搜索吗
func (n *scNodeWrapper) getNodes(query *syscontracts.NodeInfo) (string, error) {
	nodes, err := n.base.GetNodes(query)
	if err != nil {
		if errNodeNotFound == err {
			return newResult(resultCodeInternalError, err.Error(), []string{}).String(), nil
		}
		return "", err
	}

	return newSuccessResult(nodes).String(), nil
}

//todo 旧接口是定义查询错误时的返回值。 返回值0有多种含义。
func (n *scNodeWrapper) nodesNum(query *syscontracts.NodeInfo) (int, error) {
	num, err := n.base.nodesNum(query)
	if err != nil {
		if errNodeNotFound == err {
			return 0, nil
		}

		return 0, err
	}

	return num, nil
}

//for access control
func (n *scNodeWrapper) allExportFns() SCExportFns {
	return SCExportFns{
		"add":                  n.add,
		"update":               n.update,
		"getAllNodes":          n.getAllNodes,
		"getNodes":             n.getNodes,
		"getNormalEnodeNodes":  n.getENodesOfAllNormalNodes,
		"getDeletedEnodeNodes": n.getENodesOfAllDeletedNodes,
		"validJoinNode":        n.isPublicKeyExist,
		"nodesNum":             n.nodesNum,
	}
}
