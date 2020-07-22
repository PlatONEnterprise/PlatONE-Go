package main

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/platoneclient"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
)

// messageCall extract the common parts of the transaction based calls
// including eth_call, eth_sendTransaction, and eth_sendRawTransaction
func messageCall(c *cli.Context, call packet.MessageCall, to *common.Address, value string) interface{} {

	// get the global parameters
	address, keystore, gas, gasPrice, isSync, isDefault := getGlobalParam(c)
	from := common.HexToAddress(address)

	// todo: remove this statement
	if call == nil {
		utils.Fatalf("")
	}

	// combine the data based on the types of the calls (contract call, inner call or deploy call)
	data, outputType, isWrite, err := call.CombineData()
	if err != nil {
		utils.Fatalf(utl.ErrPackDataFormat, err.Error())
	}

	// packet the transaction and select the transaction based calls
	tx := packet.NewTxParams(from, to, value, gas, gasPrice, data)
	params, action := tx.SendMode(isWrite, keystore)

	// print the RPC JSON param to the terminal
	/// utl.PrintRequest(params)

	// send the RPC calls
	resp, err := platoneclient.RpcCalls(action, params)
	if err != nil {
		utils.Fatalf(utl.ErrSendTransacionFormat, err.Error())
	}

	// todo: move to another place?
	setDefault(address, keystore, isDefault)

	return platoneclient.ParseTxResponse(resp, outputType, isWrite, isSync)
}

// setDefault write values of account and keystore to config file if default flag provided
func setDefault(account, keystore string, isDefault bool) {
	if isDefault {

		runPath := utl.GetRunningTimePath()

		WriteConfigFile(runPath+defaultConfigFilePath, "account", account)
		WriteConfigFile(runPath+defaultConfigFilePath, "keystore", keystore)
	}
}

// getGlobalParam gets and converts the global parameters
func getGlobalParam(c *cli.Context) (string, string, string, string, bool, bool) {

	if c == nil {
		panic("the cli.context pointer is nill")
	}

	// set the url for http request
	setUrl(c)

	// get the global parameters from cli.context
	//TODO 分类规划???
	gas := c.String(GasFlags.Name)
	gasPrice := c.String(GasPriceFlags.Name)

	address := c.String(AccountFlags.Name)
	keystore := c.String(KeyfileFlags.Name)
	isLocal := c.Bool(LocalFlags.Name)
	isDefault := c.Bool(DefaultFlags.Name)

	isSync := !c.Bool(SyncFlags.Name)

	// check and covert the global parameters
	optionParamValid(address, "address")
	keystore = getKeystore(keystore, isLocal)
	if address == "" && keystore == "" {
		address = config.Account
		keystore = config.Keystore
	}

	gas = chainParamConvert(gas, "gas").(string)
	gasPrice = chainParamConvert(gasPrice, "gasPrice").(string)

	return address, keystore, gas, gasPrice, isSync, isDefault
}

// setUrl sets the url for http request
func setUrl(c *cli.Context) {
	if c == nil {
		panic("the cli.context pointer is nill")
	}

	url := c.String(UrlFlags.Name)

	runPath := utl.GetRunningTimePath()

	switch {
	case url != "":
		paramValid(url, "url")
		config.Url = url
		WriteConfigFile(runPath+defaultConfigFilePath, "url", config.Url)
	case config.Url == "":
		utils.Fatalf("Please set url first.\n")
	default:
		//utils.Fatalf(utl.PanicUnexpSituation, "SetUrl")
	}

	platoneclient.SetHttpUrl(config.Url)
}

// getKeystore gets the path of keystore file based on the keystore and isLocal flags
func getKeystore(keystore string, isLocal bool) string {
	if isLocal && keystore == "" {
		keystore, _ = utl.GetFileInDirt(DEFAULT_KEYSTORE_DIRT)
	}
	return keystore
}
