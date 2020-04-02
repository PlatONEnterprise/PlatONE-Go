package core

import (
	"encoding/json"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"strconv"
	"strings"
	"time"

	// newly added
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
)

var (
	// account

	AccountCmd = cli.Command{
		Name:   "account",
		Usage:  "show a refactoring account demo",
		ArgsUsage: "",
		Category:"account",
		Description: `
		null `,
		Subcommands: []cli.Command{
			TransferDemoCmd,
			RegisterRoleCmd,
		},
		// HideHelp: false,
	}

	TransferDemoCmd = cli.Command{
		Name:   "transferDemo",
		Usage:  "show a refactoring transfer demo",
		ArgsUsage: "<to>",
		Action: utils.MigrateFlags(transferDemo),
		// Category:"account",
		Flags:	[]cli.Flag{
			TransferValueFlag,
			AccountCmdFlags,
			GasCmdFlags,
			GasPriceCmdFlags,
			LocalCmdFlags,
			KeystoreCmdFlags,
			SyncCmdFlags,
		},
		Description: `
		null `,
	}

	RegisterRoleCmd = cli.Command{
		Name:   "register-role",
		Usage:  "show a refactoring role registration demo",
		ArgsUsage: "<roles>",
		Action: utils.MigrateFlags(registerRole),
		// Category:"account",
		Flags: globalCmdFlags,
			/*
			[]cli.Flag{
			RolesFlag,
			AccountCmdFlags,
			GasCmdFlags,
			GasPriceCmdFlags,
			SignCmdFlags,
			SyncCmdFlags,
		},*/
	}

//------------------------cli Command------------------------------------------------

// admin

	AdminCmd = cli.Command{
		Name:   "admin",
		Usage:  "show a refactoring admin demo",

		Subcommands: []cli.Command{
			NodeCmd,
		},
	}

	NodeCmd = cli.Command{
		Name:   "node",
		Usage:  "show a refactoring node demo",

		Subcommands: []cli.Command{
			NodeAddCmd,
		},
	}

	NodeAddCmd = cli.Command{
		Name:   "add",
		Usage:  "show a refactoring add node demo",
		Action: nodeAdd,
		Flags:  tmpCmdFlags,
	}

	// contract

	ContractCmd = cli.Command{
		Name:   "contract",
		Usage:  "show a refactoring contract demo",
		ArgsUsage: "",
		Category:"contract",
		Subcommands: []cli.Command{
			ExecuteDemoCmd,
			// DeployDemoCmd,
			{
				Name:   "deployDemo",
				Usage:  "show a refactoring deploy demo",
				ArgsUsage: "<codeFile>",
				Action: deployDemo,
				Category:"contract",
				Flags:  []cli.Flag{
					ContractAbiFilePathFlag,
					AccountCmdFlags,
					GasCmdFlags,
					GasPriceCmdFlags,
					LocalCmdFlags,
					KeystoreCmdFlags,
					SyncCmdFlags,
				},
			},
		},
	}

	ExecuteDemoCmd = cli.Command{
		Name:   "executeDemo",
		Usage:  "show a refactoring invoke/cnsInvoke demo",
		ArgsUsage: "<contract> <function>",
		Action: execute,
		Category:"contract",
		Flags:	[]cli.Flag{
			ContractAbiFilePathFlag,
			ContractParamFlag,
			AccountCmdFlags,
			GasCmdFlags,
			GasPriceCmdFlags,
			LocalCmdFlags,
			KeystoreCmdFlags,
			SyncCmdFlags,
		},
	}

	/*
	DeployDemoCmd = cli.Command{
		Name:   "deployDemo",
		Usage:  "show a refactoring deploy demo",
		ArgsUsage: "<abiFile> <codeFile>",
		Action: deployDemo,
		Category:"contract",
		Flags:  globalCmdFlags,
	}*/

	// cns

	CnsCmd = cli.Command{
		Name:   "cns",
		Usage:  "show a refactoring cns demo",
		Category:"cns",
		Subcommands: []cli.Command{
			CnsResolveCmd,
		},
	}

	CnsResolveCmd = cli.Command{
		Name:   "resolve",
		Usage:  "show a refactoring cns resolve demo",
		ArgsUsage: "<name>",
		Action: resolve,
		Flags:  globalCmdFlags,
	}


	// fire wall

	FwCmd = cli.Command{
		Name:   "fw",
		Usage:  "show a refactoring fw demo",
		Category:"fw",
		//Action: fwStatus,
		//Flags:  fwCmdFlags,
		Subcommands: []cli.Command{
			FwStatusCmd,
			FwStartCmd,
			FwExportCmd,
			FwNewCmd,
		},
	}

	FwStartCmd = cli.Command{
		Name:   "start",
		Usage:  "show a refactoring fw start demo",
		ArgsUsage: "<address>",
		Action: fwStart,
		Flags:  globalCmdFlags,
	}

	FwStatusCmd = cli.Command{
		Name:   "status",
		Usage:  "show a refactoring fw status demo",
		ArgsUsage: "<address>",
		Action: fwStatus,
		Flags:  globalCmdFlags,
	}

	FwExportCmd = cli.Command{
		Name:   "export",
		Usage:  "show a refactoring fw export demo",
		ArgsUsage: "<address>",
		Action: fwStatus,
		Flags:  globalCmdFlags,
	}

	FwNewCmd = cli.Command{
		Name:   "new",
		Usage:  "show a refactoring fw new demo",
		ArgsUsage: "<action> <address> <api>",
		Action: fwNew,
		Flags:  globalCmdFlags,
	}

)

