package packet

import (
	"encoding/json"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/utils"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
)

func TestEvmStringToEncodeByte(t *testing.T) {
	var args = make([]interface{}, 0)
	var arguments abi.Arguments

	paramsBytes, err := arguments.PackV2(args...)
	if err != nil {
		t.Fatal(err)
	}

	if paramsBytes == nil {
		t.Logf("success, the result is nil")
	}
}

func TestMarshalIndent(t *testing.T) {
	var r = new(ReceiptParsingReturn)
	rBytes, _ := json.MarshalIndent(r, "", "\t")
	t.Log(string(rBytes))
}

// a bug find in arguments.UnpackV2() method when decoding tuple[2][] type
func TestEvmUnpack(t *testing.T) {
	var respStr = "0x0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000004" +
		"21d6a643254dcb9e66c3d658d368a7bc08aeab16bff9e68045680b2ab9945504" +
		"15694c184e89310996ca3f6c7a305200b7ae625686684d52776d8fd334bce3cd" +
		"0339ede104561962311dacf91b9c5faea8f21c3df2b7810743725c943a25f6ff" +
		"0a2bba17994e1543194d6c73d4e5b7ff721e4a6cb47685a9e65e5625686f2215" +
		"2554e91676677bacdf69773da2be001f29f144ad90ae0485195a1d5619deb2bf" +
		"304a74d0dbd690c7f89d922541cb2f3047acf32b845de2d6fa816334b8a08529" +
		"0339ede104561962311dacf91b9c5faea8f21c3df2b7810743725c943a25f6ff" +
		"0a2bba17994e1543194d6c73d4e5b7ff721e4a6cb47685a9e65e5625686f2215" +
		"19fa75093db1467a9663460e5a669caeb1e13ed54249838af112478544b4cb45" +
		"10b7807ea1f11573d1c6c60c25a82d9fd1010bef37e9e6b7ea7cb0da3e8bd6cc" +
		"14bcc435f49d130d189737f9762feb25c44ef5b886bef833e31a702af6be4748" +
		"10cd33954522ad058f00a2553fd4e10d859fe125997e98adba777910dddc5322" +
		"2a4f308883aac5e565d6707d46e567c7e15f7b39a5d9c9c1cb6d455601a2ced8" +
		"078af48992fe65e8c44080311c49887bf7ec3cf4240c48326afdf1de83b86dbf" +
		"0339ede104561962311dacf91b9c5faea8f21c3df2b7810743725c943a25f6ff" +
		"0a2bba17994e1543194d6c73d4e5b7ff721e4a6cb47685a9e65e5625686f2215"

	funcAbi, err := utils.ParseFileToBytes("../test/test_case/sol/privacyToken_sol_PToken.abi")
	if err != nil {
		t.Fatal(err)
	}

	abiFunc, err := ParseFuncFromAbi(funcAbi, "simulateAccounts") //修改
	if err != nil {
		t.Fatal(err)
	}

	outputType := abiFunc.Outputs

	arguments := GenUnpackArgs(outputType)
	result := arguments.ReturnBytesUnpack(respStr)

	t.Log(result)
}
