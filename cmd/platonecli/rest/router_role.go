package rest

import (
	"net/http"
	"strings"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
	"github.com/gin-gonic/gin"
)

func registerRoleRouters(r *gin.Engine) {
	role := r.Group("/role")
	{
		roleOpt := role.Group("/role-lists")
		{
			roleOpt.POST("/super-admin", setSupAdminHandler)
			roleOpt.PUT("/super-admin", transferSupAdminHandler)

			roleOpt.PATCH("/contract-deployer", roleAddHandler)
			roleOpt.PATCH("/group-admin", roleAddHandler)
			roleOpt.PATCH("/node-admin", roleAddHandler)
			roleOpt.PATCH("/contract-admin", roleAddHandler)
			roleOpt.PATCH("/chain-admin", roleAddHandler)

			roleOpt.DELETE("/contract-deployer", roleDelHandler)
			roleOpt.DELETE("/group-admin", roleDelHandler)
			roleOpt.DELETE("/node-admin", roleDelHandler)
			roleOpt.DELETE("/contract-admin", roleDelHandler)
			roleOpt.DELETE("/chain-admin", roleDelHandler)
		}

		role.GET("/user-lists/:addressOrName", roleGetUserListsHandler) // getRolesByAddress, getRolesByName
		role.GET("/role-lists/:role", roleGetRoleListsHandler)          // getAddrListOfRole
	}
}

// ====================== Role ========================
func setSupAdminHandler(ctx *gin.Context) {
	var contractAddr = precompile.UserManagementAddress

	data := newContractParams(contractAddr, "setSuperAdmin", "wasm", nil, nil)
	posthandlerCommon(ctx, data)
}

func transferSupAdminHandler(ctx *gin.Context) {
	var contractAddr = precompile.UserManagementAddress

	funcParams := &struct {
		Address string
	}{}

	data := newContractParams(contractAddr, "transferSuperAdminByAddress", "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func roleAddHandler(ctx *gin.Context) {
	roleHandler(ctx, "add")
}

func roleDelHandler(ctx *gin.Context) {
	roleHandler(ctx, "del")
}

func roleHandler(ctx *gin.Context, prefix string) {
	var contractAddr = precompile.UserManagementAddress

	index := strings.LastIndex(ctx.FullPath(), "/")
	str := ctx.FullPath()[index+1:]
	funcName := prefix + strings.Title(UrlParamConvert(str)) + "ByAddress"

	funcParams := &struct {
		Address string
	}{}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func roleGetRoleListsHandler(ctx *gin.Context) {
	var contractAddr = precompile.UserManagementAddress

	endPoint := ctx.Query("endPoint")

	param := ctx.Param("role")
	funcParams := &struct {
		Role string
	}{Role: UrlParamConvertV2(param)}

	data := newContractParams(contractAddr, "getAddrListOfRole", "wasm", nil, funcParams)
	queryHandlerCommon(ctx, endPoint, data)
}

// UrlParamConvertV2 convert e.g. chain-admin -> CHAIN_ADMIN
func UrlParamConvertV2(str string) string {
	str = strings.ReplaceAll(str, separation, "_")
	return strings.ToUpper(str)
}

func roleGetUserListsHandler(ctx *gin.Context) {
	var contractAddr = precompile.UserManagementAddress
	var funcName string
	var funcParams interface{}

	endPoint := ctx.Query("endPoint")
	param := ctx.Param("addressOrName")

	switch cmd_common.IsNameOrAddress(param) {
	case cmd_common.CnsIsAddress:
		funcName = "getRolesByAddress"
		funcParams = &struct {
			Address string
		}{Address: param}

	case cmd_common.CnsIsName:
		funcName = "getRolesByName"
		funcParams = &struct {
			Name string
		}{Name: param}

	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errInvalidParam.Error()})
		return
	}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	queryHandlerCommon(ctx, endPoint, data)
}