//------------------------function------------------------------------------------

func deployDemo (c *cli.Context) error{

	var err error
	var data string

	// 内置
	txType := deployContract
	// 必选参数
	codePath := c.Args().First()
	abiPath :=  c.String("abi")
	// 可选参数 （无）

	// 全局参数
	Address, Gas, GasPrice, keystore, _, isLocal := getGlobalParam(c)

	/*
	if len(c.Args()) != 2 {
		utils.Fatalf("param check error, required %d inputs, recieved %d\n",2, len(c.Args()))
	}*/

	codeBytes := paramParse(codePath, "code").([]byte)
	//fmt.Printf("codeBytes is %s", codeBytes)

	// abiBytes := paramParse(abiPath,"abi").([]byte)

	if isWasmContract(codeBytes){
		abiBytes := paramParse(abiPath,"abi").([]byte)

		// data组包
		data, err = combineData(nil,Int64ToBytes(int64(txType)),codeBytes,abiBytes)
	}else{
		// data组包
		data = "0x" + string(codeBytes[:len(codeBytes)-1])
		//data, err = combineData(nil,codeBytes)
	}

	if err != nil {
		utils.Fatalf("packet data error, %s", err.Error())
	}

	// tx组包
	err = combineTx(Address,nil,"",Gas,GasPrice,data,txType,"",keystore, true , isLocal, true)
	if err != nil {
		utils.Fatalf("packet Transaction error, %s", err.Error())
	}

	return nil
}

func transferDemo (c *cli.Context) error{
	var err error

	// 内置
	txType := transfer
	// 必选参数
	to := c.Args().First()
	// 可选参数
	value :=  c.String("value")
	// 全局参数
	Address, Gas, GasPrice, keystore, _, isLocal := getGlobalParam(c)

	// 检查
	if to == "" {
		fmt.Printf("\nthe input <to> can't be empty!\n")
		return nil
	}

	toNew := paramParse(to,"to").(common.Address)

	Value := paramParse(value, "value").(string)


	err = combineTx(Address,&toNew,Value,Gas,GasPrice,"",txType,"", keystore,true, isLocal, false)
	if err != nil {
		fmt.Printf("packet Transaction error, %s", err.Error())
	}

	return nil
}

