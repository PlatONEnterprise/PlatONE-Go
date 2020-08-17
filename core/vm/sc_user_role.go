package vm

import (
	"encoding/json"
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	superAdmin int32 = iota
	chainAdmin
	groupAdmin
	nodeAdmin
	contractAdmin
	contractDeployer
	rolesCnt
)

var (
	rolesName = map[int32]string{
		superAdmin:       "SUPER_ADMIN",
		chainAdmin:       "CHAIN_ADMIN",
		groupAdmin:       "GROUP_ADMIN",
		nodeAdmin:        "NODE_ADMIN",
		contractAdmin:    "CONTRACT_ADMIN",
		contractDeployer: "CONTRACT_DEPLOYER",
	}
	rolesMap = map[string]int32{
		"SUPER_ADMIN":       superAdmin,
		"CHAIN_ADMIN":       chainAdmin,
		"GROUP_ADMIN":       groupAdmin,
		"NODE_ADMIN":        nodeAdmin,
		"CONTRACT_ADMIN":    contractAdmin,
		"CONTRACT_DEPLOYER": contractDeployer,
	}
)

const (
	roleDeactive = 0
	roleActive   = 1
)

const (
	// userRolesKey = sha3("userRoles")
	userRolesKey = "45332c40978a610bce7447f52a7278f5"
	// addressListKey = sha3("addressList")
	addressListKey = "5fec39173ed9a584db9d70c832080d80"
)

var (
	superAdminAddrListKey       = generateStateKey(addressListKey + "superAdminAddrListKey")
	chainAdminAddrListKey       = generateStateKey(addressListKey + "chainAdminAddrListKey")
	groupAdminAddrListKey       = generateStateKey(addressListKey + "groupAdminAddrListKey")
	nodeAdminAddrListKey        = generateStateKey(addressListKey + "nodeAdminAddrListKey")
	contractAdminAddrListKey    = generateStateKey(addressListKey + "contractAdminAddrListKey")
	contractDeployerAddrListKey = generateStateKey(addressListKey + "contractDeployerAddrListKey")
)

var (
	// 操作某个权限角色所需要的权限
	roleOpPermission map[int32]UserRoles
)

func init() {
	roleOpPermission = make(map[int32]UserRoles, rolesCnt)
	roleOpPermission[superAdmin] = 1 << superAdmin
	roleOpPermission[chainAdmin] = 1 << superAdmin
	roleOpPermission[groupAdmin] = 1 << chainAdmin
	roleOpPermission[nodeAdmin] = 1 << chainAdmin
	roleOpPermission[contractAdmin] = 1 << chainAdmin
	roleOpPermission[contractDeployer] = 1<<chainAdmin | 1<<contractAdmin
}

type UserRoles uint32

func (ur UserRoles) Strings() []string {
	roles := make([]string, 0)

	for i := int32(0); i < rolesCnt; i++ {
		if ur.hasRole(i) {
			roles = append(roles, rolesName[i])
		}
	}
	return roles
}

func (ur *UserRoles) setRole(role int32) error {
	if role >= rolesCnt || role < 0 {
		return errUnsupportedRole
	}
	*ur |= 1 << role
	return nil
}

func (ur *UserRoles) unsetRole(role int32) error {
	if role >= rolesCnt || role < 0 {
		return errUnsupportedRole
	}
	*ur &= ^(1 << role)

	return nil
}

func (ur UserRoles) hasRole(role int32) bool {
	if role >= rolesCnt || role < 0 {
		return false
	}
	return ur&(1<<role) != 0
}

// set superAdmin for a chain
// only one superAdmin can be set for a chain
// if there is already a superAdmin, this method will return error(errAlreadySetSuperAdmin)
func (u *UserManagement) setSuperAdmin() (int32, error) {
	var key []byte
	var err error
	var addrs []common.Address
	var topic = "setSuperAdmin"

	if key, err = generateAddressListKey(superAdmin); err != nil {
		return u.returnFail(topic, err)
	}
	if addrs, err = u.getAddrList(key); err != nil {
		return u.returnFail(topic, err)
	}

	if len(addrs) >= 1 {
		return u.returnFail(topic, errAlreadySetSuperAdmin)
	}

	ur := UserRoles(0)
	if err := ur.setRole(superAdmin); err != nil {
		return u.returnFail(topic, err)
	}

	if err := u.addAddrListOfRole(u.Caller(), superAdmin); err != nil {
		return u.returnFail(topic, err)
	}
	if err := u.setRole(u.Caller(), ur); err != nil {
		return u.returnFail(topic, err)
	}

	return u.returnSuccess(topic)
}

