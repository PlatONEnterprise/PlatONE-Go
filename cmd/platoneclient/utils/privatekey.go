package utils

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/keystore"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/console"
)

type Keyfile struct {
	Address    string `json:"address"`
	Json       []byte
	Passphrase string
}

func NewKeyfile(keyfilePath string) (*Keyfile, error) {
	var keyfile = new(Keyfile)

	if keyfilePath != "" {
		keyJson, err := ParseFileToBytes(keyfilePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(keyJson, keyfile)
		if err != nil {
			return nil, fmt.Errorf(ErrUnmarshalBytesFormat, keyJson, err.Error())
		}

		keyfile.Json = keyJson
	}

	return keyfile, nil
}

// GetPrivateKey gets the private key by decrypting the keystore file
func (k *Keyfile) GetPrivateKey() (*ecdsa.PrivateKey, error) {

	// Decrypt key with passphrase.
	if k.Passphrase == "" {
		k.Passphrase = promptPassphrase(true)
	}

	key, err := keystore.DecryptKey(k.Json, k.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("Error decrypting key: %v", err)
	}

	return key.PrivateKey, nil
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
