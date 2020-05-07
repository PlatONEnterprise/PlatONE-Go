/*
utils for packet Transactions
*/

package packet

import (
	"bytes"
	"encoding/json"
	"fmt"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"strings"
	"time"
)

const (
	// Transaction types
	TRANSFER         = 0
	DEPLOY_CONTRACT  = 1
	EXECUTE_CONTRACT = 2
	CNS_TX_TYPE      = 0x11 // Used for sending transactions without address
	FW_TX_TYPE       = 0x12 // Used fot sending transactions  about firewall
	MIG_TX_TYPE      = 0x13 // Used for update system contract.
	MIG_DP_TYPE      = 0x14 // Used for update system contract without create a new contract manually
	// Currently it's under developing

	DEFAULT_SYSTEM_CONTRACT_PATH = "../../release/linux/conf/contracts/"

	CNS_PROXY_ADDRESS = "0x0000000000000000000000000000000000000011"

	TX_RECEIPT_STATUS_SUCCESS = "0x1"
	TX_RECEIPT_STATUS_FAILURE = "0x0"

	SLEEP_TIME = 1000000000 // 1 seconds
)

// TxParamsDemo, the object of the eth_call, eth_sendTransaction
type TxParams struct {
	From     common.Address  `json:"from"` // the address used to send the transaction
	To       *common.Address `json:"to"`   // the address receives the transactions
	Gas      string          `json:"gas"`
	GasPrice string          `json:"gasPrice"`
	Value    string          `json:"value"`
	Data     string          `json:"data"`
	TxType   int             `json:"txType"`
}

