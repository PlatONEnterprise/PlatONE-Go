package utils

import (
	"crypto/ecdsa"
	"encoding/json"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/keystore"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/console"
)

const defaultAddress = "0x0000000000000000000000000000000000000000"

type KeystoreJson struct {
	Address string `json:"address"`
	Crypto  string `json:"crypto"`
}

// GetPrivateKey gets the private key by decrypting the keystore file
func GetPrivateKey(account common.Address, keyfilepath string) *ecdsa.PrivateKey {

	// Load the keyfile.
	keyJson, err := ParseFileToBytes(keyfilepath)
	if err != nil {
		utils.Fatalf("Failed to read the keyfile at '%s': %v", keyfilepath, err)
	}

	keyfile := KeystoreJson{}
	_ = json.Unmarshal(keyJson, &keyfile)

	// check if the account address is matched
	addr := account.String()
	if addr != defaultAddress && !strings.EqualFold(keyfile.Address, addr[2:]) {
		utils.Fatalf("the keystore file mismatches the account address")
	}

	// Decrypt key with passphrase.
	passphrase := promptPassphrase(true)
	key, err := keystore.DecryptKey(keyJson, passphrase)
	if err != nil {
		utils.Fatalf("Error decrypting key: %v", err)
	}

	return key.PrivateKey
}

// todo: change prompt to --password flag?
// promptPassphrase prompt the hint in the terminal to let user to input the password
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
