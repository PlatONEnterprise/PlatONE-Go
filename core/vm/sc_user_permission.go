package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
)

const (
	userOpPermission int32 = iota
	groupCreatePermission
	nodeOpPermission
	contractDeployPermission
	paramOpPermission
)

var PermissionMap = map[int32]UserRoles{
	userOpPermission:         1 << chainAdmin,
	groupCreatePermission:    1<<chainAdmin | 1<<groupAdmin,
	nodeOpPermission:         1<<chainAdmin | 1<<nodeAdmin,
	contractDeployPermission: 1<<chainAdmin | 1<<contractAdmin | 1<<contractDeployer,
	paramOpPermission:        1 << chainAdmin,
}

func checkPermission(state StateDB, user common.Address, permission int32) bool {
	um := &UserManagement{
		stateDB:      state,
		contractAddr: syscontracts.UserManagementAddress,
	}

	userRole, err := um.getRole(user)
	if err != nil {
		return false
	}

	if (userRole & PermissionMap[permission]) == 0 {
		return false
	}
	return true
}

func hasUserOpPermission(state StateDB, addr common.Address) bool {
	return checkPermission(state, addr, userOpPermission)
}

func hasNodeOpPermission(state StateDB, addr common.Address) bool {
	return checkPermission(state, addr, nodeOpPermission)
}

func HasContractDeployPermission(state StateDB, addr common.Address) bool {
	return checkPermission(state, addr, contractDeployPermission)
}

func hasParamOpPermission(state StateDB, addr common.Address) bool {
	return checkPermission(state, addr, paramOpPermission)
}

func hasGroupCreatePermission(state StateDB, addr common.Address) bool {
	return checkPermission(state, addr, groupCreatePermission)
}
