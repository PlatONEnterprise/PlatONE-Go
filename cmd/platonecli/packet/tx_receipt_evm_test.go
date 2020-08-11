package packet

import (
	"encoding/json"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
)

const (
	receiptLogJson = "[{\"address\":\"0x2b61dfc97ae67b485415c46aa8461f64248239e0\",\"topics\":[\"0xb7b07f2333658400bbfd4f0d84c4185b3c70bb6baaf5bda34130fc0ec7eaad3a\",\"0x9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658\"],\"data\":\"0x00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000a6d73672e73656e64657200000000000000000000000000000000000000000000\"}]"
	abiFilePath    = "../test/test_case/sol/abiPack/___tupleTest_sol_TupleTest.abi"
)

func TestEventParsingV2(t *testing.T) {
	var logs RecptLogs

	err := json.Unmarshal([]byte(receiptLogJson), &logs)
	if err != nil {
		t.Fatal(err)
	}

	abiBytes, err := utils.ParseFileToBytes(abiFilePath)
	if err != nil {
		t.Fatal(err)
	}

	result := EventParsingV2(logs, abiBytes)
	t.Log(result)
}
