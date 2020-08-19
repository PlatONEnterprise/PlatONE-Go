package packet

import (
	"encoding/json"
	"testing"

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

	t.Logf("the nil string is null: %v\n", string(nil) == "")
}

func TestMarshalIndent(t *testing.T) {
	var r = new(ReceiptParsingReturn)
	rBytes, _ := json.MarshalIndent(r, "", "\t")
	t.Log(string(rBytes))
}
