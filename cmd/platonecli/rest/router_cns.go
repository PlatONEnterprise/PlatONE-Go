package rest

import (
	"errors"
	"fmt"
	"net/http"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
	"github.com/gin-gonic/gin"
)

func registerCnsRouters(r *gin.Engine) {
	cns := r.Group("/cns")
	{
		cns.POST("/components", cnsRegisterHandler)        // register - resource: cnsInfo
		cns.GET("/components", cnsQueryHandler)            // query 	- resource: cnsInfo getRegisteredContracts/ByName/ByAddress/ByOrigin
		cns.GET("/components/state", cnsQueryStateHandler) // state 	- resource: cnsInfo ifRegisteredByName, ifRegisteredByAddress

		cns.GET("/mappings/:name", cnsMappingGetHandler)  // resolve 	- resource: the mapping of an address and a name
		cns.PUT("/mappings/:name", cnsMappingPostHandler) // redirect - resource: the mapping of an address and a name
	}
}

//======================= CNS ==========================

// POST/PATCH/PUT/DELETE
func cnsRegisterHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	params := &struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Address string `json:"address"`
	}{}

	data := newContractParams(contractAddr, "cnsRegister", "wasm", nil, params)

	posthandlerCommon(ctx, data)
}

func cnsQueryHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	var funcName string
	var queryRange string

	// todo: if endPoint is null?
	endPoint := ctx.Query("endPoint")

	// todo: ctx.ShouldBindQuery ???
	name := ctx.Query("name")
	address := ctx.Query("address")
	origin := ctx.Query("origin")
	pageNum := ctx.Query("page-num")
	pageSize := ctx.Query("page-size")
	if pageNum != "" || pageSize != "" {
		queryRange = "not null"
	}

	if countQueryNum(name, address, origin, queryRange) > 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errExceedQueryKeyLimit.Error()})
		return
	}

	funcParams := &struct {
		Name    string
		Address string
		Origin  string
		Range   string
	}{}

	switch {
	case name != "":
		if !cmd_common.ParamValidWrap(name, "name") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errInvalidParam.Error()})
			return
		}

		funcName = "getRegisteredContractsByName"
		funcParams.Name = name
	case address != "":
		if !cmd_common.ParamValidWrap(address, "address") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errInvalidParam.Error()})
			return
		}

		funcName = "getRegisteredContractsByAddress"
		funcParams.Address = address
	case origin != "":
		if !cmd_common.ParamValidWrap(address, "address") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errInvalidParam.Error()})
			return
		}

		funcName = "getRegisteredContractsByOrigin"
		funcParams.Origin = origin
	case queryRange != "":
		if pageNum == "" {
			pageNum = "0"
		}
		if pageSize == "" {
			pageSize = "0"
		}

		if cmd_common.ParamValidWrap(pageNum, "num") && cmd_common.ParamValidWrap(pageSize, "num") {
			queryRange = fmt.Sprintf("(%s,%s)", pageNum, pageSize)
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errInvalidParam.Error()})
			return
		}

		funcName = "getRegisteredContracts"
		funcParams.Range = queryRange
	default:
		funcName = "getRegisteredContracts"
		funcParams.Range = "(0,0)"
	}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	queryHandlerCommon(ctx, endPoint, data)
}

func cnsQueryStateHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	var funcName string
	/// var funcParams []string
	var funcParams interface{}

	name := ctx.Query("name")
	address := ctx.Query("address")
	endPoint := ctx.Query("endPoint")

	if countQueryNum(name, address) > 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errExceedQueryKeyLimit.Error()})
		return
	}

	switch {
	case name != "":
		funcName = "ifRegisteredByName"
		funcParams = &struct {
			Name string
		}{Name: name}

	case address != "":
		funcName = "ifRegisteredByAddress"
		funcParams = &struct {
			Address string
		}{Address: address}

	default:
		err := errors.New("invalid search key")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	queryHandlerCommon(ctx, endPoint, data)
}

func countQueryNum(args ...string) (count int) {
	for _, data := range args {
		if data != "" {
			count++
		}
	}

	return
}

// ---------------------- Cns Mappings --------------------------

func cnsMappingGetHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress

	name := ctx.Param("name")
	version := ctx.Query("version")
	endPoint := ctx.Query("endPoint")

	// todo: paramCheck same as in sc_cns.go

	funcName := "getContractAddress"
	/// funcParams := cmd_common.CombineFuncParams(name, version)
	funcParams := &struct {
		Name    string
		Version string
	}{name, version}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	queryHandlerCommon(ctx, endPoint, data)
}

func cnsMappingPostHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress

	name := ctx.Param("name")

	// todo: paramCheck same as in sc_cns.go

	funcName := "cnsRedirect"
	funcParams := &struct {
		name    string
		Version string
	}{name: name}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)

	posthandlerCommon(ctx, data)
}
