package rest

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/keystore"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/pborman/uuid"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/precompiled"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/platoneclient"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/gin-gonic/gin"
)

const defaultEndPoint = "http://127.0.0.1:6791"
const separation = "-"

var (
	errExceedQueryKeyLimit = errors.New("please provide at most ONE query key")
	errInvalidParam        = errors.New("invalid params")
	errPollingReceipt      = errors.New("polling transaction receipt error")
)

func StartServer(endPoint string) {
	r := genRestRouters()
	_ = r.Run(endPoint)
}

func genRestRouters() *gin.Engine {
	router := gin.New()

	// todo: custom middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	registerRouters(router)

	return router
}

func registerRouters(r *gin.Engine) {
	registerAccountRouters(r)
	registerCnsRouters(r)
	registerContractRouters(r)
	registerFwRouters(r)
	registerNodeRouters(r)
	registerRoleRouters(r)
	registerSysConfigRouters(r)

	/// registerRpcAPIs(r)
}

func registerRpcAPIs(r *gin.Engine) {
	r.GET("/blockNum", getBlockNumHandler)
}

func registerAccountRouters(r *gin.Engine) {
	cns := r.Group("/accounts")
	{
		cns.POST("", newAccountHandler)
	}
}

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

func registerContractRouters(r *gin.Engine) {
	contract := r.Group("/contracts")
	{
		contract.POST("", deployHandler)           // deploy 	- resource: contract
		contract.PUT("/:address", migrateHandler)  // migrate	- resource: contract -> new contract
		contract.POST("/:address", executeHandler) // execute	- resource: methods of contract
	}
}

func registerNodeRouters(r *gin.Engine) {
	node := r.Group("/node")
	{
		node.POST("/components", nodeAddHandler)              // add - resource: nodeInfo
		node.DELETE("/components/:nodeID", nodeDeleteHandler) // delete
		node.PATCH("/components/:nodeID", nodeUpateHandler)   // update

		node.GET("/components", nodeGetHandler)           // getAllNodes, getNodes
		node.GET("/components/statistic", nodeGetHandler) // nodesNum

		node.GET("/enode/deleted", enodeGetHandler) // getDeletedEnodeNodes
		node.GET("/enode/normal", enodeGetHandler)  // getNormalEnodeNodes

		// lack of importOldNodesData, ValidJoinNode
	}
}

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

func registerRoleRouters(r *gin.Engine) {
	role := r.Group("/role")
	{
		/// role.POST("/super-admin", setSupAdminHandler)
		/// role.PUT("/super-admin", transferSupAdminHandler)
		/// role.POST("/roles-of-user/:addressOrName", roleAddHandler)
		/// role.DELETE("/roles-of-user/:addressOrName", roleDelHandler)

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

		/// role.GET("/:param", roleGetHandler) // getRolesByAddress, getRolesByName, getAddrListOfRole
		role.GET("/user-lists/:addressOrName", roleGetUserListsHandler)
		role.GET("/role-lists/:role", roleGetRoleListsHandler)
	}
}

func registerSysConfigRouters(r *gin.Engine) {
	sysConf := r.Group("/sysConfig")
	{
		sysConf.PUT("/block-gas-limit", blockGasLimitHandler)
		sysConf.PUT("/tx-gas-limit", txGasLimitHandler)
		sysConf.PUT("/is-tx-use-gas", isTxUseGasHandler)
		sysConf.PUT("/is-approve-deployed-contract", isApproveDeployedContractHandler)
		sysConf.PUT("/is-check-contract-deploy-permission", isCheckContractDeployPermissionHandler)
		sysConf.PUT("/is-produce-empty-block", isProduceEmptyBlockHandler)
		sysConf.PUT("/gas-contract-name", gasContractNameHandler)

		sysConf.GET("/block-gas-limit", sysConfigGetHandler)
		sysConf.GET("/tx-gas-limit", sysConfigGetHandler)
		sysConf.GET("/is-tx-use-gas", sysConfigGetHandler)
		sysConf.GET("/is-approve-deployed-contract", sysConfigGetHandler)
		sysConf.GET("/is-check-contract-deploy-permission", sysConfigGetHandler)
		sysConf.GET("/is-produce-empty-block", sysConfigGetHandler)
		sysConf.GET("/gas-contract-name", sysConfigGetHandler)
	}
}