func contractCommon(c *cli.Context, funcParams []string, funcName, contract string) error{

	// 内置常量
	to := DefaultAccountAdress
	str := contract

	// 可选参数
	abiPath := c.String("abi")

	// 全局参数
	Address, Gas, GasPrice, keystore, isSync, isLocal := getGlobalParam(c)

	txType := paramParse(contract, "contract").(int)

	abiBytes := abiParse(abiPath, contract) //有问题

	// func参数组提取
	var funcParamsNew []string
	funcName, funcParamsNew = funcParamsParse(funcName)
	funcParams = append(funcParams, funcParamsNew...)

	// func参数组包与格式转换
	outputType, isWrite, funcBytes, err := combineFunc(funcName, funcParams, abiBytes)
	if err != nil {
		fmt.Printf("\npacket functions err, %s\n", err.Error())
		return nil
	}

	if txType == invokeContract { //补丁
		str = ""
		to = contract
	}

	toNew := paramParse(to,"to").(common.Address)

	// data组包
	data, e:= combineData(funcBytes,Int64ToBytes(int64(txType)),[]byte(str))
	if e != nil {
		fmt.Printf("\npacket data err, %s\n", err.Error())
	}

	// tx组包
	err = combineTx(Address,&toNew,"",Gas,GasPrice,data,txType,outputType, keystore, isWrite, isLocal, isSync)
	if err != nil {
		fmt.Printf("\npacket combineTx err, %s\n", err.Error())
	}

	return nil
}


func execute (c *cli.Context) error {

	contract := c.Args().First()
	funcName := c.Args().Get(1)
	funcParams := c.StringSlice("param")

	if len(c.Args()) != 2 {
		utils.Fatalf("param check error, required %d inputs, recieved %d\n",2, len(c.Args()))
	}

	_ = contractCommon(c, funcParams, funcName, contract)

	return nil

}


func registerRole (c *cli.Context) error{

	roles := c.Args().First()

	if roles == ""{
		utils.Fatalf("the input <roles> cannot be empty")
	}

	funcParams := []string{
		roles,
	}

	_ = contractCommon(c, funcParams, "registerRole", "__sys_RoleRegister")

	return nil
}


/*
func migrate () error{
	// 内置
	funcName := "migrateFrom"
	txType := migTxType
	// 必选参数
	funcParams := c.String("addr") //输入不为空 检查是否为地址
	funcParams += c.String("to")

	// abiPath :=  c.String("abi")
	// 可选参数

	// 全局参数
	address := c.String("account")
	gas := c.String("gas")
	gasPrice := c.String("gasPrice")

}*/


func nodeAdd (c *cli.Context) error {

	// 必选参数 组合成json形式
	var str string
	var strMustArray = []string{"name:", "publicKey:", "desc:", "externalIP:", "internalIP:",}
	var strArray = []string{"status:1,",  "type:0,"}
	//var strd = []string{"rpcPort:", "p2pPort:"}
	//var intDefaltArray = []string
	//vat strOptionalArray = []string{}

	if len(c.Args()) != len(strMustArray){
		utils.Fatalf("param check error, required %d inputs, recieved %d\n",len(strArray),len(c.Args()))
	}

	// 可选参数,但必填
	// combineJson()

	// combineJson(c.)
	str = "{"		// start
	for i, data := range strArray {
		tmp := c.Args().Get(i)
		paramParse(tmp, data)
		str += data + tmp + ","
	}


	str += "}"		//end

	funcParams := []string{
		str,
	}

	_ = contractCommon(c, funcParams, "add", "__sys_NodeManager")

	return nil

}


func resolve (c *cli.Context) error {

	name := c.Args().First()
	ver := "latest"

	if name == "" {
		utils.Fatalf("the input <name> cannot be empty\n")
	}

	funcParams := []string{
		name,
		ver,
	}

	// __sys_CnsManager不能进行名字访问 未注册
	_ = contractCommon(c, funcParams, "getContractAddress", cnsProxyAddress)

	return nil

}


