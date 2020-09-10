package main

import (
	"errors"
	"fmt"
	"strings"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
)

/*
type contract struct {
	address string
	funcAbi []byte
	vm      string
}

func newContract() *contract {
	return &contract{}
}*/

var ErrAccountNotMatch = errors.New("the keystore file mismatches the account address")

type TxAccount struct {
	address common.Address
	keyfile []byte
}

/*
func keyfileParsing(keyfilePath string) (utl.KeystoreJson, error) {

	var keyfile utl.KeystoreJson
	// Load the keyfile.
	if keyfilePath != "" {
		keyJson, err := utl.ParseFileToBytes(keyfilePath)
		if err != nil {
			return utl.KeystoreJson{}, err
		}

		err = json.Unmarshal(keyJson, &keyfile)
		if err != nil {
			return utl.KeystoreJson{}, fmt.Errorf(utl.ErrUnmarshalBytesFormat, keyJson, err.Error())
		}

		keyfile.Json = keyJson
	}

	return keyfile, nil
}*/

func isTxAccountMatch(address string, keyfile *utl.Keyfile) bool {

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

func getClientConfig(c *cli.Context) (*utl.Keyfile, bool, bool, string) {
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

	account, err := cmd_common.KeyfileParsing(keyfile)
	if err != nil {
		utils.Fatalf(err.Error())
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

// <URL>: scheme://host:port/path?query#fragment
func getUrl(c *cli.Context) string {
	url := c.String(UrlFlags.Name)
	if url != "" {
		index := strings.Index(url, "://")

		if index != -1 {
			// url = scheme://host:port
			urlNotHttpScheme := !strings.EqualFold(url[:index], "http")
			if urlNotHttpScheme {
				utils.Fatalf("invalid url scheme %s, currently only support http", url[:index])
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

	value = chainParamConvert(value, "value").(string)
	gas = chainParamConvert(gas, "gas").(string)
	gasPrice = chainParamConvert(gasPrice, "gasPrice").(string)

	return packet.NewTxParams(common.HexToAddress(""), nil, value, gas, gasPrice, "")
}
