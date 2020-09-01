package controller

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/exterror"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/model"
	webCtx "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/context"
	webEngine "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/engine"
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
	BlockHeight uint64 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
}

func (this *blockController) Block(ctx *webCtx.Context) {
	var br blockRequest
	if err := ctx.BindQuery(&br); nil != err {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(br.BlockHash) == "" && br.BlockHeight == 0 {
		ctx.AbortWithError(http.StatusBadRequest, exterror.ErrParameterInvalid)
		return
	}

	if 0 != br.BlockHeight {
		ret, err := model.DefaultBlock.BlockByHeight(ctx.DBCtx, br.BlockHeight)
		if nil != err {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(200, ret)
		return
	}

	ret, err := model.DefaultBlock.BlockByHash(ctx.DBCtx, strings.TrimSpace(br.BlockHash))
	if nil != err {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, ret)
}

func (this *blockController) Blocks(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	result, err := model.DefaultBlock.Blocks(ctx.DBCtx, p.PageIndex, p.PageSize)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	stats, err := model.DefaultStats.Stats(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	ctx.IndentedJSON(200, newPageInfo(p.PageIndex, p.PageSize, int64(stats.LatestBlock), result))
}