func fw (c *cli.Context, funcName string, funcParams []string, isSync bool) {

	txType := fwTxType

	/*
	if addr == "" {
		utils.Fatalf("the input <address> cannot be empty")
	}*/

	addr := c.Args().First()
	to := paramParse(addr,"fw address").(common.Address)

	Address, Gas, GasPrice, keystore, isSync, isLocal := getGlobalParam(c)

	outputType, isWrite, funcBytes, err := combineFunc(funcName, funcParams, nil)

	data, e:= combineData(funcBytes, Int64ToBytes(int64(txType)))
	//data = append(data, funcBytes)
	if e != nil {
		panic(fmt.Errorf("XXX error, %s", e.Error()))
	}

	err = combineTx(Address,&to,"",Gas,GasPrice,data,txType,outputType, keystore, isWrite, isLocal, isSync)
	if err != nil {
		panic(fmt.Errorf("XXX error, %s", err.Error()))
	}

}


func fwStart (c *cli.Context) {
	funcName := "__sys_FwOpen"
	isSync := c.Bool("sync")
	//addr := c.Args().First()

	//_ = paramParse(addr,"address")

	fw(c, funcName, []string{}, isSync)

	return
}


func fwStatus (c *cli.Context) {
	funcName := "__sys_FwStatus"
	fw(c, funcName, []string{},true)

	return
}


/*
func fwExport (c *cli.Context) {
	funcName := "__sys_FwStatus()"
	fw(c, funcName, []string{},true)

	return
}*/

func fwNew (c *cli.Context) {
	funcName := "__sys_FwAdd"
	action := ""

	addr := ""
	api := ""

	rules := combineRule(addr, api)

	funcParams := []string{
		action,
		rules,
	}

	fw(c, funcName, funcParams,true)

	return
}

/*
func fwExport (c *cli.Context) {
	funcName := "__sys_FwStatus()"
	fw(c, funcName, []string{},true)

	return
}*/

//------------------------param Check------------------------------------

func getGlobalParam(c *cli.Context) (string, string, string, string, bool, bool) {
	// 全局参数
	address := c.String("account")
	gas := c.String("gas")
	gasPrice := c.String("gasPrice")
	keystore := c.String("keystore")
	isSync := c.Bool("sync")
	local := c.Bool("local")

	Address, Gas, GasPrice, Keystore, isLocal := globalParamCheck(address,gas,gasPrice,"", keystore, local)

	return Address, Gas, GasPrice, Keystore, isSync, isLocal
}

func globalParamCheck(account, gas, gasPrice, url, keystore string, isLocal bool) (string, string, string, string, bool) {

	var configPath = DefaultConfigPath
	// Account := common.Address{}


	if url == "" {
		_ = parseConfigJson(configPath)
	} else {
		// isUrlFormat(url)
	}

	defaultAccount := config.From //test use only

	if account != "" {
		paramParse(account, "address")
		// Account = common.Address{[]byte(account)}
	} else{
		account = defaultAccount
	}

	/*
	AccountTest := common.HexToAddress("")
	fmt.Printf("%v, %v", Account, AccountTest)
	if Account == AccountTest {fmt.Printf("the same")}
	fmt.Printf("===============================================")*/

	if gas == ""{
		gas = "0x999999"
	}
	Gas := paramParse(gas, "gas").(string)

	if gasPrice == ""{
		gasPrice = "0x8250de00"
	}
	GasPrice := paramParse(gasPrice, "gasPrice").(string)

	// temp
	if isLocal && keystore == ""{
		keystore = DefaultPath
	}
	if keystore != ""{
		isLocal = true
	}

	return account, Gas, GasPrice, keystore, isLocal
}