// ================= RPC =========================

type tempRpc struct {
	Params map[string]string
	Rpc    *rpcClientParams
}

func newTempRpc() *tempRpc {
	return &tempRpc{
		Params: make(map[string]string, 0),
		Rpc:    new(rpcClientParams),
	}
}

func getBlockNumHandler(ctx *gin.Context) {
	// todo: how to receive parameters
	var jsonInfo = newTempRpc()

	if err := ctx.ShouldBind(jsonInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pc, _ := platoneclient.SetupClient(jsonInfo.Rpc.EndPoint)

	_, _ = pc.GetTransactionReceipt(jsonInfo.Params["txHash"])
}

// ===================== Deploy =======================
type deployInfo struct {
	codeBytes string
	abiBytes  string
	Params    string `json:"params"`
}

func deployHandler(ctx *gin.Context) {
	var jsonInfo = newTemp()
	var fileBytes = make([][]byte, 2)
	var funcParams = new(deployInfo)

	// read file
	form, _ := ctx.MultipartForm()
	files := form.File["files"]

	for i, file := range files {
		f, _ := file.Open()
		fileBytes[i], _ = ioutil.ReadAll(f)
	}
	funcParams.codeBytes = string(fileBytes[0])
	funcParams.abiBytes = string(fileBytes[1])
	jsonInfo.Contract = newContractParams("", "", "", nil, funcParams)

	// read parameters
	info := form.Value["info"][0]
	err := json.Unmarshal([]byte(info), jsonInfo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := deploy(jsonInfo)
	if err != nil {
		if err == errPollingReceipt {
			ctx.JSON(200, gin.H{
				"txHash": res[0],
			})
			return
		}

		// todo: error code
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{
		"result": res[0],
	})
}

// todo: refactory
func deploy(jsonInfo *temp) ([]interface{}, error) {

	vm := jsonInfo.Contract.Interpreter
	/// data := jsonInfo.Contract.Data.getDataParams()
	data, _ := getDataParams(jsonInfo.Contract.Data)
	dataGenerator := packet.NewDeployDataGen([]byte(data[0]), []byte(data[1]), data[2:], vm, types.CreateTxType)

	from := common.HexToAddress(jsonInfo.Tx.From)
	tx := packet.NewTxParams(from, nil, "", "", "", "")

	keyfile := parseKeyfile(jsonInfo.Tx.From)
	keyfile.Passphrase = jsonInfo.Rpc.Passphrase

	return A(jsonInfo.Rpc.EndPoint, dataGenerator, tx, keyfile)
}

//=======================
func migrateHandler(ctx *gin.Context) {

}

func executeHandler(ctx *gin.Context) {
	var jsonInfo = newTemp()
	contractAddr := ctx.Param("address")

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	f, _ := file.Open()
	abiBytes, _ := ioutil.ReadAll(f)

	// read parameters
	info := ctx.PostForm("info")

	funcParams := &struct {
		Params string
	}{}
	data := newContractParams(contractAddr, "", "", abiBytes, funcParams)
	jsonInfo.Contract = data

	err = json.Unmarshal([]byte(info), jsonInfo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := handlerCallCommon(jsonInfo)
	if err != nil {
		if err == errPollingReceipt {
			ctx.JSON(200, gin.H{
				"txHash": res[0],
			})
			return
		}
		// todo: error code
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(res) == 1 {
		ctx.JSON(200, gin.H{
			"result": res[0],
		})
		return
	}

	ctx.JSON(200, gin.H{
		"result": res,
	})
}

// ===================== SYS ===========================

type contractData interface {
	/// getDataParams() []string
	/// paramsCheck() bool
}

type txParams struct {
	From string `json:"from"` // the address used to send the transaction
	/// To       string  `json:"to,omitempty"`   // the address receives the transactions
	Gas      string `json:"gas,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	Value    string `json:"value,omitempty"`
	Data     string `json:"data,omitempty"`
}

type contractParams struct {
	ContractAddr string `json:"contractAddr,omitempty"`
	Method       string `json:"method,omitempty"`
	Interpreter  string `json:"interpreter,omitempty"`
	abiMethods   []byte `json:"-"`
	/// Data         contractData `json:"data"`
	Data interface{} `json:"data"`
}

func newContractParams(defaultAddr, defaultMethod, defaultInter string, abiBytes []byte, dataParams interface{}) *contractParams {
	return &contractParams{
		ContractAddr: defaultAddr,
		Method:       defaultMethod,
		Interpreter:  defaultInter,
		abiMethods:   abiBytes,
		Data:         dataParams,
	}
}

type rpcClientParams struct {
	EndPoint   string `json:"endPoint"`
	Passphrase string `json:"passphrase"`
}

// josnInfo:
// {
// 	"tx": {
//		"from": "0x9ccf0b561c9142d3a771ce2131db8bc9fba61f6f"
//	},
//	"contract": {
//		"data": {
//			"name": "tofu",
//			"version": "0.0.0.1",
//		},
//		"interpreter": "wasm"
//	},
//	"rpc": {
//		"endPoint": "http://127.0.0.1:6791"
//	}
// }
type temp struct {
	Tx       *txParams
	Contract *contractParams
	Rpc      *rpcClientParams
}

func newTemp() *temp {
	return &temp{
		Tx:       new(txParams),
		Contract: new(contractParams),
		Rpc:      &rpcClientParams{EndPoint: defaultEndPoint},
	}
}

// =====================================================
// todo: code optimization
func getDataParams(i interface{}) ([]string, error) {
	var funcParams []string
	if i == nil {
		return nil, nil
	}

	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, errors.New("todo")
	}

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)

		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() == reflect.Struct || value.Kind() == reflect.Interface {
			marshalBytes, _ := json.Marshal(value.Interface())
			funcParams = append(funcParams, string(marshalBytes))
			continue
		} else if value.Type().Kind() != reflect.String {
			return nil, errors.New("todo")
		}

		temp := value.String()
		temp = strings.TrimSpace(temp)
		if temp != "" {
			if strings.Index(temp, "(") == 0 && strings.LastIndex(temp, ")") == len(temp)-1 {
				/// temp = abi.TrimSpace(temp)
				funcParams = append(funcParams, abi.GetFuncParams(temp[1:len(temp)-1])...)
			} else {
				funcParams = append(funcParams, temp)
			}
		}
	}

	return funcParams, nil
}

func paramsCheck(i interface{}) bool {
	var valid = true
	if i == nil {
		return true
	}

	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Name
		val := v.Field(i)

		if val.Kind() == reflect.Struct || val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
			/// valid = paramsCheck(val.Interface())
		} else if val.Kind() != reflect.String {
			return false
		}

		value := v.Field(i).String()
		if value != "" {
			valid = cmd_common.ParamValidWrap(value, strings.ToLower(key))
		}
	}

	return valid
}

// ====================== ACCOUNT ======================
const defaultKeyfile = "./keystore"

func newAccountHandler(ctx *gin.Context) {
	// password
	passphrase := ctx.PostForm("passphrase")
	if passphrase == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "the passphrase can not be empty"})
		return
	}

	runPath := utl.GetRunningTimePath()
	keyfileDirt := runPath + defaultKeyfile

	utl.FileDirectoryInit(keyfileDirt)
	pathSep := string(os.PathSeparator)

	var privateKey *ecdsa.PrivateKey
	var err error
	if file := ctx.PostForm("privatekey"); file != "" {
		// Load private key from file.
		privateKey, err = crypto.LoadECDSA(file)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"Can't load private key: ": err.Error()})
			return
		}
	} else {
		// If not loaded, generate random.
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"Failed to generate random private key: ": err.Error()})
			return
		}
	}

	// Create the keyfile object with a random UUID.
	id := uuid.NewRandom()
	key := &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}

	// Encrypt key with passphrase.
	keyjson, err := keystore.EncryptKey(key, passphrase, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error encrypting key: ": err.Error()})
		return
	}

	keyfilepath := keyfileDirt + pathSep + "UTC--" + time.Now().Format("2006-01-02") + "--" + key.Address.Hex()

	// Store the file to disk.
	if err := os.MkdirAll(filepath.Dir(keyfilepath), 0700); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Failed to generate random private key: ": err.Error()})
		return
	}
	if err := ioutil.WriteFile(keyfilepath, keyjson, 0600); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Failed to write keyfile: ": err.Error()})
		return
	}

	// Output some information.
	ctx.JSON(200, gin.H{
		"Address": key.Address.Hex(),
	})

}

//======================= CNS ==========================
type cnsInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Address string `json:"address"`
}

func (c *cnsInfo) getDataParams() []string {
	return cmd_common.CombineFuncParams(c.Name, c.Version, c.Address)
}

func (c *cnsInfo) paramsCheck() bool {
	valid := c.Address == "" || utl.IsMatch(c.Address, "address")
	valid = c.Name == "" || utl.IsMatch(c.Name, "name")
	valid = c.Version == "" || utl.IsMatch(c.Version, "version")
	return valid
}

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

// GET
type GetS struct {
	funcParams []string
}

func newGetS(funcParams []string) *GetS {
	return &GetS{funcParams: funcParams}
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

func cnsQueryStateByNameHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	name := ctx.Param("name")
	endPoint := ctx.Query("endPoint")

	funcName := "ifRegisteredByName"
	funcParams := []string{name}

	data := newContractParams(contractAddr, funcName, "wasm", nil, newGetS(funcParams))
	queryHandlerCommon(ctx, endPoint, data)
}

func cnsQueryStateByAddressHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	address := ctx.Param("address")
	endPoint := ctx.Query("endPoint")

	funcName := "ifRegisteredByAddress"
	funcParams := []string{address}

	data := newContractParams(contractAddr, funcName, "wasm", nil, newGetS(funcParams))
	queryHandlerCommon(ctx, endPoint, data)
}

func cnsQueryHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	/// var funcParams = make([]string, 1)
	var funcName string

	// todo: if endPoint is null?
	endPoint := ctx.Query("endPoint")

	// todo: ctx.ShouldBindQuery ???
	name := ctx.Query("name")
	address := ctx.Query("address")
	origin := ctx.Query("origin")

	if countQueryNum(name, address, origin) > 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errExceedQueryKeyLimit.Error()})
		return
	}

	funcParams := &struct {
		Name    string
		Address string
		Origin  string
	}{}

	switch {
	case name != "":
		// param check
		funcName = "getRegisteredContractsByName"
		funcParams.Name = name
	case address != "":
		// param check
		funcName = "getRegisteredContractsByAddress"
		funcParams.Address = address
	case origin != "":
		// param check
		funcName = "getRegisteredContractsByOrigin"
		funcParams.Origin = origin
	default:
		funcName = "getRegisteredContracts"
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

func cnsQueryAllHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	endPoint := ctx.Query("endPoint")

	funcName := "getRegisteredContracts"
	funcParams := []string{}

	data := newContractParams(contractAddr, funcName, "wasm", nil, newGetS(funcParams))
	queryHandlerCommon(ctx, endPoint, data)
}

func cnsQueryByAddressHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	address := ctx.Param("address")
	endPoint := ctx.Query("endPoint")

	funcName := "getRegisteredContractsByAddress"
	funcParams := []string{address}

	data := newContractParams(contractAddr, funcName, "wasm", nil, newGetS(funcParams))
	queryHandlerCommon(ctx, endPoint, data)

}

func cnsQueryByNameHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	name := ctx.Param("name")
	endPoint := ctx.Query("endPoint")

	funcName := "getRegisteredContractsByName"
	funcParams := []string{name}

	data := newContractParams(contractAddr, funcName, "wasm", nil, newGetS(funcParams))
	queryHandlerCommon(ctx, endPoint, data)

}

func cnsQueryByOriginHandler(ctx *gin.Context) {
	var contractAddr = precompile.CnsManagementAddress
	origin := ctx.Param("origin")
	endPoint := ctx.Query("endPoint")

	funcName := "getRegisteredContractsByOrigin"
	funcParams := []string{origin}

	data := newContractParams(contractAddr, funcName, "wasm", nil, newGetS(funcParams))
	queryHandlerCommon(ctx, endPoint, data)

}

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

// ====================== Role ========================
type roleInfo struct {
	Address string `json:"address"`
}

func (c *roleInfo) getDataParams() []string {
	return cmd_common.CombineFuncParams(c.Address)
}

func (c *roleInfo) paramsCheck() bool {
	valid := c.Address == "" || utl.IsMatch(c.Address, "address")
	return valid
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

	/*
		switch utl.IsNameOrAddress(Id) {
		case utl.CnsIsAddress:
			funcName = funcName + "ByAddress"
		case utl.CnsIsName:
			funcName = funcName + "ByName"
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errInvalidParam.Error()})
			return
		}*/

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

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

func roleGetUserListsHandler(ctx *gin.Context) {
	var contractAddr = precompile.UserManagementAddress
	var funcName string
	var funcParams interface{}

	endPoint := ctx.Query("endPoint")
	param := ctx.Param("addressOrName")

	switch utl.IsNameOrAddress(param) {
	case utl.CnsIsAddress:
		funcName = "getRolesByAddress"
		funcParams = &struct {
			Address string
		}{Address: param}

	case utl.CnsIsName:
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

// UrlParamConvert convert e.g. chain-admin -> chainAdmin
func UrlParamConvert(str string) string {
	var slice = make([]byte, 0)
	var count int

	for {
		index := strings.Index(str[count:], separation)
		if index == -1 {
			slice = append(slice, str[count:]...)
			break
		}

		slice = append(slice, str[count:count+index]...)
		slice = append(slice, str[count+index+1]-32)
		count += index + 2
	}

	return string(slice)
}

// UrlParamConvertV2 convert e.g. chain-admin -> CHAIN_ADMIN
func UrlParamConvertV2(str string) string {
	str = strings.ReplaceAll(str, separation, "_")
	return strings.ToUpper(str)
}

/*
func roleGetByRoleHandler(ctx *gin.Context) {
	var contractAddr = precompile.UserManagementAddress
	var funcName = "getAddrListOfRole"

	endPoint := ctx.Query("endPoint")

	role := ctx.Param("role")
	funcParams := &struct {
		Role string
	}{Role: role}

	data := newContractParams(contractAddr, funcName, "wasm", funcParams)
	queryHandlerCommon(ctx, endPoint, data)
}*/

//======================= FW ==========================
type fwInfo struct {
	Address string `json:"address"`
	Action  string `json:"action"`
	Rules   string `json:"rules"`
}

func (c *fwInfo) getDataParams() []string {
	return cmd_common.CombineFuncParams(c.Address, c.Action, c.Rules)
}

func (c *fwInfo) paramsCheck() bool {
	valid := c.Address == "" || utl.IsMatch(c.Address, "address")
	valid = c.Action == "" || (strings.EqualFold(c.Action, "accept") || strings.EqualFold(c.Action, "reject"))
	valid = c.Rules == ""
	return valid
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

// ===================== Node ========================

type nodeInfo struct {
	Name  string    `json:"name,omitempty"`
	Value *NodeInfo `json:"value"`
}

type NodeInfo struct {
	// required
	Name      string `form:"name"`
	Status    uint32
	PublicKey string
	P2pPort   uint32

	// optional
	Owner      string
	Desc       string
	Type       uint32
	ExternalIP string
	InternalIP string
	RpcPort    uint32
	DelayNum   uint64
}

func (c *NodeInfo) string() string {
	jsonBytes, _ := json.Marshal(c)
	return string(jsonBytes)
}

func (c *nodeInfo) getDataParams() []string {
	return cmd_common.CombineFuncParams(c.Name, c.Value.string())
}

func (c *nodeInfo) paramsCheck() bool {
	return true
}

func nodeAddHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	funcParams := &struct {
		Info *NodeInfo
	}{}

	data := newContractParams(contractAddr, "add", "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func nodeDeleteHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	name := ctx.Param("nodeID")

	funcParams := &struct {
		name string
		Info interface{}
	}{
		name: name,
		Info: &struct {
			Status uint32
		}{},
	}

	data := newContractParams(contractAddr, "update", "wasm", nil, funcParams)
	posthandlerCommon(ctx, data)
}

func nodeUpateHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	name := ctx.Param("nodeID")

	funcParams := &struct {
		name string
		Info interface{}
	}{
		name: name,
		Info: &struct {
			Desc     string
			Type     uint32
			DelayNum uint64
		}{},
	}

	data := newContractParams(contractAddr, "update", "wasm", nil, funcParams)

	posthandlerCommon(ctx, data)
}

func nodeGetHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	var node = new(NodeInfo)
	var funcName string

	endPoint := ctx.Query("endPoint")
	err := ctx.BindQuery(node)
	if err != nil {

	}

	if !reflect.ValueOf(node).Elem().IsZero() {
		funcName = "getNodes"
		if strings.Contains(ctx.FullPath(), "/statistic") {
			funcName = "nodesNum"
		}

	} else {
		funcName = "getAllNodes"
		node = nil
	}

	funcParams := &struct {
		Param *NodeInfo
	}{Param: node}

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)
	queryHandlerCommon(ctx, endPoint, data)

}

func enodeGetHandler(ctx *gin.Context) {
	var contractAddr = precompile.NodeManagementAddress
	var funcName string

	endPoint := ctx.Query("endPoint")

	funcName = "getNormalEnodeNodes"
	if strings.Contains(ctx.FullPath(), "/enode/deleted") {
		funcName = "getDeletedEnodeNodes"
	}

	data := newContractParams(contractAddr, funcName, "wasm", nil, nil)
	queryHandlerCommon(ctx, endPoint, data)
}

// ===================== sys config ====================
type sysConfigInfo struct {
	Value string `json:"value"`
}

func (c *sysConfigInfo) getDataParams() []string {
	return cmd_common.CombineFuncParams(c.Value)
}

func (c *sysConfigInfo) paramsCheck() bool {
	// todo:
	return true
}

func blockGasLimitHandler(ctx *gin.Context) {
	funcParams := &struct {
		BlockGasLimit string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func txGasLimitHandler(ctx *gin.Context) {
	funcParams := &struct {
		TxGasLimit string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func isTxUseGasHandler(ctx *gin.Context) {
	funcParams := &struct {
		SysParam string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func isApproveDeployedContractHandler(ctx *gin.Context) {
	funcParams := &struct {
		SysParam string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func isCheckContractDeployPermissionHandler(ctx *gin.Context) {
	funcParams := &struct {
		SysParam string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func isProduceEmptyBlockHandler(ctx *gin.Context) {
	funcParams := &struct {
		SysParam string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func gasContractNameHandler(ctx *gin.Context) {
	funcParams := &struct {
		ContractName string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func sysConfigHandler(ctx *gin.Context, funcParams interface{}) {
	var contractAddr = precompile.ParameterManagementAddress

	index := strings.LastIndex(ctx.FullPath(), "/")
	str := ctx.FullPath()[index+1:]
	funcName := "set" + strings.Title(UrlParamConvert(str))

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)

	posthandlerCommon(ctx, data)
}

func sysConfigGetHandler(ctx *gin.Context) {
	var contractAddr = precompile.ParameterManagementAddress
	endPoint := ctx.Query("endPoint")

	index := strings.LastIndex(ctx.FullPath(), "/")
	str := ctx.FullPath()[index+1:]
	funcName := "get" + strings.Title(UrlParamConvert(str))

	data := newContractParams(contractAddr, funcName, "wasm", nil, nil)
	queryHandlerCommon(ctx, endPoint, data)
}

// ===================== COMMON =======================

func queryHandlerCommon(ctx *gin.Context, endPoint string, data *contractParams) {
	var jsonInfo = newTemp()
	jsonInfo.Contract = data
	if endPoint != "" {
		jsonInfo.Rpc.EndPoint = endPoint
	}

	res, err := handlerCallCommon(jsonInfo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{
		"result": res[0],
	})
}

func posthandlerCommon(ctx *gin.Context, data *contractParams) {
	var jsonInfo = newTemp()
	jsonInfo.Contract = data

	if err := ctx.ShouldBind(jsonInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// todo: param check is incomplete
	if !paramsCheck(jsonInfo.Contract.Data) || !paramsCheck(jsonInfo.Tx) || !paramsCheck(jsonInfo.Rpc) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errInvalidParam.Error()})
		return
	}

	res, err := handlerCallCommon(jsonInfo)
	if err != nil {
		if err == errPollingReceipt {
			ctx.JSON(200, gin.H{
				"txHash": res[0],
			})
			return
		}

		// todo: error code
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{
		"result": res[0],
	})
}

func handlerCallCommon(jsonInfo *temp) ([]interface{}, error) {
	var funcAbi []byte

	// todo: function parameters
	/// funcParams := jsonInfo.Contract.Data.getDataParams()
	funcParams, _ := getDataParams(jsonInfo.Contract.Data)

	if p := precompile.List[jsonInfo.Contract.ContractAddr]; p != "" {
		funcAbi, _ = precompile.Asset(p)
	} else {
		funcAbi = jsonInfo.Contract.abiMethods
	}

	cns, err := cmd_common.CnsParse(jsonInfo.Contract.ContractAddr)
	if err != nil {
		return nil, err
	}
	to := cmd_common.ChainParamConvert(cns.To, "to").(common.Address)

	// todo: lack of virtual machine interpreter
	vm := jsonInfo.Contract.Interpreter

	dataGenerator := packet.NewContractDataGenWrap(jsonInfo.Contract.Method, funcParams, funcAbi, *cns, vm)

	// todo: lack of tx sender
	from := common.HexToAddress(jsonInfo.Tx.From)
	tx := packet.NewTxParams(from, &to, "", jsonInfo.Tx.Gas, "", "")

	keyfile := parseKeyfile(jsonInfo.Tx.From)
	keyfile.Passphrase = jsonInfo.Rpc.Passphrase

	return A(jsonInfo.Rpc.EndPoint, dataGenerator, tx, keyfile)
}

func A(url string, dataGen packet.MsgDataGen, tx *packet.TxParams, keyfile *utl.Keyfile) ([]interface{}, error) {

	pc, err := platoneclient.SetupClient(url)
	if err != nil {
		return nil, err
	}

	result, isTxHash, err := pc.MessageCall(dataGen, keyfile, tx)
	if err != nil {
		return nil, err
	}

	// todo: isSync
	if true && isTxHash {
		res, err := pc.GetReceiptByPolling(result[0].(string))
		if err != nil {
			return result, errPollingReceipt
		}

		receiptBytes, _ := json.MarshalIndent(res, "", "\t")
		fmt.Println(string(receiptBytes))

		result[0] = dataGen.ReceiptParsing(res)
	}

	return result, nil
}

func parseKeyfile(from string) *utl.Keyfile {
	fileName := utl.GetFileByKey(defaultKeyfile, from)

	if fileName != "" {
		path := defaultKeyfile + "/" + fileName
		keyfile, _ := cmd_common.KeyfileParsing(path)
		return keyfile
	}

	return &utl.Keyfile{}
}
