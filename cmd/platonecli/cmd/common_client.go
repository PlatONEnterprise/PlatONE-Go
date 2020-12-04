package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/utils"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/packet"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
)

func clientCommonV2(c *cli.Context, dataGen packet.MsgDataGen, to *common.Address) []interface{} {
	// get the client global parameters
	keyfile, isSync, isDefault, url := getClientConfig(c)

	pc, err := platoneclient.SetupClient(url)
	if err != nil {
		utl.Fatalf("set up client failed: %s\n", err.Error())
	}
	// form transaction
	tx := getTxParams(c)
	tx.From = common.HexToAddress(keyfile.Address)
	tx.To = to
	result, err := pc.MessageCallV2(dataGen, tx, keyfile, isSync)
	if err != nil {
		utl.Fatalf("to do: %s\n", err.Error())
	}

	// store default values to config file
	if isDefault && !reflect.ValueOf(result).IsZero() {
		runPath := utils.GetRunningTimePath()
		WriteConfig(runPath+DefaultConfigFilePath, config)
	}

	return result
}

// ===========================================================

func getClientConfig(c *cli.Context) (*utils.Keyfile, bool, bool, string) {
	var account *utils.Keyfile
	var err error

	address := c.String(AccountFlags.Name)
	keyfile := c.String(KeyfileFlags.Name)
	isDefault := c.Bool(DefaultFlags.Name)
	isSync := !c.Bool(SyncFlags.Name)
	url := getUrl(c)

	optionParamValid(address, "address")
	if address == "" && keyfile == "" {
		address = config.Account
		keyfile = config.Keystore
	}

	if keyfile != "" {
		account, err = utils.NewKeyfile(keyfile)
		if err != nil {
			utl.Fatalf(err.Error())
		}

		account.Passphrase = utils.PromptPassphrase(false)

		err = account.ParsePrivateKey()
		if err != nil {
			utl.Fatalf(err.Error())
		}
	}

	if !isTxAccountMatch(address, account) {
		fmt.Printf("there is conflict in --account and --keyfile, " +
			"the result is subject to --keyfile")
	}

	if isDefault {
		config.Account = address
		config.Keystore = keyfile
		config.Url = url
	}

	return account, isSync, isDefault, url
}

func isTxAccountMatch(address string, keyfile *utils.Keyfile) bool {

	// check if the account address is matched
	if keyfile.Address != "" && address != "" &&
		!strings.EqualFold(keyfile.Address, address[2:]) {
		return false
	}

	if keyfile.Address == "" {
		keyfile.Address = address
	}

	return true
}

// <URL>: scheme://host:port/path?query#fragment
func getUrl(c *cli.Context) string {
	url := c.String(UrlFlags.Name)
	if url != "" {
		index := strings.Index(url, "://")

		if index != -1 {
			// url = scheme://host:port
			urlNotHttpScheme := !strings.EqualFold(url[:index], "http")
			if urlNotHttpScheme {
				utl.Fatalf("invalid url scheme %s, currently only support http", url[:index])
			}

			paramValid(url[index+3:], "ipAddress")
		} else {
			// url = host:port (default scheme: http://)
			paramValid(url, "ipAddress")
			url = "http://" + url
		}

		return url
	}

	if config.Url == "" {
		config.Url = "http://127.0.0.1:6791"
	}

	return config.Url
}

func getTxParams(c *cli.Context) *packet.TxParams {

	//
	value := c.String(TransferValueFlag.Name)
	gas := c.String(GasFlags.Name)
	gasPrice := c.String(GasPriceFlags.Name)

	value = cmd_common.ChainParamConvert(value, "value").(string)
	gas = cmd_common.ChainParamConvert(gas, "gas").(string)
	gasPrice = cmd_common.ChainParamConvert(gasPrice, "gasPrice").(string)

	return packet.NewTxParams(common.HexToAddress(""), nil, value, gas, gasPrice, "")
}

// ======================== Deprecated ======================================
/*
// todo: rename genTxAndCall?
func clientCommon(c *cli.Context, dataGen packet.MsgDataGen, to *common.Address) []interface{} {

	// get the client global parameters
	keyfile, isSync, isDefault, url := getClientConfig(c)

	pc, err := platoneclient.SetupClient(url)
	if err != nil {
		utl.Fatalf("set up client failed: %s\n", err.Error())
	}

	// form transaction
	tx := getTxParams(c)
	tx.From = common.HexToAddress(keyfile.Address)
	tx.To = to

	// do message call
	result, isTxHash, err := pc.MessageCall(dataGen, keyfile, tx)
	if err != nil {
		utl.Fatalf(err.Error())
	}

	// store default values to config file
	if isDefault && !reflect.ValueOf(result).IsZero() {
		runPath := utils.GetRunningTimePath()
		WriteConfig(runPath+defaultConfigFilePath, config)
	}

	// todo: move isSync from [pc.MessageCall] to here???
	if isSync && isTxHash {
		res, err := pc.GetReceiptByPolling(result[0].(string))
		if err != nil {
			return result
		}

		receiptBytes, _ := json.MarshalIndent(res, "", "\t")
		fmt.Println(string(receiptBytes))

		recpt := dataGen.ReceiptParsing(res)
		if recpt.Status != packet.TxReceiptSuccessMsg {
			result, _ := pc.GetRevertMsg(tx, recpt.BlockNumber)
			if len(result) >= 4 {
				recpt.Err, _ = packet.UnpackError(result)
			}
		}

		result[0] = recpt.String()
	}

	return result
}*/