func paramParse(param, paramName string) interface{} {
	var err error
	var i interface{}

	if param == ""{
		utils.Fatalf("the input <%s> cannot be empty!\n", paramName)
	}

	switch paramName{
		case "address":
			_, err = isAddressFormat(param)
		case "contract":
			i, _ = isAddressFormat(param)
		case "value", "gasPrice":
			i, err = isValueCheck(param, true) // 转换
		case "gas":
			i, err = isValueCheck(param, false) // 转换
		case "action":
			// check =
		case "fw address", "to":
			_, err = isAddressFormat(param)
			i = common.HexToAddress(param)
		case "code", "abi":
			i, err = parseFileToBytesDemo(param)
	}

	if err != nil {
		utils.Fatalf("%s param parse error, %s",paramName, err.Error())
	}

	//return i, check
	return i

}


func abiParse(abiFilePath, str string) []byte {
	var err error
	var abiBytes []byte

	if abiFilePath != ""{
		abiBytes, err = parseFileToBytesDemo(abiFilePath)
		if err != nil{
			utils.Fatalf("parse abi file from --abi failed, %s", err.Error())
		}

	} else {
		abiFilePath = getAbiFileFromLocal(str)
		abiBytes, err = parseFileToBytesDemo(abiFilePath)
		if err != nil{
			abiBytes = getAbiOnchain(str)
		}
	}

	return abiBytes
}


func getAbiFileFromLocal(str string) string {

	if str == cnsProxyAddress { // 补丁
		str = "__sys_CnsManager"
	}

	if strings.HasPrefix(str, "__sys_") {
		abiFilePath := DefaultContractPath + strings.ToLower(str[6:7]) + str[7:] + ".cpp.abi.json"
		return abiFilePath
	}

	return ""
}


func getAbiOnchain(addr string) []byte{
	//var tmp [][]byte
	var abiBytes []byte

	if !strings.HasPrefix(addr,"0x") {
		addr = getAddressByName(addr)
	}

	code, err := getCodeByAddress(addr)
	if err != nil{
		utils.Fatalf("Get abi data from chain error, %s",err.Error())
	}

	//TODO EVM Contract

	abiBytes,_ = hexutil.Decode(code)

	_, abiBytes,_, err = common.ParseWasmCodeRlpData(abiBytes)
	if err != nil{
		utils.Fatalf("Parse abi from chain error, %s",err.Error())
	}
	// _ = rlp.DecodeBytes(abiBytes, &tmp)
	// abiBytes = tmp[2]

	return abiBytes
}


func isHex(str string) bool {
	// strconv.ParseInt(str, 0, ) 是否会越界?
	// [a-f] 正则表达式
	// 引入以太坊检查机制 - future work
	str = strings.ToLower(str)

	for _, ch := range str{
		if ch<=47 || (ch>=58&&ch<=96) || ch>=103 {
			return false
		}
	}

	return true
}


func isAddressFormat(str string) (int, error) {
	if !strings.HasPrefix(str, "0x")  {
		// fmt.Errorf("the address does not have prefix '0x'\n")
		return cnsTxType, fmt.Errorf("the address does not have prefix '0x'\n")
	}

	if len(str) != 42 {
		// fmt.Printf("the address length is incorrect\n")
		return cnsTxType, fmt.Errorf("the address length %d is incorrect\n", len(str)-2)
	}

	if !isHex(str[2:]){
		// fmt.Printf("the address is not in Hex format\n")
		return cnsTxType, fmt.Errorf("the address is not in Hex format\n")
	}

	return invokeContract,nil
}


func isValueCheck(value string, isInt bool) (string, error) {
	var err error
	var Value string

	if isInt {
		Value, err = isIntCheck(value)
	}else {
		Value, err = isUintCheck(value)
	}

	return Value, err
}


