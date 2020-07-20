package main

import (
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
)

type contract struct {
	address string
	funcAbi []byte
	vm      string
}

func newContract() *contract {
	return &contract{}
}

type TxAccount struct {
	address common.Address
	keyfile string
}

func NewTxAccount(address, keyfile string) *TxAccount {
	return &TxAccount{
		address: common.HexToAddress(address),
		keyfile: keyfile,
	}
}

func getClientConfig(c *cli.Context) (*TxAccount, bool, bool, string) {
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

	account := NewTxAccount(address, keyfile)

	if isDefault {
		config.Account = address
		config.Keystore = keyfile
		config.Url = url
	}

	return account, isSync, isDefault, url
}

func getUrl(c *cli.Context) string {
	url := c.String(UrlFlags.Name)

	if url != "" {
		url = reformatUrl(url)
		paramValid(url, "url")
		return url
	}

	if config.Url == "" {
		utils.Fatalf("Please set url first.\n")
	}

	return config.Url
}

func reformatUrl(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[len("http://"):]
	}

	return url
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
