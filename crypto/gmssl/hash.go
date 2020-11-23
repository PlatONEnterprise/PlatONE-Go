package gmssl

import (
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl/gmssl"
	"hash"
)

func HashWithSM3(msg []byte) []byte {
	hasher := New("SM3")
	hasher.Write(msg)
	return hasher.Sum(nil)
}

func HashWithSha256(msg []byte) []byte {
	hasher := New("SHA256")
	hasher.Write(msg)
	return hasher.Sum(nil)
}

type digest struct {
	ctx *gmssl.DigestContext
}

func New(algo string) hash.Hash {
	d := new(digest)
	ctx, err := gmssl.NewDigestContext(algo)
	if err != nil {
		return nil
	}
	d.ctx = ctx
	return d
}

func (d *digest) BlockSize() int {
	ret, err := gmssl.GetDigestBlockSize("SM3")
	if err != nil {
		return 0
	}
	return ret
}

func (d *digest) Size() int {
	ret, err := gmssl.GetDigestLength("SM3")
	if err != nil {
		return 0
	}
	return ret
}

func (d *digest) Reset() {
	_ = d.ctx.Reset()
}

func (d *digest) Write(p []byte) (int, error) {
	err := d.ctx.Update(p)
	return len(p), err
}

func (d *digest) Sum(in []byte) []byte {
	d.ctx.Update(in)
	ret, err := d.ctx.Final()
	if err != nil {
		return nil
	}
	return ret
}



