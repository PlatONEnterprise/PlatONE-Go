package controller

import (
	"data-manager/exterror"
	"data-manager/model"
	webCtx "data-manager/web/context"
	webEngine "data-manager/web/engine"
	"net/http"
	"strings"
)

func init() {
	webEngine.Default.GET("/blocks", webEngine.NewHandler(defaultBlockController.Blocks))
	webEngine.Default.GET("/block", webEngine.NewHandler(defaultBlockController.Block))
}

type blockController struct{}

var defaultBlockController = &blockController{}

type blockRequest struct {
	BlockHeight uint64 `form:"block_height"`
	BlockHash   string `form:"block_hash"`
}

func (this *blockController) Block(ctx *webCtx.Context) {
	var br blockRequest
	if err := ctx.BindQuery(&br); nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(br.BlockHash) == "" && br.BlockHeight == 0 {
		ctx.IndentedJSON(http.StatusBadRequest, exterror.ErrParameterInvalid)
		return
	}

	if 0 != br.BlockHeight {
		ret, err := model.DefaultBlock.BlockByHeight(ctx.DBCtx, br.BlockHeight)
		if nil != err {
			ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(200, ret)
		return
	}

	ret, err := model.DefaultBlock.BlockByHash(ctx.DBCtx, strings.TrimSpace(br.BlockHash))
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(200, ret)
}

func (this *blockController) Blocks(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	setPageDefaultIfEmpty(&p)

	result, err := model.DefaultBlock.Blocks(ctx.DBCtx, p.PageIndex, p.PageSize)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	block, err := model.DefaultBlock.LatestBlock(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.IndentedJSON(200, newPageInfo(p.PageIndex, p.PageSize, int64(block.Height), result))
}
