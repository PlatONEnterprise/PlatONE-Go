package controller

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/model"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/util"
	webCtx "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/context"
	webEngine "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/engine"
	"github.com/sirupsen/logrus"
	"net/http"
)

func init() {
	webEngine.Default.GET("/contracts", webEngine.NewHandler(defaultContractController.Contracts))
	webEngine.Default.GET("/cns", webEngine.NewHandler(defaultContractController.CNS))
}

type contractController struct{}

var defaultContractController = &contractController{}

func (this *contractController) Contracts(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	result, err := model.DefaultTx.Contracts(ctx.DBCtx, p.PageIndex, p.PageSize)
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	var cs []*contract
	for _, tx := range result {
		var c contract
		c.Address = tx.Receipt.ContractAddress
		c.Creator = tx.From
		c.Timestamp = tx.Timestamp
		c.TxHash = tx.Hash
		cns, err := util.GetCNSByAddress(c.Address)
		if nil != err {
			logrus.Errorln(err)
			//ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
			//return
		} else {
			c.CNSName = cns.Name
		}

		cs = append(cs, &c)
	}

	stats, err := model.DefaultStats.Stats(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.IndentedJSON(200, newPageInfo(p.PageIndex, p.PageSize, int64(stats.TotalContract), cs))
}

func (this *contractController) CNS(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	result, err := model.DefaultCNS.QueryCNS(ctx.DBCtx, p.PageIndex, p.PageSize)
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	count, err := model.DefaultCNS.Total(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.IndentedJSON(200, newPageInfo(p.PageIndex, p.PageSize, count, result))
}
