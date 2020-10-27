package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"io/ioutil"
)

var privateKey *ecdsa.PrivateKey

func getEnode() string {
	if nil != privateKey {
		genNodeKey()
	}

	pubkey := hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey)[1:])

	return fmt.Sprintf("enode://%s:127.0.0.1:1680", pubkey)
}

//return privatekey
func genNodeKey() string {
	if nil != privateKey {
		return hex.EncodeToString(crypto.FromECDSA(privateKey))
	}

	var err error
	// generate random.
	privateKey, err = crypto.GenerateKey()
	if err != nil {
		panic(fmt.Errorf("Failed to generate random private key: %w", err))
	}

	// Output some information.
	//	Address:    crypto.PubkeyToAddress(privateKey.PublicKey).Hex(),
	//	PublicKey:  hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey)[1:]),
	//	PrivateKey: hex.EncodeToString(crypto.FromECDSA(privateKey)),
	return hex.EncodeToString(crypto.FromECDSA(privateKey))
}

func genNodeKeyFile(filename string) {
	if err := ioutil.WriteFile(filename, []byte(genNodeKey()), 0666); nil != err {
		panic(err)
	}
}
