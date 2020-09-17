package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/packet"
	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/gin-gonic/gin"
)

const defaultEndPoint = "http://127.0.0.1:6791"

var (
	errExceedQueryKeyLimit = errors.New("please provide at most ONE query key")
	errInvalidParam        = errors.New("invalid params")
	errPollingReceipt      = errors.New("polling transaction receipt error")
)

// ===================== Input Json Obj ===========================
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

// ===================== Handler COMMON =======================
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

	ctx.JSON(200, jsonStringPatch(res[0]))
}

func jsonStringPatch(value interface{}) interface{} {
	if reflect.ValueOf(value).Kind() == reflect.String {
		str := value.(string)
		// string starts with {"
		if bytes.Equal([]byte(str)[:2], []byte{123, 34}) {
			var m map[string]interface{}
			json.Unmarshal([]byte(value.(string)), &m)
			return m
		}
	}

	return value
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

	ctx.JSON(200, res[0])
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

	contractAbi, _ := packet.ParseAbiFromJson(funcAbi)
	methodAbi, _ := contractAbi.GetFuncFromAbi(jsonInfo.Contract.Method)
	funcArgs, _ := methodAbi.StringToArgs(funcParams)

	cns, to, err := cmd_common.CnsParse(jsonInfo.Contract.ContractAddr)
	if err != nil {
		return nil, err
	}

	// todo: lack of virtual machine interpreter
	vm := jsonInfo.Contract.Interpreter

	data := packet.NewData(funcArgs, methodAbi)
	dataGenerator := packet.NewContractDataGen(data, contractAbi, cns.TxType)
	dataGenerator.SetInterpreter(vm, cns.Name, cns.TxType)

	// todo: lack of tx sender
	from := common.HexToAddress(jsonInfo.Tx.From)
	tx := packet.NewTxParams(from, &to, "", jsonInfo.Tx.Gas, "", "")

	keyfile := parseKeyfile(jsonInfo.Tx.From)
	keyfile.Passphrase = jsonInfo.Rpc.Passphrase

	return A(jsonInfo.Rpc.EndPoint, dataGenerator, tx, keyfile)
}

func A(url string, dataGen packet.MsgDataGen, tx *packet.TxParams, keyfile *utils.Keyfile) ([]interface{}, error) {
	pc, err := platoneclient.SetupClient(url)
	if err != nil {
		return nil, err
	}

	return pc.MessageCallV2(dataGen, tx, keyfile, true)
}

func parseKeyfile(from string) *utils.Keyfile {
	fileName := utils.GetFileByKey(defaultKeyfile, from)

	if fileName != "" {
		path := defaultKeyfile + "/" + fileName
		keyfile, _ := utils.NewKeyfile(path)
		return keyfile
	}

	return &utils.Keyfile{}
}
