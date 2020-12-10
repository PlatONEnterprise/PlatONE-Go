// Copyright 2018-2019 The PlatON Network Authors
// This file is part of the PlatON-Go library.
//
// The PlatON-Go library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The PlatON-Go library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the PlatON-Go library. If not, see <http://www.gnu.org/licenses/>.

package backend

import (
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
)

func TestVrfProof_gen(t *testing.T) {
	pri, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	backend := &backend{
		privateKey: pri,
	}
	genesisNonce := types.EncodeByteNonce(hexutil.MustDecode("0x0376e56dffd12ab53bb149bda4e0cbce2b6aabe4cccc0df0b5a39e12977a2fcd23"))
	if err != nil {
		panic(err)
	}
	nonce, err := backend.GenerateNonce(genesisNonce[:])
	if err != nil {
		panic(err)
	}

	if len(nonce) != 81 {
		panic("nonce length not equal to 81")
	}

	if err := backend.VerifyVrf(&(pri.PublicKey), genesisNonce[:], nonce); err != nil {
		panic("nonce verify filed")
	}
}
