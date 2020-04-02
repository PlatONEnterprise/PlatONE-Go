package core

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/PlatONEnetwork/PlatONE-Go/accounts/keystore"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/console"
)

func getPrivateKey(account, keyfilepath string) *ecdsa.PrivateKey {

	// Load the keyfile.
	keyJson, err := parseFileToBytesDemo(keyfilepath)
	//keyJson, err := ioutil.ReadFile(keyfilepath)
	if err != nil {
		utils.Fatalf("Failed to read the keyfile at '%s': %v", keyfilepath, err)
	}

	keyInfo, _ := parseKeystoreFromJsonDemo(keyJson)
	if account[2:] != keyInfo.Address {
		utils.Fatalf("the keystore file mismatch the account address")
	}

	// Decrypt key with passphrase.
	passphrase := promptPassphrase(true)
	key, err := keystore.DecryptKey(keyJson, passphrase)
	if err != nil {
		utils.Fatalf("Error decrypting key: %v", err)
	}

	return key.PrivateKey
}

func parseKeystoreFromJsonDemo(keystore []byte) (keystoreJson, error) {

	var a keystoreJson
	if err := json.Unmarshal(keystore, &a); err != nil {
		//return  ,fmt.Errorf("parse keystore to json error: %s", err.Error())
	}
	return a, nil
}

func promptPassphrase(confirmation bool) string {
	passphrase, err := console.Stdin.PromptPassword("Passphrase: ")
	if err != nil {
		utils.Fatalf("Failed to read passphrase: %v", err)
	}

	if confirmation {
		confirm, err := console.Stdin.PromptPassword("Repeat passphrase: ")
		if err != nil {
			utils.Fatalf("Failed to read passphrase confirmation: %v", err)
		}
		if passphrase != confirm {
			utils.Fatalf("Passphrases do not match")
		}
	}

	return passphrase
}