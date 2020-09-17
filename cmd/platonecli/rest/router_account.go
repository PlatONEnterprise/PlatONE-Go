package rest

import (
	"crypto/ecdsa"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/keystore"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
)

func registerAccountRouters(r *gin.Engine) {
	cns := r.Group("/accounts")
	{
		cns.POST("", newAccountHandler)
	}
}

// ====================== ACCOUNT ======================
const defaultKeyfile = "./keystore"

func newAccountHandler(ctx *gin.Context) {
	// password
	passphrase := ctx.PostForm("passphrase")
	if passphrase == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "the passphrase can not be empty"})
		return
	}

	var privateKey *ecdsa.PrivateKey
	var err error
	if file := ctx.PostForm("privatekey"); file != "" {
		// Load private key from file.
		privateKey, err = crypto.LoadECDSA(file)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"Can't load private key: ": err.Error()})
			return
		}
	} else {
		// If not loaded, generate random.
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"Failed to generate random private key: ": err.Error()})
			return
		}
	}

	// Create the keyfile object with a random UUID.
	id := uuid.NewRandom()
	key := &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}

	// Encrypt key with passphrase.
	keyjson, err := keystore.EncryptKey(key, passphrase, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error encrypting key: ": err.Error()})
		return
	}

	runPath := utils.GetRunningTimePath()
	keyfileDirt := runPath + defaultKeyfile
	pathSep := string(os.PathSeparator)
	keyfilepath := keyfileDirt + pathSep + "UTC--" + time.Now().Format("2006-01-02") + "--" + key.Address.Hex()

	// Store the file to disk.
	if err := os.MkdirAll(filepath.Dir(keyfilepath), 0700); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Failed to generate random private key: ": err.Error()})
		return
	}
	if err := ioutil.WriteFile(keyfilepath, keyjson, 0600); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Failed to write keyfile: ": err.Error()})
		return
	}

	// Output some information.
	ctx.JSON(200, gin.H{
		"Address": key.Address.Hex(),
	})
}
