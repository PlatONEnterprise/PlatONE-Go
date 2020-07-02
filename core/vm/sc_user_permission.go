package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
)


func checkPermission(state StateDB, user common.Address, role int32) bool{
	um := &UserManagement{
		state:state,
		address: syscontracts.USER_MANAGEMENT_ADDRESS,
	}
	roleName,ok := rolesName[role]
	if !ok{
		return false
	}

	b, e := um.hasRole(user, roleName)
	if e != nil{
		return false
	}

	return b == 1
}

func hasNodeOpPermmision(state  StateDB, addr common.Address) bool{
	return checkPermission(state,addr, CHAIN_ADMIN) ||
		   checkPermission(state,addr, NODE_ADMIN)
}

func hasContractDeployPermmision(state  StateDB, addr common.Address) bool{
	return checkPermission(state,addr, CHAIN_ADMIN) ||
		   checkPermission(state,addr, CONTRACT_DEPLOYER) ||
		   checkPermission(state,addr, CONTRACT_ADMIN)
}

func hasParamOpPermmision(state  StateDB, addr common.Address) bool{
	return checkPermission(state,addr, CHAIN_ADMIN)
}

func hasGroupCreatePermmision(state  StateDB, addr common.Address) bool{
	return checkPermission(state,addr, CHAIN_ADMIN)
}
