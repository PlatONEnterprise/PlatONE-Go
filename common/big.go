// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package common

<<<<<<< HEAD
import "math/big"
=======
import (
	"math/big"
)
>>>>>>> develop

// Common big integers often used
var (
	Big1   = big.NewInt(1)
	Big2   = big.NewInt(2)
	Big3   = big.NewInt(3)
	Big0   = big.NewInt(0)
	Big32  = big.NewInt(32)
	Big256 = big.NewInt(256)
	Big257 = big.NewInt(257)
)
<<<<<<< HEAD
=======

// BigToByte128 convert big.Int to 128-bit big endian,assuming it will not overflow
func BigToByte128(I *big.Int) ([]byte, bool) {
	if len(I.Bytes()) > 16 {
		return []byte{}, false
	}
	res := make([]byte, 16)

	if I.Sign() == -1 {
		// invert then add 1 equals to sub 1 then invert
		Iminus := new(big.Int).Neg(I)
		Iminus.Sub(Iminus, Big1)
		copy(res[16-len(Iminus.Bytes()):], Iminus.Bytes())
		for i := range res {
			res[i] ^= 0xff
		}
	} else {
		copy(res[16-len(I.Bytes()):], I.Bytes())
	}
	return res, true
}

// Byte128ToBig convert byte[] big endian to big.Int,
// s indecates whether b is signed
func Byte128ToBig(b []byte, s bool) *big.Int {
	r := new(big.Int)
	if s && b[0]&0x80 != 0 {
		//invert b
		for i := range b {
			b[i] ^= 0xff
		}
		r.SetBytes(b)
		r.Add(r, Big1)
		r.Neg(r)
		return r
	}

	r.SetBytes(b)
	return r
}
>>>>>>> develop
