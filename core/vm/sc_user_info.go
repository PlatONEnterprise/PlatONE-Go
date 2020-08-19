package vm

import (
	"encoding/json"
	"errors"

	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	// userInfoKey = sha3("userInfo")
	userInfoKey = "bebc082d6e6b7cce489569e3716ad317"
	// userAddressMapKey = sha3("userAddressMap")
	userAddressMapKey = "25e73d394205b4eb66cfddc8c77e0e6e"
	// AddressUserMap = sha3("addressUserMap")
	addressUserMapKey = "a3b6f5702a6c188c1558e4f8bb686c56"
	// userListKey = sha3("userList")
	userListKey = "5af8141e0ecb4e3df3f35f6e6b0b387b"
)

var (
	errUserNameAlreadyExist   = errors.New(" UserName Already Exist ")
	errAlreadySetUserName     = errors.New("Already Set UserName ")
	errNoUserInfo             = errors.New("No User Info ")
)

type UserInfo = syscontracts.UserInfo
type DescInfo = syscontracts.UserDescInfo

func checkDescInfo(descInfo DescInfo) (bool, error) {
	if descInfo.Organization != "" {
		if b, err := checkNameFormat(descInfo.Organization); !b || err != nil {
			return b, errOrgnizationUnsupported
		}
	}
	if descInfo.Email != "" {
		if b, err := checkEmailFormat(descInfo.Email); !b || err != nil {
			return b, errEmailUnsupported
		}
	}

	if descInfo.Phone != "" {
		if b, err := checkPhoneFormat(descInfo.Phone); !b || err != nil {
			return b, errPhoneUnsupported
		}
	}
	return true, nil
}

func isZeroDescInfo(descInfo DescInfo) bool {
	return descInfo.Organization == "" && descInfo.Email == "" && descInfo.Phone == ""
}

func updateDescInfo(src *DescInfo, dest *DescInfo) {
	if dest.Email != "" && src.Email != dest.Email {
		src.Email = dest.Email
	}
	if dest.Organization != "" && src.Organization != dest.Organization {
		src.Organization = dest.Organization
	}
	if dest.Phone != "" && src.Phone != dest.Phone {
		src.Phone = dest.Phone
	}
}

// export function
// 管理员操作
func (u *UserManagement) addUser(info *UserInfo) (int32, error) {
	topic := "addUser"
	if info.Address == ZeroAddress{
		return u.returnFail(topic, errors.New("User's Address must be a nonzero address!"))
	}

	if !u.callerPermissionCheck() {
		return u.returnFail(topic, errNoPermission)
	}

	addr, err := u.getAddrByName(info.Name)
	if err != nil && err != errNoUserInfo {
		return u.returnFail(topic, err)
	}

	if addr != ZeroAddress {
		return u.returnFail(topic, errUserNameAlreadyExist)
	}

	// addr := common.HexToAddress(info.Address)
	addr = info.Address
	if u.getNameByAddr(addr) != "" {
		return u.returnFail(topic, errAlreadySetUserName)
	}

	if info.DescInfo != "" {
		descInfo := &DescInfo{}
		if err := json.Unmarshal([]byte(info.DescInfo), descInfo); err != nil {
			log.Error("json.Unmarshal([]byte(info.DescInfo), descInfo)")
			return u.returnFail(topic, err)
		}

		if b, err := checkDescInfo(*descInfo); !b || err != nil {
			return u.returnFail(topic, err)
		}

		if data, err := json.Marshal(descInfo); err != nil{
			return u.returnFail(topic, err)
		}else {
			info.DescInfo = string(data)
		}
	}

	info.Authorizer = u.Caller()

	if err := u.setUserInfo(info); err != nil {
		return u.returnFail(topic, err)
	}
	u.addAddrNameMap(addr, info.Name)
	u.addNameAddrMap(addr, info.Name)

	if err := u.addUserList(addr); err != nil {
		return u.returnFail(topic, err)
	}

	return u.returnSuccess(topic)
}

// 管理员操作，可以更新用户信息中的DescInfo字段
func (u *UserManagement) updateUserDescInfo(addr common.Address, info *DescInfo) (int32, error) {
	topic := "updateUserDescInfo"
	if !u.callerPermissionCheck() {
		return u.returnFail(topic, errNoPermission)
	}

	if b, err := checkDescInfo(*info); !b || err != nil {
		return u.returnFail(topic, err)
	}

	userInfo, err := u.getUserInfo(addr)
	if err != nil {
		return u.returnFail(topic, err)
	}

	infoOnChain := &DescInfo{}
	if userInfo.DescInfo != "" {
		if err := json.Unmarshal([]byte(userInfo.DescInfo), infoOnChain); err != nil {
			return u.returnFail(topic, err)
		}
	}
	updateDescInfo(infoOnChain, info)

	ser, err := json.Marshal(infoOnChain)
	userInfo.DescInfo = string(ser)

	if err := u.setUserInfo(userInfo); err != nil {
		return u.returnFail(topic, err)
	}

	return u.returnSuccess(topic)
}

