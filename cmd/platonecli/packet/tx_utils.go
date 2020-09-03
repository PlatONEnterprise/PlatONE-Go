/*
utils for packet Transactions
*/

package packet

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
)

// ContractReturn, system contract return object
type ContractReturn struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ParseSysContractResult parsed the result to ContractReturn object
func ParseSysContractResult(result []byte) (*ContractReturn, error) {
	a := ContractReturn{} //删除？
	err := json.Unmarshal(result, &a)
	if err != nil {
		// utils.Fatalf(utl.ErrUnmarshalBytesFormat, "contract return", err.Error())
		errStr := fmt.Sprintf(utl.ErrUnmarshalBytesFormat, "contract return", err.Error())
		return nil, errors.New(errStr)
	}

	return &a, nil
}

//================================CNS=================================

type Cns struct {
	To     string
	Name   string // the cns name of contract
	txType uint64 // the transaction type of the contract execution (EXECUTE_CONTRACT or CNS_TX_TYPE)
}

func NewCns(to, name string, txType uint64) *Cns {
	return &Cns{
		To:     to,
		Name:   name,
		txType: txType,
	}
}

//================================ABI================================
// todo: change FuncDesc to Method in abi/method.go
// FuncDesc, the object of the contract abi files
type FuncDesc struct {
	Name            string                   `json:"name"`
	Inputs          []abi.ArgumentMarshaling `json:"inputs"`
	Outputs         []abi.ArgumentMarshaling `json:"outputs"`
	Constant        interface{}              `json:"constant"` // ???
	Type            string                   `json:"type"`
	StateMutability string                   `json:"stateMutability,omitempty"` // tag for solidity ver > 0.6.0
}

type FuncIO struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Indexed      bool     `json:"indexed,omitempty"`
	InternalType string   `json:"internalType,omitempty"`
	Components   []FuncIO `json:"components,omitempty"`
}

// ParseFuncFromAbi searches the function (or event) names in the []FuncDesc object array
func ParseFuncFromAbi(abiBytes []byte, name string) (*FuncDesc, error) {
	funcs, err := ParseAbiFromJson(abiBytes)
	if err != nil {
		return nil, err
	}

	for _, value := range funcs {
		if value.Name == name {
			return &value, nil
		}
	}

	funcList := ListAbiFuncName(funcs)

	return nil, fmt.Errorf("function/event %s is not found in\n%s", name, funcList)
}

//
func ListAbiFuncName(abiFuncs []FuncDesc) string {
	var result string

	result = fmt.Sprintf("-------------------contract methods list------------------------\n")

	for _, function := range abiFuncs {
		strInput := []string{}
		strOutput := []string{}
		for _, param := range function.Inputs {
			strInput = append(strInput, param.Name+" "+param.Type)
		}
		for _, param := range function.Outputs {
			strOutput = append(strOutput, param.Name+" "+param.Type)
		}
		result += fmt.Sprintf("%s: ", function.Type)
		result += fmt.Sprintf("%s(%s)%s\n", function.Name, strings.Join(strInput, ","), strings.Join(strOutput, ","))
	}

	return result
}

// ParseAbiFromJson parses the application binary interface(abi) files to []FuncDesc object array
func ParseAbiFromJson(abiBytes []byte) ([]FuncDesc, error) {
	var a []FuncDesc

	if abiBytes == nil {
		return nil, errors.New("abiBytes are null")
	}

	if err := json.Unmarshal(abiBytes, &a); err != nil {
		return nil, fmt.Errorf(utl.ErrUnmarshalBytesFormat, "abi", err.Error())
	}

	return a, nil
}

//=========================Transaction=================================

// TxParamsDemo, the object of the eth_call, eth_sendTransaction
type TxParams struct {
	From     common.Address  `json:"from"` // the address used to send the transaction
	To       *common.Address `json:"to"`   // the address receives the transactions
	Gas      string          `json:"gas"`
	GasPrice string          `json:"gasPrice"`
	Value    string          `json:"value"`
	Data     string          `json:"data"`
}

// NewTxParams news a TxParams object
func NewTxParams(from common.Address, to *common.Address, value, gas, gasPrice, data string) *TxParams {

	tx := &TxParams{
		From:     from,
		To:       to,
		GasPrice: gasPrice,
		Gas:      gas,
		Value:    value,
		Data:     data,
	}

	return tx
}

// SendMode selects the rpc calls (eth_call, eth_sendTransaction, and eth_sendRawTransaction)
func (tx *TxParams) SendMode(isWrite bool, keyfileJson []byte) ([]interface{}, string) {
	var action string
	var params = make([]interface{}, 0)

	switch {
	case !isWrite:
		params = append(params, tx)
		params = append(params, "latest")
		action = "eth_call"
	case len(keyfileJson) != 0:
		signedTx := tx.GetSignedTx(keyfileJson)
		params = append(params, signedTx)
		action = "eth_sendRawTransaction"
	default:
		params = append(params, tx)
		action = "eth_sendTransaction"
	}

	return params, action
}

// GetSignedTx gets the signed transaction
func (tx *TxParams) GetSignedTx(keyfileJson []byte) string {

	var txSign *types.Transaction

	// convert the TxParams object to types.Transaction object
	nonce := getNonceRand()
	value, _ := hexutil.DecodeBig(tx.Value)
	gas, _ := hexutil.DecodeUint64(tx.Gas)
	gasPrice, _ := hexutil.DecodeBig(tx.GasPrice)
	data, _ := hexutil.Decode(tx.Data)

	if tx.To == nil {
		txSign = types.NewContractCreation(nonce, value, gas, gasPrice, data)
	} else {
		txSign = types.NewTransaction(nonce, *tx.To, value, gas, gasPrice, data, common.DefaultTxType)
	}

	// extract pk from keystore file and sign the transaction
	priv := utl.GetPrivateKey(keyfileJson)

	// todo: choose the correct signer
	txSign, _ = types.SignTx(txSign, types.HomesteadSigner{}, priv)
	/// txSign, _ = types.SignTx(txSign, types.NewEIP155Signer(big.NewInt(300)), priv)
	/// utl.Logger.Printf("the signed transaction is %v\n", txSign)

	str, err := rlpEncode(txSign)
	if err != nil {
		panic(err)
	}

	return str
}

// getNonceRand generate a random nonce
// Warning: if the design of the nonce mechanism is modified
// this part should be modified as well
func getNonceRand() uint64 {
	rand.Seed(time.Now().Unix())
	return rand.Uint64()
}

/*
func getSigner() {
	readChainConfig()

	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainID
	}
}*/
