package rest

import (
	"strings"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
	"github.com/gin-gonic/gin"
)

func registerFwRouters(r *gin.Engine) {
	fw := r.Group("/fw/:address")
	{
		fw.PUT("/on", fwHandler)
		fw.PUT("/off", fwHandler)

		fw.POST("/lists", fwNewHandler)     // new
		fw.PUT("/lists", fwResetHandler)    // reset
		fw.DELETE("/lists", fwClearHandler) // clear
		fw.PATCH("/lists", fwDeleteHandler) // delete

		fw.GET("", fwGetHandler) // status
	}
}

//======================= FW ==========================
type fwInfo struct {
	Address string `json:"address"`
	Action  string `json:"action"`
	Rules   string `json:"rules"`
}

func fwHandler(ctx *gin.Context) {
	var contractAddr = precompile.FirewallManagementAddress
	var funcName string

	fwAddress := ctx.Param("address")
	funcParams := &struct {
		address string
	}{address: fwAddress}

	switch {
	case strings.Contains(ctx.FullPath(), "/on"):
		funcName = "__sys_FwOpen"
	case strings.Contains(ctx.FullPath(), "/off"):
		funcName = "__sys_FwClose"
	}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func fwNewHandler(ctx *gin.Context) {
	fwWriteHandler(ctx, "__sys_FwAdd")
}

func fwResetHandler(ctx *gin.Context) {
	fwWriteHandler(ctx, "__sys_FwSet")
}

func fwDeleteHandler(ctx *gin.Context) {
	fwWriteHandler(ctx, "__sys_FwDel")
}

func fwWriteHandler(ctx *gin.Context, funcName string) {
	var contractAddr = precompile.FirewallManagementAddress

	funcParams := new(fwInfo)
	funcParams.Address = ctx.Param("address")

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func fwClearHandler(ctx *gin.Context) {
	var contractAddr = precompile.FirewallManagementAddress

	funcParams := &struct {
		Address string
		Action  string
	}{}
	funcParams.Address = ctx.Param("address")

	data := newContractParams(contractAddr, "__sys_FwClear", "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func fwGetHandler(ctx *gin.Context) {
	var contractAddr = precompile.FirewallManagementAddress
	endPoint := ctx.Query("endPoint")

	funcParams := &struct {
		Address string
	}{}
	funcParams.Address = ctx.Param("address")

	data := newContractParams(contractAddr, "__sys_FwStatus", "wasm", nil, funcParams)
	queryHandlerCommon(ctx, endPoint, data)
}
