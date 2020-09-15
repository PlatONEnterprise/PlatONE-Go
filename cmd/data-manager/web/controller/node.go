package controller

import (
	"data-manager/model"
	webCtx "data-manager/web/context"
	webEngine "data-manager/web/engine"
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
