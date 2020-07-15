package vm

import (
	"encoding/json"

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
	var roles []string

	for i := int32(0); i < rolesCnt; i++ {
		if ur.hasRole(i) {
			roles = append(roles, rolesName[i])
		}
	}
	return roles
}

func (ur *UserRoles) FromStrings(rolesStr []string) error {
	for _, s := range rolesStr {
		if role, ok := rolesMap[s]; ok {
			if err := ur.setRole(role); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ur *UserRoles) setRole(role int32) error {
	if role >= rolesCnt || role < 0 {
		return ErrUnsupportedRole
	}
	*ur |= 1 << role
	return nil
}

func (ur *UserRoles) unsetRole(role int32) error {
	if role >= rolesCnt || role < 0 {
		return ErrUnsupportedRole
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

// export function
func (u *UserManagement) setSuperAdmin() (int32, error) {
	var key []byte
	var err error
	var addrs []common.Address

	if key, err = generateAddressListKey(superAdmin); err != nil {
		return -1, err
	}
	if addrs, err = u.getAddrList(key); err != nil {
		return -1, err
	}

	if len(addrs) >= 1 {
		return -1, ErrAlreadySetSuperAdmin
	}

	ur := UserRoles(0)
	if err := ur.setRole(superAdmin); err != nil {
		return -1, err
	}

	if err := u.addAddrListOfRole(u.Caller(), superAdmin); err != nil {
		return -1, err
	}
	if err := u.setRole(u.Caller(), ur); err != nil {
		return -1, err
	}

	return 0, nil
}

func (u *UserManagement) transferSuperAdminByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, superAdmin, roleActive); err != nil {
		return -1, err
	}
	if err := u.setRoleWithPermissionCheckByAddress(u.Caller(), superAdmin, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}
func (u *UserManagement) transferSuperAdminByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, superAdmin, roleActive); err != nil {
		return -1, err
	}
	if err := u.setRoleWithPermissionCheckByAddress(u.Caller(), superAdmin, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) addChainAdminByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, chainAdmin, roleActive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) addChainAdminByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, chainAdmin, roleActive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) delChainAdminByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, chainAdmin, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) delChainAdminByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, chainAdmin, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) addNodeAdminByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, nodeAdmin, roleActive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) addNodeAdminByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, nodeAdmin, roleActive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) delNodeAdminByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, nodeAdmin, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) delNodeAdminByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, nodeAdmin, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) addContractAdminByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, contractAdmin, roleActive); err != nil {
		return -1, err
	}
	return 0, nil
}
func (u *UserManagement) addContractAdminByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, contractAdmin, roleActive); err != nil {
		return -1, err
	}
	return 0, nil
}
func (u *UserManagement) delContractAdminByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, contractAdmin, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}
func (u *UserManagement) delContractAdminByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, contractAdmin, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) addContractDeployerByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, contractDeployer, roleActive); err != nil {
		return -1, err
	}
	return 0, nil
}
func (u *UserManagement) addContractDeployerByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, contractDeployer, roleActive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) delContractDeployerByAddress(addr common.Address) (int32, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, contractDeployer, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) delContractDeployerByName(name string) (int32, error) {
	if err := u.setRoleWithPermissionCheckByName(name, contractDeployer, roleDeactive); err != nil {
		return -1, err
	}
	return 0, nil
}

func (u *UserManagement) getRolesByName(name string) (string, error) {
	addr := u.getAddrByName(name)
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
	return "", ErrUnsupportedRole
}
func (u *UserManagement) getAddrListOfRole(targetRole int32) (string, error) {
	var key []byte
	var err error
	var addrs []common.Address

	if key, err = generateAddressListKey(targetRole); err != nil {
		return "", err
	}
	if addrs, err = u.getAddrList(key); err != nil {
		return "", err
	}
	str, err := json.Marshal(addrs)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

// 0： the account(addr) doesn't has role
// 1: the account(addr) doesn't role
func (u *UserManagement) hasRole(addr common.Address, roleName string) (int32, error) {
	key := append(addr[:], []byte(userRolesKey)...)
	data := u.getState(key)
	if len(data) == 0 {
		return 0, nil
	}

	ur := UserRoles(0)
	if err := rlp.DecodeBytes(data, &ur); nil != err {
		return 0, nil
	}

	if role, ok := rolesMap[roleName]; ok && ur.hasRole(role) {
		return 1, nil
	}

	return 0, nil
}

//internal function
func (u *UserManagement) getRole(addr common.Address) (UserRoles, error) {
	key := append(addr[:], []byte(userRolesKey)...)
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
	key := append(addr[:], []byte(userRolesKey)...)
	u.setState(key, data)
	return nil
}
func (u *UserManagement) setRoleWithPermissionCheckByName(name string, targetRole int32, status uint8) error {
	addr := u.getAddrByName(name)
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
		return ErrNoPermission
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

func generateAddressListKey(targetRole int32) ([]byte, error) {
	key := addressListKey
	switch targetRole {
	case superAdmin:
		key += "superAdmin"
	case chainAdmin:
		key += "chainAdmin"
	case nodeAdmin:
		key += "nodeAdmin"
	case contractAdmin:
		key += "contractAdmin"
	case contractDeployer:
		key += "contractAdmin"
	default:
		return nil, ErrUnsupportedRole
	}
	return []byte(key), nil
}
