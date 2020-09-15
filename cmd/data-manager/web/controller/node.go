package controller

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/model"
	webCtx "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/context"
	webEngine "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/engine"
	"net/http"
)

func init() {
	webEngine.Default.GET("/nodes", webEngine.NewHandler(defaultNodeController.Nodes))
}

type nodeController struct{}

var defaultNodeController = &nodeController{}

func (this *nodeController) Nodes(ctx *webCtx.Context) {
	ret, err := model.DefaultNode.AllNodes(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.IndentedJSON(200, ret)
}
