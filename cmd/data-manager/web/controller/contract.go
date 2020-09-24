package controller

import (
	"data-manager/model"
	"data-manager/util"
	webCtx "data-manager/web/context"
	webEngine "data-manager/web/engine"
	"encoding/hex"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/sirupsen/logrus"
	"net/http"
)

func init() {
	webEngine.Default.GET("/contracts", webEngine.NewHandler(defaultContractController.Contracts))
	webEngine.Default.GET("/cns", webEngine.NewHandler(defaultContractController.CNS))
	webEngine.Default.GET("/contract/:address", webEngine.NewHandler(defaultContractController.Contract))
}

type contractController struct{}

var defaultContractController = &contractController{}

func (this *contractController) Contracts(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	setPageDefaultIfEmpty(&p)

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
			logrus.Warningln(err)
			//ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
			//return
		} else {
			c.CNSName = cns.Name
		}

		cs = append(cs, &c)
	}

	totalContract, err := model.DefaultTx.TotalContract(ctx.DBCtx)
	if nil != err {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.IndentedJSON(200, newPageInfo(p.PageIndex, p.PageSize, totalContract, cs))
}

func (this *contractController) CNS(ctx *webCtx.Context) {
	var p page
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	setPageDefaultIfEmpty(&p)

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

func (this *contractController) Contract(ctx *webCtx.Context) {
	contractAddress := ctx.Param("address")

	result, err := model.DefaultTx.ContractByAddress(ctx.DBCtx, contractAddress)
	if nil != err {
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	contract := struct {
		Address   string `json:"address"`
		CNSName   string `json:"name"`
		Creator   string `json:"creator"`
		TxHash    string `json:"tx_hash"`
		Timestamp int64  `json:"timestamp"`
		Code      string `json:"code"`
	}{
		result.Receipt.ContractAddress,
		"",
		result.From,
		result.Hash,
		result.Timestamp,
		"",
	}

	cns, err := util.GetCNSByAddress(result.Receipt.ContractAddress)
	if nil != err {
		logrus.Warningln(err)
	} else {
		contract.CNSName = cns.Name
	}

	code, err := util.DefaultNode.CodeAt(common.HexToAddress(result.Receipt.ContractAddress))
	if nil != err {
		logrus.Warningln(err)
	} else {
		contract.Code = hex.EncodeToString(code)
	}

	ctx.IndentedJSON(200, contract)
}
