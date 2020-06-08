package vm

import "github.com/PlatONEnetwork/PlatONE-Go/common"

// the system contract addr  table
var (
	USER_MANAGEMENT_ADDRESS      = common.HexToAddress("0x1000000000000000000000000000000000000001") // The PlatONE Precompiled contract addr for user management
	NODE_MANAGEMENT_ADDRESS      = common.HexToAddress("0x1000000000000000000000000000000000000002") // The PlatONE Precompiled contract addr for node management
	CNS_MANAGEMENT_ADDRESS       = common.HexToAddress("0x1000000000000000000000000000000000000003") // The PlatONE Precompiled contract addr for CNS
	PARAMETER_MANAGEMENT_ADDRESS = common.HexToAddress("0x1000000000000000000000000000000000000004") // The PlatONE Precompiled contract addr for parameter management
)