// 查询用户信息，任意用户可查
func (u *UserManagement) getUserByAddress(addr common.Address) ([]byte, error) {
	user, err := u.getUserInfo(addr)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (u *UserManagement) getUserByName(name string) (string, error) {
	if b, err := checkNameFormat(name); err != nil || !b {
		return errUsernameUnsupported.Error(), errUsernameUnsupported
	}

	addr, err := u.getAddrByName(name)
	if err != nil {
		return err.Error(), err
	}
	if addr == ZeroAddress {
		return errNoUserInfo.Error(), errNoUserInfo
	}

	user, err := u.getUserInfo(addr)
	if err != nil {
		return err.Error(), err
	}

	data, err := json.Marshal(user)
	if err != nil {
		return err.Error(), err
	}
	return string(data), nil
}

// 查询登记的所有用户，任意用户可查
func (u *UserManagement) getAllUsers() ([]byte, error) {
	var (
		addrs []common.Address
		users []*UserInfo
		err   error
	)
	addrs, err = u.getUserList()
	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		return []byte("[]"), nil
	}

	for _, v := range addrs {
		info, err := u.getUserInfo(v)
		if err != nil {
			return nil, err
		}
		users = append(users, info)
	}

	return json.Marshal(users)
}

//internal function
func (u *UserManagement) setUserInfo(info *UserInfo) error {
	// addr := common.HexToAddress(info.Address)
	addr := info.Address
	key1 := append(addr[:], []byte(userInfoKey)...)
	data, err := rlp.EncodeToBytes(info)
	if err != nil {
		return err
	}
	u.setState(key1, data)

	return nil
}
func (u *UserManagement) getUserInfo(addr common.Address) (*UserInfo, error) {
	key := append(addr[:], []byte(userInfoKey)...)
	data := u.getState(key)
	if len(data) == 0 {
		return nil, errNoUserInfo
	}

	info := &UserInfo{}
	if err := rlp.DecodeBytes(data, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (u *UserManagement) addAddrNameMap(addr common.Address, name string) {
	key := append(addr[:], []byte(addressUserMapKey)...)
	u.setState(key, []byte(name))
}
func (u *UserManagement) getNameByAddr(addr common.Address) string {
	key := append(addr[:], []byte(addressUserMapKey)...)
	data := u.getState(key)
	return string(data)
}
func (u *UserManagement) addNameAddrMap(addr common.Address, name string) {
	key := append([]byte(name), []byte(userAddressMapKey)...)
	u.setState(key, addr[:])
}
func (u *UserManagement) getAddrByName(name string) (common.Address, error) {
	addr := common.Address{}
	if b, err := checkNameFormat(name); err != nil || !b {
		return addr, errUsernameUnsupported
	}

	key := append([]byte(name), []byte(userAddressMapKey)...)
	data := u.getState(key)
	if len(data) == 0 {
		return addr, errNoUserInfo
	}
	if len(data) == 20 {
		copy(addr[:], data)
	}
	return addr, nil
}

func (u *UserManagement) addUserList(addr common.Address) error {
	var addrs []common.Address
	var err error
	if addrs, err = u.getUserList(); err != nil {
		return err
	}

	for _, v := range addrs {
		if v == addr {
			return nil
		}
	}
	addrs = append(addrs, addr)

	if err = u.setUserList(addrs); err != nil {
		return err
	}
	return nil
}

func (u *UserManagement) delUserList(addr common.Address) error {
	var (
		addrs []common.Address
		err   error
	)
	if addrs, err = u.getUserList(); err != nil {
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
		addrs = append(addrs[:pos], addrs[pos+1:]...)
		if err := u.setUserList(addrs); err != nil {
			return err
		}
	}
	return nil
}

func (u *UserManagement) getUserList() ([]common.Address, error) {
	var addrs []common.Address

	key := []byte(userListKey)
	data := u.getState(key)
	if len(data) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(data, &addrs); err != nil {
		return nil, err
	}

	return addrs, nil
}

func (u *UserManagement) setUserList(addrs []common.Address) error {
	data, err := json.Marshal(addrs)
	if err != nil {
		return err
	}

	key := []byte(userListKey)
	u.setState(key, data)
	return nil
}

func (u *UserManagement) callerPermissionCheck() bool {
	return hasUserOpPermission(u.stateDB, u.caller)
}