func (u *UserManagement) transferSuperAdminByAddress(addr common.Address) (int32, error) {
	var topic = "transferSuperAdminByAddress"

	if err := u.setRoleWithPermissionCheckByAddress(addr, superAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	if addr == u.Caller() {
		return u.returnSuccess(topic)
	}

	if err := u.setRoleWithPermissionCheckByAddress(u.Caller(), superAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}
func (u *UserManagement) transferSuperAdminByName(name string) (int32, error) {
	topic := "transferSuperAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, superAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	if err := u.setRoleWithPermissionCheckByAddress(u.Caller(), superAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) addChainAdminByAddress(addr common.Address) (int32, error) {
	topic := "addChainAdminByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, chainAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) addChainAdminByName(name string) (int32, error) {
	topic := "addChainAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, chainAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) delChainAdminByAddress(addr common.Address) (int32, error) {
	topic := "delChainAdminByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, chainAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) delChainAdminByName(name string) (int32, error) {
	topic := "delChainAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, chainAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) addGroupAdminByAddress(addr common.Address) (int32, error) {
	topic := "addGroupAdminByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, groupAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) addGroupAdminByName(name string) (int32, error) {
	topic := "addGroupAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, groupAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) delGroupAdminByAddress(addr common.Address) (int32, error) {
	topic := "delGroupAdminByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, groupAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) delGroupAdminByName(name string) (int32, error) {
	topic := "delGroupAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, groupAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) addNodeAdminByAddress(addr common.Address) (int32, error) {
	topic := "addNodeAdminByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, nodeAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) addNodeAdminByName(name string) (int32, error) {
	topic := "addNodeAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, nodeAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) delNodeAdminByAddress(addr common.Address) (int32, error) {
	topic := "delNodeAdminByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, nodeAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) delNodeAdminByName(name string) (int32, error) {
	topic := "delNodeAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, nodeAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) addContractAdminByAddress(addr common.Address) (int32, error) {
	topic := "addContractAdminByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, contractAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}
func (u *UserManagement) addContractAdminByName(name string) (int32, error) {
	topic := "addContractAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, contractAdmin, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}
func (u *UserManagement) delContractAdminByAddress(addr common.Address) (int32, error) {
	topic := "delContractAdminByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, contractAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}
func (u *UserManagement) delContractAdminByName(name string) (int32, error) {
	topic := "delContractAdminByName"
	if err := u.setRoleWithPermissionCheckByName(name, contractAdmin, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) addContractDeployerByAddress(addr common.Address) (int32, error) {
	topic := "addContractDeployerByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, contractDeployer, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}
func (u *UserManagement) addContractDeployerByName(name string) (int32, error) {
	topic := "addContractDeployerByName"
	if err := u.setRoleWithPermissionCheckByName(name, contractDeployer, roleActive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) delContractDeployerByAddress(addr common.Address) (int32, error) {
	topic := "delContractDeployerByAddress"
	if err := u.setRoleWithPermissionCheckByAddress(addr, contractDeployer, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) delContractDeployerByName(name string) (int32, error) {
	topic := "delContractDeployerByName"
	if err := u.setRoleWithPermissionCheckByName(name, contractDeployer, roleDeactive); err != nil {
		return u.returnFail(topic, err)
	}
	return u.returnSuccess(topic)
}

func (u *UserManagement) getRolesByName(name string) (string, error) {
	addr, err := u.getAddrByName(name)
	if err != nil {
		return err.Error(), err
	}
	return u.getRolesByAddress(addr)
}

func (u *UserManagement) getRolesByAddress(addr common.Address) (string, error) {
	ur, err := u.getRole(addr)
	if err != nil {
		return "", err
	}
	roles := ur.Strings()
	str, err := json.Marshal(roles)
	if err != nil {
		return "", err
	}
	return string(str), nil
}
func (u *UserManagement) getAddrListOfRoleStr(targetRole string) (string, error) {
	if role, ok := rolesMap[targetRole]; ok {
		return u.getAddrListOfRole(role)
	}
	return fmt.Sprintf("Unsupported Role: %s", targetRole), errUnsupportedRole
}
func (u *UserManagement) getAddrListOfRole(targetRole int32) (string, error) {
	var key []byte
	var err error
	addrs := make([]common.Address, 0)

	if key, err = generateAddressListKey(targetRole); err != nil {
		return "", err
	}
	if addrs, err = u.getAddrList(key); err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "[]", nil
	}
	str, err := json.Marshal(addrs)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func (u *UserManagement) hasRole(addr common.Address, roleName string) (int32, error) {
	ur, err := u.getRole(addr)
	if err != nil {
		return roleDeactive, err
	}
	if role, ok := rolesMap[roleName]; ok && ur.hasRole(role) {
		return roleActive, nil
	}
	return roleDeactive, nil
}

//internal function
func (u *UserManagement) getRole(addr common.Address) (UserRoles, error) {
	key := generateRoleKey(addr)
	data := u.getState(key)
	if len(data) == 0 {
		return UserRoles(0), nil
	}

	// ur, err := RetrieveUserRoles(data)
	ur := UserRoles(0)
	if err := rlp.DecodeBytes(data, &ur); nil != err {
		return UserRoles(0), err
	}
	return ur, nil
}
func (u *UserManagement) setRole(addr common.Address, roles UserRoles) error {
	data, err := rlp.EncodeToBytes(roles)
	if err != nil {
		return err
	}
	key := generateRoleKey(addr)
	u.setState(key, data)
	return nil
}
func (u *UserManagement) setRoleWithPermissionCheckByName(name string, targetRole int32, status uint8) error {
	addr, err := u.getAddrByName(name)
	if err != nil {
		return err
	}
	if addr == ZeroAddress {
		return errNoUserInfo
	}
	return u.setRoleWithPermissionCheckByAddress(addr, targetRole, status)
}

func (u *UserManagement) setRoleWithPermissionCheckByAddress(addr common.Address, targetRole int32, status uint8) error {
	// 调用者权限判断
	caller := u.Caller()
	callerRole, err := u.getRole(caller)
	if err != nil {
		return err
	}

	permissionRoles := roleOpPermission[targetRole]
	if permissionRoles&callerRole == 0 {
		return errNoPermission
	}

	ur, err := u.getRole(addr)
	if err != nil {
		return err
	}

	if status == roleActive {
		if err := ur.setRole(targetRole); err != nil {
			return err
		}
		if err := u.addAddrListOfRole(addr, targetRole); err != nil {
			return err
		}
	} else if status == roleDeactive {
		if err := ur.unsetRole(targetRole); err != nil {
			return err
		}
		if err := u.delAddrListOfRole(addr, targetRole); err != nil {
			return err
		}
	}

	if err := u.setRole(addr, ur); err != nil {
		return err
	}
	return nil
}

func (u *UserManagement) addAddrListOfRole(addr common.Address, targetRole int32) error {
	var key []byte
	var err error

	if key, err = generateAddressListKey(targetRole); err != nil {
		return err
	}
	if err = u.addAddrList(key, addr); err != nil {
		return err
	}
	return nil
}

func (u *UserManagement) delAddrListOfRole(addr common.Address, targetRole int32) error {
	var key []byte
	var err error

	if key, err = generateAddressListKey(targetRole); err != nil {
		return err
	}

	if err = u.delAddrList(key, addr); err != nil {
		return err
	}
	return nil
}

func (u *UserManagement) addAddrList(key []byte, addr common.Address) error {
	addrs, err := u.getAddrList(key)
	if err != nil {
		return err
	}
	for _, v := range addrs {
		if v == addr {
			return nil
		}
	}
	addrs = append(addrs, addr)

	if err := u.setAddrList(key, addrs); err != nil {
		return err
	}
	return nil
}
func (u *UserManagement) delAddrList(key []byte, addr common.Address) error {
	addrs, err := u.getAddrList(key)
	if err != nil {
		return err
	}
	pos := -1
	for i, v := range addrs {
		if v == addr {
			pos = i
			break
		}
	}
	if pos != -1 {
		addrs = append(addrs[0:pos], addrs[pos+1:]...)
		if err := u.setAddrList(key, addrs); err != nil {
			return err
		}
	}
	return nil
}
func (u *UserManagement) getAddrList(key []byte) ([]common.Address, error) {
	var addrs []common.Address

	data := u.getState(key)
	if len(data) == 0 {
		return addrs, nil
	}

	if err := json.Unmarshal(data, &addrs); err != nil {
		return nil, err
	}

	return addrs, nil
}
func (u *UserManagement) setAddrList(key []byte, addrs []common.Address) error {
	data, err := json.Marshal(addrs)
	if err != nil {
		return err
	}
	u.setState(key, data)
	return nil
}

func generateRoleKey(addr common.Address) []byte {
	return generateStateKey(addr.String() + userRolesKey)
}

func generateAddressListKey(targetRole int32) ([]byte, error) {
	var key []byte
	switch targetRole {
	case superAdmin:
		key = superAdminAddrListKey
	case chainAdmin:
		key = chainAdminAddrListKey
	case groupAdmin:
		key = groupAdminAddrListKey
	case nodeAdmin:
		key = nodeAdminAddrListKey
	case contractAdmin:
		key = contractAdminAddrListKey
	case contractDeployer:
		key = contractDeployerAddrListKey
	default:
		return nil, errUnsupportedRole
	}
	return key, nil
}
