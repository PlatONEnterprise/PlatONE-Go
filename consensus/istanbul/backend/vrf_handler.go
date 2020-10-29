package backend

import (
	"crypto/ecdsa"
	//"encoding/hex"
	"errors"
	"fmt"
	//"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/crypto/vrf"
)

var (
	ErrInvalidVrfProve = errors.New("invalid vrf prove")
	ErrStorageNonce    = errors.New("storage previous nonce failed")
)

func (sb *backend) GenerateNonce(parentNonce []byte) ([]byte, error) {

	if nonce, err := vrf.Prove(sb.privateKey, parentNonce); nil != err {
		return nil, err
	} else {
		if len(nonce) > 0 {
			return nonce, nil
		}
	}
	return nil, fmt.Errorf("generate proof failed, seed:%x", parentNonce)
}

func (sb *backend) VerifyVrf(pk *ecdsa.PublicKey, parentNonce []byte, proof []byte) error {
	// Verify VRF Proof
	if value, err := vrf.Verify(pk, proof, parentNonce); nil != err {
		return err
	} else if !value {
		return ErrInvalidVrfProve
	}
	return nil
}
