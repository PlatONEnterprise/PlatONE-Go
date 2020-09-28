package keystore

import "github.com/PlatONEnetwork/PlatONE-Go/common"

func KeyFileName(keyAddr common.Address) string {
	return keyFileName(keyAddr)
}
