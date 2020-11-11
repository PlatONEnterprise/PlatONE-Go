package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

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
	contractAddr string
	Method       string `json:"method,omitempty"`
	Interpreter  string `json:"interpreter,omitempty"`
	abiMethods   []byte `json:"-"`
	/// Data         contractData `json:"data"`
	Data interface{} `json:"data"`
}

func newContractParams(defaultAddr, defaultMethod, defaultInter string, abiBytes []byte, dataParams interface{}) *contractParams {
	return &contractParams{
		contractAddr: defaultAddr,
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
	if data.Method == "getAllNodes" {
		//var nodes []NodeInfo
		jsonres := jsonStringPatch(res[0])
		//restring, ok := jsonres.(string)
		//if ok {
		//	json.Unmarshal([]byte(restring), &nodes)
		//}
		ctx.JSON(200, jsonres)
	} else {

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, jsonStringPatch(res[0]))
	}
}

func jsonStringPatch(value interface{}) interface{} {
	if reflect.ValueOf(value).Kind() == reflect.String {
		str := value.(string)
		// string starts with {"
		res := bytes.Trim([]byte(str), "\u0000")
		if bytes.Equal([]byte(str)[:2], []byte{123, 34}) {
			var m map[string]interface{}
			b := bytes.Trim([]byte(str), "\x00")
			//c := bytes.Trim([]byte(b), "\u0000")
			_ = json.Unmarshal(b, &m)
			return m
		}
		return string(res)
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

		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, res[0])
}

func handlerCallCommon(jsonInfo *temp) ([]interface{}, error) {
	var funcAbi []byte

	funcParams, _ := getDataParams(jsonInfo.Contract.Data)

	if p := precompile.List[jsonInfo.Contract.contractAddr]; p != "" {
		funcAbi, _ = precompile.Asset(p)
	} else {
		funcAbi = jsonInfo.Contract.abiMethods
	}

	contractAbi, _ := packet.ParseAbiFromJson(funcAbi)
	methodAbi, err := contractAbi.GetFuncFromAbi(jsonInfo.Contract.Method)
	if err != nil {
		return nil, err
	}
	funcArgs, _ := methodAbi.StringToArgs(funcParams)

	cns, to, err := cmd_common.CnsParse(jsonInfo.Contract.contractAddr)
	if err != nil {
		return nil, err
	}

	vm := jsonInfo.Contract.Interpreter

	data := packet.NewData(funcArgs, methodAbi)
	dataGenerator := packet.NewContractDataGen(data, contractAbi, cns.TxType)
	dataGenerator.SetInterpreter(vm, cns.Name, cns.TxType)

	from := common.HexToAddress(jsonInfo.Tx.From)
	tx := packet.NewTxParams(from, &to, "", jsonInfo.Tx.Gas, "", "")

	keyfile, err := parseKeyfile(jsonInfo.Tx.From)
	if err == nil {
		keyfile.Passphrase = jsonInfo.Rpc.Passphrase

		err := keyfile.ParsePrivateKey()
		if err != nil {
			return nil, err
		}
	}

	return A(jsonInfo.Rpc.EndPoint, dataGenerator, tx, keyfile)
}

func A(url string, dataGen packet.MsgDataGen, tx *packet.TxParams, keyfile *utils.Keyfile) ([]interface{}, error) {
	pc, err := platoneclient.SetupClient(url)
	if err != nil {
		return nil, err
	}

	return pc.MessageCallV2(dataGen, tx, keyfile, true)
}

func parseKeyfile(from string) (*utils.Keyfile, error) {
	if strings.HasPrefix(from, "0x") {
		from = from[2:]
	}
	fileName, err := utils.GetFileByKey(defaultKeyfile, from)
	if err != nil {
		return &utils.Keyfile{}, err
	}

	path := defaultKeyfile + "/" + fileName
	keyfile, _ := utils.NewKeyfile(path)
	return keyfile, nil
}
