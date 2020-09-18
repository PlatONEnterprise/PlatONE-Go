package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/gin-gonic/gin"
)

func registerContractRouters(r *gin.Engine) {
	contract := r.Group("/contracts")
	{
		contract.POST("", deployHandler)           // deploy 	- resource: contract
		contract.PUT("/:address", migrateHandler)  // migrate	- resource: contract -> new contract
		contract.POST("/:address", executeHandler) // execute	- resource: methods of contract
	}
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
	/// files := form.File["files"]
	files := append(form.File["code"], form.File["abi"][0])

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

		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, res[0])
}

// todo: refactory
func deploy(jsonInfo *temp) ([]interface{}, error) {

	var consArgs = make([]interface{}, 0)
	var constructor *packet.FuncDesc

	vm := jsonInfo.Contract.Interpreter
	data, _ := getDataParams(jsonInfo.Contract.Data)
	codeBytes := []byte(data[0])
	abiBytes := []byte(data[1])
	consParams := data[2:]

	conAbi, _ := packet.ParseAbiFromJson(abiBytes)
	if constructor = conAbi.GetConstructor(); constructor != nil {
		consArgs, _ = constructor.StringToArgs(consParams)
	}

	dataGenerator := packet.NewDeployDataGen(conAbi, types.CreateTxType)
	dataGenerator.SetInterpreter(vm, abiBytes, codeBytes, consArgs, constructor)

	from := common.HexToAddress(jsonInfo.Tx.From)
	tx := packet.NewTxParams(from, nil, "", "", "", "")

	keyfile := parseKeyfile(jsonInfo.Tx.From)
	keyfile.Passphrase = jsonInfo.Rpc.Passphrase

	return A(jsonInfo.Rpc.EndPoint, dataGenerator, tx, keyfile)
}

// ===================== Migration =======================
func migrateHandler(ctx *gin.Context) {

}

// ===================== Execution =======================
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

		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(res) == 1 {
		ctx.JSON(200, res[0])
		return
	}

	ctx.JSON(200, res)
}
