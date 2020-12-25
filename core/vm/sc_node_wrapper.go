package vm

import (
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

const (
	publicKeyNotExist = 0
	publicKeyExist    = 1
)

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
	fnName, ret, err := execSC(input, n.allExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
		n.base.emitEvent(fnName, operateFail, err.Error())

		if strings.Contains(fnName, "get") {
			return MakeReturnBytes([]byte(newInternalErrorResult(err).String())), err
		}
	}
	return ret, nil
}

func (n *scNodeWrapper) add(node *syscontracts.NodeInfo) (int, error) {
	if err := n.base.add(node); nil != err {
		switch err {
		case errNoPermissionManageSCNode:
			return int(addNodeNoPermission), err
		case errParamsInvalid:
			return int(addNodeBadParameter), err
		default:
			return int(addNodeBadParameter), err
		}
	}

	return int(addNodeSuccess), nil
}

func (n *scNodeWrapper) update(name string, node *syscontracts.UpdateNode) (int, error) {
	err := n.base.update(name, node)
	if err != nil {
		if err == errNoPermissionManageSCNode {
			return int(updateNodeNoPermission), err
		}

		return int(updateNodeBadParameter), err
	}

	return int(updateNodeSuccess), nil
}

func (n *scNodeWrapper) getAllNodes() (string, error) {
	nodes, err := n.base.GetAllNodes()
	if err != nil && err != errNodeNotFound{
		return "", err
	} else if errNodeNotFound == err {
		nodes = []*syscontracts.NodeInfo{}
	}

	return newSuccessResult(nodes).String(), nil
}

func (n *scNodeWrapper) importOldNodesData(data string) (int, error) {
	err := n.base.importOldNodesData(data)
	if err != nil {
		return -1, err
	}
	return 0, nil
}

func (n *scNodeWrapper) isPublicKeyExist(pub string) (int, error) {
	err := n.base.checkPublicKeyExist(pub)
	if err != nil {
		if errPublicKeyExist == err {
			return publicKeyExist, err
		}

		return 0, err
	}

	return publicKeyNotExist, nil
}

func (n *scNodeWrapper) getENodesOfAllNormalNodes() (string, error) {
	enodes, err := n.base.getENodesOfAllNormalNodes()
	if err != nil {
		if err == errNodeNotFound {
			return newInternalErrorResult(err).String(), err
		}

		return "", err
	}

	return newSuccessResult(enodes).String(), nil
}

func (n *scNodeWrapper) getENodesOfAllDeletedNodes() (string, error) {
	enodes, err := n.base.getENodesOfAllDeletedNodes()
	if err != nil {
		if errNodeNotFound == err {
			return newInternalErrorResult(err).String(), err
		}

		return "", err
	}

	return newSuccessResult(enodes).String(), nil
}

func (n *scNodeWrapper) getNodes(query *syscontracts.NodeInfo) (string, error) {
	nodes, err := n.base.GetNodes(query)
	if err != nil {
		if errNodeNotFound == err {
			return newInternalErrorResult(err).String(), err
		}
		return "", err
	}

	return newSuccessResult(nodes).String(), nil
}

func (n *scNodeWrapper) nodesNum(query *syscontracts.NodeInfo) (int, error) {
	num, err := n.base.nodesNum(query)
	if err != nil {
		if errNodeNotFound == err {
			return 0, err
		}

		return 0, err
	}

	return num, nil
}

func (n *scNodeWrapper) getVrfConsensusNodes() (string, error) {
	nodes, err := n.base.GetVrfConsensusNodes()
	if err != nil {
		if errNodeNotFound == err {
			return newInternalErrorResult(err).String(), err
		}
		return "", err
	}

	return newSuccessResult(nodes).String(), nil
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
		"importOldNodesData":   n.importOldNodesData,
		"getVrfConsensusNodes": n.getVrfConsensusNodes,
	}
}
