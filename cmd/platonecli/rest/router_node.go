package rest

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
	"github.com/gin-gonic/gin"
)

func registerNodeRouters(r *gin.Engine) {
	node := r.Group("/node")
	{
		node.POST("/components", nodeAddHandler)              // add - resource: nodeInfo
		node.DELETE("/components/:nodeID", nodeDeleteHandler) // delete
		node.PATCH("/components/:nodeID", nodeUpateHandler)   // update

		node.GET("/components", nodeGetHandler)           // getAllNodes, getNodes
		node.GET("/components/statistic", nodeGetHandler) // nodesNum

		node.GET("/enode/deleted", enodeGetHandler) // getDeletedEnodeNodes
		node.GET("/enode/normal", enodeGetHandler)  // getNormalEnodeNodes

		// lack of importOldNodesData, ValidJoinNode
	}
}

// ===================== Node ========================
type NodeInfo struct {
	// required
	Name      string `form:"name"`
	Status    uint32
	PublicKey string
	P2pPort   uint32

	// optional
	Owner      string
	Desc       string
	Type       uint32
	ExternalIP string
	InternalIP string
	RpcPort    uint32
	DelayNum   uint64
}

func (c *NodeInfo) string() string {
	jsonBytes, _ := json.Marshal(c)
	return string(jsonBytes)
}

func nodeAddHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	funcParams := &struct {
		Info *NodeInfo
	}{}

	data := newContractParams(contractAddr, "add", "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func nodeDeleteHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	name := ctx.Param("nodeID")

	funcParams := &struct {
		name string
		Info interface{}
	}{
		name: name,
		Info: &struct {
			Status uint32
		}{},
	}

	data := newContractParams(contractAddr, "update", "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func nodeUpateHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	name := ctx.Param("nodeID")

	funcParams := &struct {
		name string
		Info interface{}
	}{
		name: name,
		Info: &struct {
			Desc     string
			Type     uint32
			DelayNum uint64
		}{},
	}

	data := newContractParams(contractAddr, "update", "wasm", nil, funcParams)

	posthandlerCommon(ctx, data)
}

func nodeGetHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	var node = new(NodeInfo)
	var funcName string
	var funcParams interface{}

	endPoint := ctx.Query("endPoint")
	err := ctx.BindQuery(node)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !reflect.ValueOf(node).Elem().IsZero() {
		funcName = "getNodes"
		if strings.Contains(ctx.FullPath(), "/statistic") {
			funcName = "nodesNum"
		}

		funcParams = &struct {
			Param *NodeInfo
		}{Param: node}
	} else {
		funcName = "getAllNodes"
		funcParams = nil
	}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	queryHandlerCommon(ctx, endPoint, data)

}

func enodeGetHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	var funcName string

	endPoint := ctx.Query("endPoint")

	funcName = "getNormalEnodeNodes"
	if strings.Contains(ctx.FullPath(), "/enode/deleted") {
		funcName = "getDeletedEnodeNodes"
	}

	data := newContractParams(contractAddr, funcName, "wasm", nil, nil)
	queryHandlerCommon(ctx, endPoint, data)
}