func isIntCheck(value string) (string, error){
	var err error
	var intValue int64

	if !strings.HasPrefix(value, "0x") {
		intValue, err = strconv.ParseInt(value, 10, 64)
	} else {
		intValue, err = strconv.ParseInt(value, 0, 64)
	}
	value = hexutil.EncodeBig(big.NewInt(intValue))

	return value, err
}


func isUintCheck(value string) (string, error){
	var err error
	var intValue uint64

	if !strings.HasPrefix(value, "0x") {
		intValue, err = strconv.ParseUint(value, 10, 64)
	} else {
		intValue, err = strconv.ParseUint(value, 0, 64)
	}
	value = hexutil.EncodeUint64(intValue)

	return value, err
}


// ----------------------------------------------------------------

func getCodeByAddress(addr string) (string, error) {

	params := []string{addr, "latest"}
	r, err := Send(params, "eth_getCode")
	if err != nil {
		return "", fmt.Errorf("send http post to get contract address error ")
	}

	var resp = Response{}
	err = json.Unmarshal([]byte(r), &resp)
	if err != nil {
		return "", fmt.Errorf("parse eth_getCode result error ! \n %s", err.Error())
	}

	if resp.Error.Code != 0 {
		return "", fmt.Errorf("eth_getCode error ,error: %v", resp.Error.Message)
	}

	return resp.Result, nil
}


//--------------------------------------------------------------------

func funcParamsParse(funcName string) (string,[]string) {

	var funcParams []string

	/*
		if funcName == ""{
			return "", nil, fmt.Errorf("function name can not be empty")
		}*/

	hasBracket := strings.Contains(funcName, "(") && strings.Contains(funcName, ")")
	if hasBracket {
		funcName, funcParams = GetFuncNameAndParams(funcName)
	}

	return funcName, funcParams
}


func parseFuncFromAbiDemo(abiBytes []byte, funcName string) (*FuncDesc, error) {
	funcs, err := parseAbiFromJsonDemo(abiBytes)
	if err != nil {
		return nil, err
	}

	for _, value := range funcs {
		if value.Name == funcName {
			return &value, nil
		}
	}
	return nil, fmt.Errorf("function %s not found in %s", funcName, abiBytes)
}


func parseAbiFromJsonDemo(abiBytes []byte) ([]FuncDesc, error) {

	var a []FuncDesc
	if err := json.Unmarshal(abiBytes, &a); err != nil {
		return nil, fmt.Errorf("parse abi to json error: %s", err.Error())
	}
	return a, nil
}


//-----------------------------------------------------------------------------
/*
func combineJson(array []string, arg...){
	strDefaltArray = var tmp string
	if tmp = c.String("rpc-port"); tmp == ""{
		tmp = "6791"
	}
	str += "\"rpcPort\":" + tmp + ","
	strDefaltArray = append(strDefaltArray,str)

	if tmp = c.String("p2p-port"); tmp == ""{
		tmp = "16791"
	}
	str += "," + "\"p2pPort\":" + tmp


	// 可选参数,fei必填
	if tmp = c.String("delay"); tmp == ""{

	} else {
		str += "," + "\"delay\":" + tmp
	}
}*/


func combineRule(addr, api string) string{
	return addr + ":" + api
}


