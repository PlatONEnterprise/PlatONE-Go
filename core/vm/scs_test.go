package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math/big"
	"testing"
)

func TestA(t *testing.T) {
	a := big.NewInt(-1)
	bin,err := rlp.EncodeToBytes(a)
	if nil != err{
		t.Error(err)
		return
	}

	t.Logf("%b",bin)
	t.Logf("%v",string(bin))
}

