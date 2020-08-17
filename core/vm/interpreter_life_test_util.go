package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

type ContractRefSelf struct {
}

func (c ContractRefSelf) Address() common.Address {
	return common.BigToAddress(big.NewInt(66666))
}

type ContractRefCaller struct {
}

func (c ContractRefCaller) Address() common.Address {
	return common.BigToAddress(big.NewInt(77777))
}

func genInput() []byte {
	var input [][]byte
	input = make([][]byte, 0)
	input = append(input, utils.Int64ToBytes(1))
	input = append(input, []byte("transfer"))
	input = append(input, []byte("0x0000000000000000000000000000000000000001"))
	input = append(input, []byte("0x0000000000000000000000000000000000000002"))
	input = append(input, utils.Int64ToBytes(100))

	buffer := new(bytes.Buffer)
	err := rlp.Encode(buffer, input)
	if err != nil {
		fmt.Println("geninput fail.", err)
	}
	return buffer.Bytes()
}

type StateDBTest struct {
}

func bytes2int64(byt []byte) int64 {
	bytesBuf := bytes.NewBuffer(byt)
	var tmp int64
	binary.Read(bytesBuf, binary.BigEndian, &tmp)
	return tmp
}

func genGetInput() []byte {
	var input [][]byte
	input = make([][]byte, 0)
	input = append(input, utils.Int64ToBytes(1))
	input = append(input, []byte("Get"))
	input = append(input, []byte(time.Now().Add(time.Duration(rand.Int63())).String()))

	buffer := new(bytes.Buffer)
	err := rlp.Encode(buffer, input)
	if err != nil {
		fmt.Println("genGetInput fail.", err)
	}
	return buffer.Bytes()
}

func genSetInput() []byte {
	var input [][]byte
	input = make([][]byte, 0)
	input = append(input, utils.Int64ToBytes(1))
	input = append(input, []byte("Set"))
	input = append(input, []byte(time.Now().Add(time.Duration(rand.Int63())).String()))
	input = append(input, utils.Int32ToBytes(rand.Int31()))

	buffer := new(bytes.Buffer)
	err := rlp.Encode(buffer, input)
	if err != nil {
		fmt.Println("genSetInput fail.", err)
	}
	return buffer.Bytes()
}

func genSetFixedInput() []byte {
	var input [][]byte
	input = make([][]byte, 0)
	input = append(input, utils.Int64ToBytes(1))
	input = append(input, []byte("Set"))
	input = append(input, []byte("platone"))
	input = append(input, utils.Int32ToBytes(11))

	buffer := new(bytes.Buffer)
	err := rlp.Encode(buffer, input)
	if err != nil {
		fmt.Println("genSetFixedInput fail.", err)
	}
	return buffer.Bytes()
}

func genGetFixedInput() []byte {
	var input [][]byte
	input = make([][]byte, 0)
	input = append(input, utils.Int64ToBytes(1))
	input = append(input, []byte("Get"))
	input = append(input, []byte("platone"))

	buffer := new(bytes.Buffer)
	err := rlp.Encode(buffer, input)
	if err != nil {
		fmt.Println("genGetFixedInput fail.", err)
	}
	return buffer.Bytes()
}
