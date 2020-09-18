package utils

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/keystore"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/console"
)

var ErrNullPassphrase = errors.New("passphrase is null")

type Keyfile struct {
	Address    string `json:"address"`
	Json       []byte
	Passphrase string
	privateKey *ecdsa.PrivateKey
}

func NewKeyfile(keyfilePath string) (*Keyfile, error) {
	var keyfile = new(Keyfile)

	keyJson, err := ParseFileToBytes(keyfilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(keyJson, keyfile)
	if err != nil {
		return nil, fmt.Errorf(ErrUnmarshalBytesFormat, keyJson, err.Error())
	}

	keyfile.Json = keyJson

	return keyfile, nil
}

func (k *Keyfile) GetPrivateKey() *ecdsa.PrivateKey {
	return k.privateKey
}

// GetPrivateKey gets the private key by decrypting the keystore file
func (k *Keyfile) ParsePrivateKey() error {

	// Decrypt key with passphrase.
	/*
		if k.Passphrase == "" {
			return nil, ErrNullPassphrase
		}*/

	key, err := keystore.DecryptKey(k.Json, k.Passphrase)
	if err != nil {
		return fmt.Errorf("Error decrypting key: %v", err)
	}

	k.privateKey = key.PrivateKey
	return nil
}

// todo: change prompt to --password flag?
// promptPassphrase prompt the hint in the terminal to let user to input the password
func PromptPassphrase(confirmation bool) string {
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
