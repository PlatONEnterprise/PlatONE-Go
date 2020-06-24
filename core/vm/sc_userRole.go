package vm

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	SUPER_ADMIN = iota
	CHAIN_ADMIN
	NODE_ADMIN
	CONTRACT_ADMIN
	CONTRACT_DEPLOYER
	ROLES_CNT
)

var (
	rolesName = []string{
		"SUPER_ADMIN",
		"CHAIN_ADMIN",
		"NODE_ADMIN",
		"CONTRACT_ADMIN",
		"CONTRACT_DEPLOYER",
	}
	rolesMap = map[string]int32{
		"SUPER_ADMIN":       SUPER_ADMIN,
		"CHAIN_ADMIN":       CHAIN_ADMIN,
		"NODE_ADMIN":        NODE_ADMIN,
		"CONTRACT_ADMIN":    CONTRACT_ADMIN,
		"CONTRACT_DEPLOYER": CONTRACT_DEPLOYER,
	}
)

const (
	ROLE_DEACTIVE = 0
	ROLE_ACTIVE   = 1
)

const (
	// UserRolesKey = sha3("userRoles")
	UserRolesKey = "45332c40978a610bce7447f52a7278f5"
	// AddressListKey = sha3("addressList")
	AddressListKey = "5fec39173ed9a584db9d70c832080d80"
)

var (
	ErrUnsupportedRole      = errors.New("Unsupported role")
	ErrNoPermission         = errors.New("No Permmision")
	ErrAlreadySetSuperAdmin = errors.New("Already Set SuperAdmin")
)

var (
	// 操作某个权限角色所需要的权限
	// permissionRoles[NODE_ADMIN]表示操作NODE_ADMIN所需的权限，为CHAIN_ADMIN
	permissionRolesMap map[int32][]int32
)

func init() {
	permissionRolesMap = make(map[int32][]int32, ROLES_CNT)
	permissionRolesMap[SUPER_ADMIN] = []int32{SUPER_ADMIN}
	permissionRolesMap[CHAIN_ADMIN] = []int32{SUPER_ADMIN}
	permissionRolesMap[NODE_ADMIN] = []int32{CHAIN_ADMIN}
	permissionRolesMap[CONTRACT_ADMIN] = []int32{CHAIN_ADMIN}
	permissionRolesMap[CONTRACT_DEPLOYER] = []int32{CHAIN_ADMIN, CONTRACT_ADMIN}
}

type UserRoles struct {
	roles uint32
}

func (ur *UserRoles) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, ur.roles)
}

func (ur *UserRoles) DecodeRLP(r io.Reader) error {
	return rlp.Decode(r, ur.roles)
}

func (ur *UserRoles) Strings() []string {
	roles := []string{}

	for i := int32(0); i < ROLES_CNT; i++ {
		if ur.hasRole(i) {
			roles = append(roles, rolesName[i])
		}
	}
	return roles
}

func (ur *UserRoles) FromStrings(rolesStr []string) {
	ur.roles = 0
	for _, s := range rolesStr {
		if role, ok := rolesMap[s]; ok {
			ur.setRole(role)
		}
	}
}

func (ur *UserRoles) setRole(role int32) error {
	if role >= ROLES_CNT || role < 0 {
		return ErrUnsupportedRole
	}
	ur.roles |= 1 << role
	return nil
}

func (ur *UserRoles) unsetRole(role int32) error {
	if role >= ROLES_CNT || role < 0 {
		return ErrUnsupportedRole
	}
	ur.roles &= ^(1 << role)

	return nil
}

func (ur UserRoles) hasRole(role int32) bool {
	if role >= ROLES_CNT || role < 0 {
		return false
	}
	return ur.roles&(1<<role) != 0
}

// decode
func RetrieveUserRoles(data []byte) (*UserRoles, error) {
	ur := &UserRoles{}
	if err := rlp.DecodeBytes(data, ur); nil != err {
		return nil, err
	}

	return ur, nil
}

// export function
func (u *UserManagement) setSuperAdmin() ([]byte, error) {
	var key []byte
	var err error
	var addrs []common.Address

	if key, err = generateAddressListKey(SUPER_ADMIN); err != nil {
		return nil, err
	}
	if addrs, err = u.getAddrList(key); err != nil {
		return nil, err
	}

	if len(addrs) >= 1 {
		return nil, ErrAlreadySetSuperAdmin
	}

	ur := &UserRoles{}
	ur.setRole(SUPER_ADMIN)

	u.addAddrListOfRole(u.Caller(), SUPER_ADMIN)
	if err := u.setRole(u.Caller(), ur); err != nil {
		return nil, err
	}

	return nil, nil
}

