package controller

import (
	"data-manager/model"
	webCtx "data-manager/web/context"
	webEngine "data-manager/web/engine"
	"net/http"
	"strconv"
)

func init() {
	webEngine.Default.GET("/stats", webEngine.NewHandler(defaultStatsController.Stats))
	webEngine.Default.GET("/stats/tx/count", webEngine.NewHandler(defaultStatsController.TxAmountStats))
}

type statsController struct{}

var defaultStatsController = &statsController{}

func (this *statsController) Stats(ctx *webCtx.Context) {
	ret, err := model.GetStats(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(200, ret)
}

func (this *statsController) TxAmountStats(ctx *webCtx.Context) {
	num := ctx.Query("num")
	period, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ret, err := model.DefaultTxStats.History(ctx.DBCtx, int64(period))
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, ret)
}