func combineFunc(funcName string, inputParams []string, abiBytes []byte) (string,bool,[][]byte, error) {
	var p []byte
	var outputType string

	if abiBytes != nil {

		//Judging whether this method exists or not
		abiFunc, err := parseFuncFromAbiDemo(abiBytes, funcName) //修改
		if err != nil {
			return "", false, nil, err
		}

		if len(abiFunc.Inputs) != len(inputParams) {
			return "", false, nil, fmt.Errorf("incorrect number of parameters ,request=%d,get=%d\n", len(abiFunc.Inputs), len(inputParams))
		}

		funcByte := [][]byte{
			[]byte(funcName),
		}

		for i, v := range inputParams {
			input := abiFunc.Inputs[i]
			p, err = StringConverter(v, input.Type)
			if err != nil {
				return "", false, nil, fmt.Errorf("incorrect param type: %s,index: %d", v, i)
			}
			funcByte = append(funcByte, p)
		}

		//TODO EVM CONSTANT

		// hasFuncOutput := len(abiFunc.Outputs) != 0 && abiFunc.Outputs[0].Type != "void"
		if len(abiFunc.Outputs) != 0{
			outputType = abiFunc.Outputs[0].Type
		}
		//outputType := abiFunc.Outputs[0].Type
		isWrite := abiFunc.Constant == "false"

		return outputType, isWrite, funcByte, nil

	}else{

		funcByte := [][]byte{
			[]byte(funcName),
		}

		for _, input := range inputParams {
			funcByte = append(funcByte, []byte(input))
		}

		if funcName == "__sys_FwStatus" {
			return "string", false, funcByte, nil
		}

		return "", true, funcByte, nil
	}
}


func combineData(argFunc [][]byte, args ...[]byte) (string, error) {
	dataParams := make([][]byte, 0)

	for _, data := range args {

		if string(data) != "" {
			dataParams = append(dataParams, data)
		} else{
			continue
		}
		// s[i] = data
	}

	// apend function params (funcname and params) to data
	if argFunc != nil{
		/*
		for _, funcParam := range argFunc {
			dataParams = append(dataParams, funcParam)
		}*/
		dataParams = append(dataParams, argFunc...)
	}

	dataParamsRlp, err := rlp.EncodeToBytes(dataParams)
	if err != nil {
		return "", fmt.Errorf("rlp encode error,%s", err.Error())
	}

	return hexutil.Encode(dataParamsRlp), nil
}

func sendParams(action string, args... interface{}) *Response {
	params := make([]interface{},0)

	for _, param := range args{
		params = append(params, param)
	}

	res, _ := Send(params, action)
	response := parseResponse(res)

	return response
}

func getNonce(addr common.Address) uint64 {
	//params := make([]interface{},0)
	//params = append(params, addr)
	//params = append(params, "latest")

	//action := "eth_getTransactionCount"

	response := sendParams("eth_getTransactionCount", addr, "latest")

	nonce, _ := hexutil.DecodeBig(response.Result)
	fmt.Println(addr, nonce)

	return nonce.Uint64()
}

func getAddressByName(name string) string {

	_, _, funcBytes, _ := combineFunc("getContractAddress", []string{name, "latest"}, nil)

	data, _:= combineData(funcBytes, Int64ToBytes(int64(cnsTxType)))

	to := common.HexToAddress(cnsProxyAddress)

	tx := TxParams{
		To:       	&to,
		Data:     	data,
		TxType:   	cnsTxType,
	}

	response := sendParams("eth_call", tx, "latest")
	bytes, _ := hexutil.Decode(response.Result)
	result := BytesConverter(bytes, "string")

	return result.(string)
}

func combineTx(from string, to *common.Address, value, gas, gasPrice, data string, txType int, outputType string, keystore string, isWrite, isLocal, isSync bool) error {

	var action string
	var txSign *types.Transaction

	From := common.HexToAddress(from)

	tx := TxParams{
		From:     From,
		To:       to,
		GasPrice: gasPrice,
		Gas:      gas,
		Value:    value,
		Data:     data,
		TxType:   txType,
	}

	params := make([]interface{}, 0)

	switch {
		case !isWrite:
			params = append(params, tx)
			params = append(params, "latest")
			action = "eth_call"

		case isLocal:
			nonce := getNonce(From)
			priv := getPrivateKey(from, keystore)
			Value,_ := hexutil.DecodeBig(value)
			Gas,_ := hexutil.DecodeUint64(gas)
			GasPrice,_ := hexutil.DecodeBig(gasPrice)
			Data, _ := hexutil.Decode(data)

			if txType == deployContract {
				txSign = types.NewContractCreation(nonce, Value, Gas, GasPrice, Data)
			}else{
				txSign = types.NewTransaction(nonce, *to, Value, Gas, GasPrice, Data, uint64(txType))
			}

			txNew, _ := types.SignTx(txSign, types.HomesteadSigner{}, priv)

			bytes, _ := rlp.EncodeToBytes(txNew)

			params = append(params, hexutil.Encode(bytes))
			action = "eth_sendRawTransaction"

		default:
			params = append(params, tx)
			action = "eth_sendTransaction"
	}

	paramJson, _ := json.Marshal(params)
	fmt.Printf("\n request json data: %s \n", string(paramJson))


	r, err := Send(params, action)

	resp := parseResponse(r)
	fmt.Printf("\n response json: %s \n", r)

	getResponse(resp, outputType, isWrite, isSync)

	return err


	// if action == "" {}
	// return nil

}