func (u *UserManagement) transferSuperAdminByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, SUPER_ADMIN, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	if err := u.setRoleWithPermissionCheckByAddress(u.Caller(), SUPER_ADMIN, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}
func (u *UserManagement) transferSuperAdminByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, SUPER_ADMIN, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	if err := u.setRoleWithPermissionCheckByAddress(u.Caller(), SUPER_ADMIN, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) addChainAdminByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, CHAIN_ADMIN, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) addChainAdminByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, CHAIN_ADMIN, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) delChainAdminByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, CHAIN_ADMIN, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) delChainAdminByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, CHAIN_ADMIN, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) addNodeAdminByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, NODE_ADMIN, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) addNodeAdminByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, NODE_ADMIN, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) delNodeAdminByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, NODE_ADMIN, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) delNodeAdminByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, NODE_ADMIN, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) addContractAdminByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, CONTRACT_ADMIN, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}
func (u *UserManagement) addContractAdminByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, CONTRACT_ADMIN, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}
func (u *UserManagement) delContractAdminByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, CONTRACT_ADMIN, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}
func (u *UserManagement) delContractAdminByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, CONTRACT_ADMIN, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) addContractDeployerByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, CONTRACT_DEPLOYER, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}
func (u *UserManagement) addContractDeployerByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, CONTRACT_DEPLOYER, ROLE_ACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) delContractDeployerByAddress(addr common.Address) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByAddress(addr, CONTRACT_DEPLOYER, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) delContractDeployerByName(name string) ([]byte, error) {
	if err := u.setRoleWithPermissionCheckByName(name, CONTRACT_DEPLOYER, ROLE_DEACTIVE); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserManagement) getRolesByAddress(addr common.Address) ([]byte, error) {
	ur, err := u.getRole(addr)
	if err != nil {
		return nil, err
	}

	roles := ur.Strings()
	return json.Marshal(roles)
}

func (u *UserManagement) getAddrListOfRole(targetRole int32) ([]byte, error) {
	var key []byte
	var err error
	var addrs []common.Address

	if key, err = generateAddressListKey(targetRole); err != nil {
		return nil, err
	}
	if addrs, err = u.getAddrList(key); err != nil {
		return nil, err
	}
	return json.Marshal(addrs)
}

//internal function
func (u *UserManagement) getRole(addr common.Address) (*UserRoles, error) {
	key := append(addr[:], []byte(UserRolesKey)...)
	data := u.getState(key)
	if len(data) == 0 {
		return &UserRoles{}, nil
	}

	// ur, err := RetrieveUserRoles(data)
	roles := uint32(0)
	if err := rlp.DecodeBytes(data, &roles); nil != err {
		return nil, err
	}
	ur := &UserRoles{roles: roles}

	return ur, nil
}
func (u *UserManagement) setRole(addr common.Address, roles *UserRoles) error {
	data, err := rlp.EncodeToBytes(roles)
	if err != nil {
		return err
	}
	key := append(addr[:], []byte(UserRolesKey)...)
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

	permissionRoles := permissionRolesMap[targetRole]
	for _, permissionRole := range permissionRoles {
		if !callerRole.hasRole(permissionRole) {
			return ErrNoPermission
		}
	}

	ur, err := u.getRole(addr)
	if err != nil {
		return err
	}

	if status == ROLE_ACTIVE {
		ur.setRole(targetRole)
		u.addAddrListOfRole(addr, targetRole)
	} else if status == ROLE_DEACTIVE {
		ur.unsetRole(targetRole)
		u.delAddrListOfRole(addr, targetRole)
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

	u.setAddrList(key, addrs)
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
	data := u.getState(key)

	if len(data) == 0 {
		return []common.Address{}, nil
	}

	addrs := []common.Address{}
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
	key := AddressListKey
	switch targetRole {
	case SUPER_ADMIN:
		key += "SUPER_ADMIN"
	case CHAIN_ADMIN:
		key += "CHAIN_ADMIN"
	case NODE_ADMIN:
		key += "NODE_ADMIN"
	case CONTRACT_ADMIN:
		key += "CONTRACT_ADMIN"
	case CONTRACT_DEPLOYER:
		key += "CONTRACT_ADMIN"
	default:
		return nil, ErrUnsupportedRole
	}
	return []byte(key), nil
}
