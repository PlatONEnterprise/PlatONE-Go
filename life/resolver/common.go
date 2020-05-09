package resolver

import (
	"encoding/hex"
	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/life/exec"
	"strings"
)

func parseWasmCallSolInput(vm *exec.VirtualMachine, address, input []byte) ([]byte, error) {
	// Only used in compatibility mode
	if !strings.EqualFold(common.GetCurrentInterpreterType(), "all") {
		return input, nil
	}

	code := vm.Context.StateDB.GetCode(common.HexToAddress(hex.EncodeToString(address)))
	if ok, _, _, _ := common.IsWasmContractCode(code); ok {
		return input, nil
	}
	return abi.ParseWasmCallSolInput(input)
}
