// It is used for generate abi files in the default location

package main

import (
	"strings"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/precompiled"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
)

const (
	DEAFAULT_ABI_DIRT            = "./abi"
	DEFAULT_SYSTEM_CONTRACT_PATH = "../../release/linux/conf/contracts/"
)

var abiFileDirt string

func abiInit() {
	runPath := utl.GetRunningTimePath()
	abiFileDirt = runPath + DEAFAULT_ABI_DIRT

	utl.FileDirectoryInit(abiFileDirt)
	_ = utl.DeleteOldFile(abiFileDirt)
}

// getAbiFile gets the abi file that matches the keywords provided
func getAbiFile(key string) string {

	fileName := utl.GetFileByKey(abiFileDirt, key)

	if fileName != "" {
		return abiFileDirt + "/" + fileName
	}

	return ""
}

// storeAbiFile stores the abi files in the default abi file directory
// if the abi file is already exist, the abi file will not be stored
func storeAbiFile(key string, abiBytes []byte) {

	if len(abiBytes) == 0 {
		return
	}

	fileName := getAbiFile(key)
	if fileName == "" {
		filePath := abiFileDirt + "/" + key + ".abi.json"
		_ = utl.WriteFile(abiBytes, filePath)
	}
}

// todo: update the notation
// AbiParse gets the abi bytes by the input parameters provided
// The abi file can be obtained through following ways:
// 1. user provide the abi file path
// 2. get the abi files from default file locations (for example, the system contracts are
// all stored in ./PlatONE/release/linux/conf/contracts)
// 3. get the abi bytes on chain (wasm contract only).
func AbiParse(abiFilePath, str string) []byte {
	var err error
	var abiBytes []byte

	if abiFilePath == "" {
		if p := precompiledList[str]; p != "" { // todo: equalFold string?
			precompiledAbi, _ := precompile.Asset(p)
			return precompiledAbi
		}

		/// abiFilePath = getAbiFileFromLocal(str)
	}

	abiBytes, err = utl.ParseFileToBytes(abiFilePath)
	if err != nil {
		utils.Fatalf(utl.ErrParseFileFormat, "abi", err.Error())
	}

	return abiBytes
}

// getAbiFileFromLocal get the abi files from default directory by file name
// currently it is designed to get the system contract abi files
// [2020 - 07 -07] added, the method is deprecated, the system contract is moved to pre compiled contract
func getAbiFileFromLocal(str string) string {

	var abiFilePath string

	// (patch) convert CNS_PROXY_ADDRESS to cnsManager system contract
	if str == cnsManagementAddress {
		str = "__sys_CnsManager"
	}

	runPath := utl.GetRunningTimePath()

	if strings.HasPrefix(str, "__sys_") {
		sysFileName := strings.ToLower(str[6:7]) + str[7:] + ".cpp.abi.json"
		abiFilePath = runPath + DEFAULT_SYSTEM_CONTRACT_PATH + sysFileName
	} else {
		abiFilePath = getAbiFile(str)
	}

	return abiFilePath
}

// todo: deprecated?
/*
// getAbiOnchain get the abi files from chain
// it is only available for wasm contracts
func getAbiOnchain(addr string) ([]byte, error) {
	var abiBytes []byte
	var err error

	paramValid(addr, "contract")

	// if the input parameter is a contract name, convert the name to address by executing cns
	if utl.IsMatch(addr, "name") {
		addr, err = GetAddressByName(addr)
		if err != nil {
			return nil, err
		}
	}

	// get the contract code by address through eth_getCode
	code, err := platoneclient.GetCodeByAddress(addr)
	if err != nil {
		return nil, err
	}

	// parse the encoding contract code and get abi bytes
	abiBytes, _ = hexutil.Decode(code)
	_, abiBytes, _, err = common.ParseWasmCodeRlpData(abiBytes)
	if err != nil {
		return nil, fmt.Errorf(utl.ErrRlpDecodeFormat, "abi data", err.Error())
	}

	return abiBytes, nil
}*/

/*
// 2020.7.6 modified, moved from tx_call.go
// GetAddressByName wraps the RpcCalls used to get the contract address by cns name
// the parameters are packet into transaction before packet into rpc json data struct
func GetAddressByName(name string) (string, error) {

	// chain defined data type convert
	to := common.HexToAddress(cnsManagementAddress)
	from := common.HexToAddress("")

	// packet the contract all data
	rawData := packet.NewData("getContractAddress", []string{name, "latest"}, nil)
	call := packet.NewInnerCallDemo(rawData, types.NormalTxType)
	data, _, _, _ := call.CombineData()

	tx := packet.NewTxParams(from, &to, "", "", "", data)
	params := utl.CombineParams(tx, "latest")

	response, err := platoneclient.RpcCalls("eth_call", params)
	if err != nil {
		return "", err
	}

	// parse the rpc response
	resultBytes, _ := hexutil.Decode(response.(string))
	bytesTrim := bytes.TrimRight(resultBytes, "\x00")
	result := utl.BytesConverter(bytesTrim, "string")

	return result.(string), nil
}*/
