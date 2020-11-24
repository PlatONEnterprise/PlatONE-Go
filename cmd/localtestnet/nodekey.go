package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"io/ioutil"
)

var bootnode = ""

func PrkToEnode(prk *ecdsa.PrivateKey) string {
	pubkey := hex.EncodeToString(crypto.FromECDSAPub(&prk.PublicKey)[1:])

	return fmt.Sprintf("enode://%s@127.0.0.1:%d", pubkey, p2pPortBase)
}

func PrkToHex(prk *ecdsa.PrivateKey) string {
	return hex.EncodeToString(crypto.FromECDSA(prk))
}

//return privatekey
func genNodeKey() *ecdsa.PrivateKey {
	// generate random.
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(fmt.Errorf("Failed to generate random private key: %w", err))
	}

	if "" == bootnode {
		bootnode = PrkToEnode(privateKey)
	}
	// Output some information.
	//	Address:    crypto.PubkeyToAddress(privateKey.PublicKey).Hex(),
	//	PublicKey:  hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey)[1:]),
	//	PrivateKey: hex.EncodeToString(crypto.FromECDSA(privateKey)),

	return privateKey
}

//return privatekey
func genNodeKeyFile(filename string) *ecdsa.PrivateKey {
	prk := genNodeKey()
	if err := ioutil.WriteFile(filename, []byte(PrkToHex(prk)), 0666); nil != err {
		panic(err)
	}

	return prk
}