// ContractReturn, system contract return object
type ContractReturn struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// FuncDesc, the object of the contract abi files
type FuncDesc struct {
	Name   string `json:"name"`
	Inputs []struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		InternalType string `json:"internalType,omitempty"`
	} `json:"inputs"`
	Outputs []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"outputs"`
	Constant        interface{} `json:"constant"` // ???
	Type            string      `json:"type"`
	StateMutability string      `json:"stateMutability,omitempty"` // tag for solidity ver > 0.6.0
}

// Cns,
type Cns struct {
	To     string
	Name   string // the cns name of contract
	txType int    // the transaction type of the contract execution (EXECUTE_CONTRACT or CNS_TX_TYPE)
}

// ParseFuncFromAbi searches the function names in the []FuncDesc object array
func ParseFuncFromAbi(abiBytes []byte, funcName string) (*FuncDesc, error) {
	funcs, err := ParseAbiFromJson(abiBytes)
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

// ParseAbiFromJson parses the application binary interface(abi) files to []FuncDesc object array
func ParseAbiFromJson(abiBytes []byte) ([]FuncDesc, error) {
	var a []FuncDesc
	if err := json.Unmarshal(abiBytes, &a); err != nil {
		return nil, fmt.Errorf(utl.ErrUnmarshalBytesFormat, "abi", err.Error())
	}
	return a, nil
}

// CnsParse judge whether the input string is contract address or contract name
// and return the corresponding infos
func CnsParse(contract string) *Cns {
	isAddress, _ := utl.IsNameOrAddress(contract)

	if isAddress {
		return NewCns(contract, "", EXECUTE_CONTRACT)
	} else {
		return NewCns("", contract, CNS_TX_TYPE)
	}
}

func NewCns(to, name string, txType int) *Cns {
	return &Cns{
		To:     to,
		Name:   name,
		txType: txType,
	}
}

// AbiParse gets the abi bytes by the input parameters provided
// The abi file can be obtained through following ways:
// 1. user provide the abi file path
// 2. get the abi files from default file locations (for example, the system contracts are
// all stored in ./PlatONE/release/linux/conf/contracts)
// 3. get the abi bytes on chain (wasm contract only).
func AbiParse(abiFilePath, str string) []byte {
	var err error
	var abiBytes []byte

	/*
	if abiFilePath != "" {
		abiBytes, err = utl.ParseFileToBytes(abiFilePath)
	} else {
		abiFilePath = getAbiFileFromLocal(str)
		if abiFilePath != "" {
			abiBytes, err = utl.ParseFileToBytes(abiFilePath)
		} else {
			abiBytes, err = getAbiOnchain(str)
			if err != nil {
				utils.Fatalf(utl.ErrGetFromChainFormat, "abi data", err.Error())
			}
		}
	}
	if err != nil {
		utils.Fatalf(utl.ErrParseFileFormat, "abi", err.Error())
	}*/

	if abiFilePath == "" {
		abiFilePath = getAbiFileFromLocal(str)
	}

	abiBytes, err = utl.ParseFileToBytes(abiFilePath)
	if err != nil {
		utils.Fatalf(utl.ErrParseFileFormat, "abi", err.Error())
	}

	return abiBytes
}

// getAbiFileFromLocal get the abi files from default directory by file name
// currently it is designed to get the system contract abi files
func getAbiFileFromLocal(str string) string {

	var abiFilePath string

	// (patch) convert CNS_PROXY_ADDRESS to cnsManager system contract
	if str == CNS_PROXY_ADDRESS {
		str = "__sys_CnsManager"
	}

	runPath := utl.GetRunningTimePath()

	if strings.HasPrefix(str, "__sys_") {
		sysFileName := strings.ToLower(str[6:7]) + str[7:] + ".cpp.abi.json"
		abiFilePath = runPath + DEFAULT_SYSTEM_CONTRACT_PATH + sysFileName
	}else{
		abiFilePath = GetAbiFile(str)
	}

	return abiFilePath
}

func GetAbiFile(key string) string {
	fileName := utl.GetFileByKey(utl.DEAFAULT_ABI_DIRT, key)

	if fileName != ""{
		return utl.DEAFAULT_ABI_DIRT +  "/" + fileName
	}

	return ""
}


// TODO
// getAbiFileFromLocal get the abi files from chain
// it is only available for wasm contracts
func getAbiOnchain(addr string) ([]byte, error) {
	var abiBytes []byte
	var err error

	utl.ParamValid(addr, "contract")

	// if the input parameter is a contract name, convert the name to address by executing cns
	if utl.IsMatch(addr, "name") {
		addr, err = GetAddressByName(addr)
		if err != nil {
			return nil, err
		}
	}

	// get the contract code by address through eth_getCode
	code, err := getCodeByAddress(addr)
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
}

// ExtractContractData extract the role info from the contract return result
func ExtractContractData(result, role string) string {
	var inter = make([]interface{}, 0)
	var count int

	r := ParseSysContractResult([]byte(result))
	data := r.Data.([]interface{})

	length := len(data)
	for i := 0; i < length; i++ {
		temp, _ := json.Marshal(data[0])
		if strings.Contains(string(temp), role) {
			inter = append(inter, data[i])
			count++
		}
	}

	if count == 0 {
		return fmt.Sprintf("no %s in registration\n", role)
	} else {
		r.Data = inter
		newContractData, _ := json.Marshal(r)
		return string(newContractData)
	}
}

// ParseSysContractResult parsed the result to ContractReturn object
func ParseSysContractResult(result []byte) *ContractReturn {
	a := ContractReturn{} //删除？
	err := json.Unmarshal(result, &a)
	if err != nil {
		utils.Fatalf(utl.ErrUnmarshalBytesFormat, "contract return", err.Error())
	}

	return &a
}

// ParseSysContractResult parsed the rpc response to Receipt object
func ParseTxReceipt(response interface{}) *Receipt {
	var receipt = &Receipt{}

	if response == nil{
		return nil
	}

	temp, _ := json.Marshal(response)
	err := json.Unmarshal(temp, receipt)
	if err != nil {
		utl.LogErr.Printf(utl.ErrUnmarshalBytesFormat, "transaction receipt", err.Error())
		utils.Fatalf(utl.ErrTODO, utl.DEFAULT_LOG_DIRT)
	}

	return receipt
}

// NewTxParams news a TxParams object
func NewTxParams(from common.Address, to *common.Address, value, gas, gasPrice, data string, txType int) *TxParams {

	tx := &TxParams{
		From:     from,
		To:       to,
		GasPrice: gasPrice,
		Gas:      gas,
		Value:    value,
		Data:     data,
		TxType:   txType,
	}

	return tx
}

// SendMode selects the rpc calls (eth_call, eth_sendTransaction, and eth_sendRawTransaction)
func (tx *TxParams) SendMode(isWrite bool, keystore string) ([]interface{}, string) {
	var action string
	var params = make([]interface{}, 0)

	switch {
	case !isWrite:
		params = append(params, tx)
		params = append(params, "latest")
		action = "eth_call"
	case keystore != "":
		signedTx := tx.GetSignedTx(keystore)
		params = append(params, signedTx)
		action = "eth_sendRawTransaction"
	default:
		params = append(params, tx)
		action = "eth_sendTransaction"
	}

	return params, action
}

// GetSignedTx gets the signed transaction
func (tx *TxParams) GetSignedTx(keystore string) string {

	var txSign *types.Transaction

	// convert the TxParams object to types.Transaction object
	nonce := GetNonce(tx.From)
	value, _ := hexutil.DecodeBig(tx.Value)
	gas, _ := hexutil.DecodeUint64(tx.Gas)
	gasPrice, _ := hexutil.DecodeBig(tx.GasPrice)
	data, _ := hexutil.Decode(tx.Data)
	txType := uint64(tx.TxType)

	if tx.To == nil {
		txSign = types.NewContractCreation(nonce, value, gas, gasPrice, data)
	} else {
		txSign = types.NewTransaction(nonce, *tx.To, value, gas, gasPrice, data, txType)
	}

	// extract pk from keystore file and sign the transaction
	priv := utl.GetPrivateKey(tx.From, keystore)
	txSign, _ = types.SignTx(txSign, types.HomesteadSigner{}, priv)
	utl.Logger.Printf("the signed transaction is %v\n", txSign)

	str, _ := rlpEncode(txSign)
	return str
}

// ParseTxResponse parse result based on the function constant and output type
// if the isSync is ture, the function will get the receipt of the transaction in further
func ParseTxResponse(resp interface{}, outputType string, isWrite, isSync bool) interface{} {

	var respStr string

	//TODO
	temp, _ := json.Marshal(resp)
	_ = json.Unmarshal(temp, &respStr)

	switch {
	case !isWrite:
		return ParseNonConstantRespose(respStr, outputType)
	case isSync:
		return GetResponseByReceipt(respStr)
	default:
		return fmt.Sprintf("trasaction hash is %s\n", respStr)
	}
}

// ParseNonConstantRespose wraps the utl.BytesConverter,
// it converts the hex string response based the output type provided
func ParseNonConstantRespose(respStr, outputType string) interface{} {
	if outputType != "" {
		b, _ := hexutil.Decode(respStr)
		bytesTrim := bytes.TrimRight(b, "\x00")
		return utl.BytesConverter(bytesTrim, outputType)
		//utl.Logger.Printf("result: %v\n", result)
	} else {
		return fmt.Sprintf("message call has no return value\n")
	}
}

func isSpecialChar(r rune) bool {

	if r >= 32 && r <= 126 { 			// ascii char
		return false
	}else if r >= 19968 && r <= 40869 {	// unicode \u4e00-\u9fa5
		return false
	}else{
		return true
	}
}

func TrimSpecialChar(trimRunes []rune) []rune{

	var newBytes = make([]rune, 0)

	for _, v := range trimRunes{
		if !isSpecialChar(v){
			newBytes = append(newBytes,v)
		}
	}

	return newBytes
}

// GetReceiptByPolling creates a channel to get the transaction receipt by polling
// The timeout is setted to 10 seconds
func GetResponseByReceipt(respStr string) interface{} {
	ch := make(chan string, 1)
	go GetReceiptByPolling(respStr, ch)

	select {
	case str := <-ch:
		runesTrim := TrimSpecialChar([]rune(str))
		str = string(runesTrim)
		utl.Logger.Printf("result: %s\n", str)
		return str
	case <-time.After(time.Second * 10):
		temp1 := fmt.Sprintf("\nget contract receipt timeout...more than 10 second.\n")
		temp2 := fmt.Sprintf("trasaction hash is %s\n", respStr)
		return temp1 + temp2
	}
}

// GetReceiptByPolling gets transaction receipt by polling. After getting the receipt, it
// parses the receipt and get the infos (contract address, transaction status, logs, etc.)
// The sleep time is designed to limit the times of the polling
func GetReceiptByPolling(txHash string, ch chan string) {

	for {
		receipt, err := GetTransactionReceipt(txHash)

		switch {
		case err != nil:
			fmt.Println(err.Error())
			fmt.Printf("try again 5s later...")
			time.Sleep(5*SLEEP_TIME)
			fmt.Printf("try again...\n")
			continue
		case receipt == nil:
			time.Sleep(2*SLEEP_TIME)
			continue
		}

		switch {
		case receipt.Status == TX_RECEIPT_STATUS_FAILURE:
			ch <- "Operation Failed"
			break
		case receipt.ContractAddress != "":
			ch <- receipt.ContractAddress
			break
		case len(receipt.Logs) != 0:
			tmp, _ := hexutil.Decode(receipt.Logs[0].Data)
			ch <- string(tmp)
			break
		case receipt.Status == TX_RECEIPT_STATUS_SUCCESS:
			ch <- "Operation Succeeded"
			break
		}

	}

}
