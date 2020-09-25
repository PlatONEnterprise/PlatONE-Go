package controller

import (
	"data-manager/model"
	webCtx "data-manager/web/context"
	webEngine "data-manager/web/engine"
	"net/http"
	"strconv"
)

func init() {
	webEngine.Default.GET("/txs", webEngine.NewHandler(defaultTxController.Txs))
	webEngine.Default.GET("/tx/:hash", webEngine.NewHandler(defaultTxController.Tx))
	webEngine.Default.GET("/block/:block_height/txs", webEngine.NewHandler(defaultTxController.TxsInHeight))
	webEngine.Default.GET("/address/from/:from_address/txs", webEngine.NewHandler(defaultTxController.TxsFromAddress))
}

type txController struct{}

var defaultTxController = &txController{}

func (this *txController) Tx(ctx *webCtx.Context) {
	hash := ctx.Param("hash")

	ret, err := model.DefaultTx.TxByHash(ctx.DBCtx, hash)
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(200, ret)
}

func (this *txController) Txs(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	setPageDefaultIfEmpty(&p)

	result, err := model.DefaultTx.Txs(ctx.DBCtx, p.PageIndex, p.PageSize)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	totalTx, err := model.DefaultTx.TotalTx(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.IndentedJSON(200, newPageInfo(p.PageIndex, p.PageSize, totalTx, result))
}

func (this *txController) TxsInHeight(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	setPageDefaultIfEmpty(&p)

	height, err := strconv.ParseUint(ctx.Param("block_height"), 10, 64)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	result, err := model.DefaultTx.TxsInHeight(ctx.DBCtx, p.PageIndex, p.PageSize, height)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	block, err := model.DefaultBlock.BlockByHeight(ctx.DBCtx, height)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.IndentedJSON(200, newPageInfo(p.PageIndex, p.PageSize, int64(block.TxAmount), result))
}

func (this *txController) TxsFromAddress(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	fromAddr := ctx.Param("from_address")
	setPageDefaultIfEmpty(&p)

	result, err := model.DefaultTx.TxsFromAddress(ctx.DBCtx, p.PageIndex, p.PageSize, fromAddr)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	totalTx, err := model.DefaultTx.TotalTx(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.IndentedJSON(200, newPageInfo(p.PageIndex, p.PageSize, totalTx, result))
}
