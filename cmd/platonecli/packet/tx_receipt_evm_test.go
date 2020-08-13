package packet

import (
	"encoding/json"
	"testing"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/precompiled"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
)

const (
	receiptLogJson  = "[{\"address\":\"0x2b61dfc97ae67b485415c46aa8461f64248239e0\",\"topics\":[\"0xb7b07f2333658400bbfd4f0d84c4185b3c70bb6baaf5bda34130fc0ec7eaad3a\",\"0x9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658\"],\"data\":\"0x00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000a6d73672e73656e64657200000000000000000000000000000000000000000000\"}]"
	receiptLogJson2 = "[{\"address\":\"0x0bd5566919ade57ccb3cdceda890947ab40f477f\",\"topics\":[\"0xe9295ed99d8e710c9ebd8d937d49892b715d35c09c29f1ab96f1b5a1b3e17fb5\"],\"data\":\"0xd3925065726d697373696f6e2044656e69656421\"}]"
	receiptLogJson3 = "[{\"address\":\"0x0000000000000000000000000000000000000000\",\"topics\":[\"0x03beaa4c9a962fe1f4b68bed05e2e1a015fd09fbfd4d8e690e35a86eb8e0683e\"],\"data\":\"0xd78095636e73696e766f6b65207375636365737366756c2e\"}]"

	abiFilePath = "../test/test_case/sol/abiPack/___tupleTest_sol_TupleTest.abi"
)

func TestEventParsingV2(t *testing.T) {
	var logs RecptLogs

	err := json.Unmarshal([]byte(receiptLogJson2), &logs)
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

func TestSysEventParsing(t *testing.T) {
	var logs RecptLogs

	err := json.Unmarshal([]byte(receiptLogJson3), &logs)
	if err != nil {
		t.Fatal(err)
	}

	result := SysEventParsing(logs, []string{precompile.CnsInvokeEvent})
	t.Log(result)
}