func getResponse(resp *Response, outputType string, isWrite, isSync bool) {

	// 同步等待
	// fmt.Printf("\ntrasaction hash: %s\n", resp.Result)

	if !isWrite {
		if outputType != "" {
			bytes, _ := hexutil.Decode(resp.Result)

			// hex.DecodeString(strings.TrimPrefix(resp.Result, "0x"))

			result := BytesConverter(bytes, outputType)
			fmt.Printf("\nresult: %v\n", result)
			// return nil
		} else {
			fmt.Printf("\nresult: null\n")
		}
	} else {
		fmt.Printf("\ntrasaction hash: %s\n", resp.Result)

		if isSync {
			// 异步阻塞等待
			ch := make(chan string, 1)
			go GetTransactionReceiptDemo(resp.Result, ch)

			select {
			case str := <-ch:
				fmt.Printf("\n result: %s\n", str)
			case <-time.After(time.Second * 10):
				fmt.Printf("\n get contract receipt timeout...more than 200 second.\n")
			}
		}
	}
}


func GetTransactionReceiptDemo(txHash string, ch chan string) {
	var receipt = Receipt{}
	var str string

	for {
		res, _ := Send([]string{txHash}, "eth_getTransactionReceipt")

		e := json.Unmarshal([]byte(res), &receipt)
		if e != nil {
			utils.Fatalf("parse get receipt result error ! \n %s", e.Error())
		}

		if receipt.Result.Status == TxReceiptStatus_Failed{
			ch<- "Operation Failed"
			break
		}

		str = receipt.Result.ContractAddress
		if str != "" {
			ch<-"contract address is " + str
			break
		}

		log := receipt.Result.Logs
		if len(log) != 0 {

			tmp, _ := hexutil.Decode(log[0].Data)
			str = string(tmp)
			index := strings.Index(str, "ERR")

			ch<- str[index:]
			break
		}

		// 可否不要if判断？
		if receipt.Result.Status == TxReceiptStatus_Success{
			ch<- "Operation Succeeded"
			break
		}

	}
}


// uils?
/*
func signTx () error {
	message := getMessage(ctx, 1)

	// Load the keyfile.
	keyfilepath := ctx.Args().First()
	keyjson, err := ioutil.ReadFile(keyfilepath)
	if err != nil {
		utils.Fatalf("Failed to read the keyfile at '%s': %v", keyfilepath, err)
	}

	// Decrypt key with passphrase.
	passphrase := getPassphrase(ctx)
	key, err := keystore.DecryptKey(keyjson, passphrase)
	if err != nil {
		utils.Fatalf("Error decrypting key: %v", err)
	}

	signature, err := crypto.Sign(signHash(message), key.PrivateKey)
	if err != nil {
		utils.Fatalf("Failed to sign message: %v", err)
	}
	out := outputSign{Signature: hex.EncodeToString(signature)}
	if ctx.Bool(jsonFlag.Name) {
		mustPrintJSON(out)
	} else {
		fmt.Println("Signature:", out.Signature)
	}
	return nil
}*/



