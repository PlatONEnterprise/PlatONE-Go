package utils

import (
	"crypto/ecdsa"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/keystore"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/console"
)

const defaultAddress = "0x0000000000000000000000000000000000000000"

type KeystoreJson struct {
	Address string `json:"address"`
	Json    []byte
}

// GetPrivateKey gets the private key by decrypting the keystore file
func GetPrivateKey(keyJson []byte) *ecdsa.PrivateKey {

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
