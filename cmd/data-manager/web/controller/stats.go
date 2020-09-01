package controller

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/model"
	webCtx "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/context"
	webEngine "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/engine"
	"net/http"
)

func init() {
	webEngine.Default.GET("/stats", webEngine.NewHandler(defaultStatsController.Stats))
}

type statsController struct{}

var defaultStatsController = &statsController{}

func (this *statsController) Stats(ctx *webCtx.Context) {
	ret, err := model.DefaultStats.Stats(ctx.DBCtx)
	if nil != err {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, ret)
}
