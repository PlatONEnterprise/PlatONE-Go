package vm

import (
	"encoding/json"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	// UserInfoKey = sha3("userInfo")
	UserInfoKey = "bebc082d6e6b7cce489569e3716ad317"
	// UserAddressMapKey = sha3("userAddressMap")
	UserAddressMapKey = "25e73d394205b4eb66cfddc8c77e0e6e"
	// AddressUserMap = sha3("addressUserMap")
	AddressUserMapKey = "a3b6f5702a6c188c1558e4f8bb686c56"
	// UserListKey = sha3("userList")
	UserListKey = "5af8141e0ecb4e3df3f35f6e6b0b387b"
)

// 具备操作user权限的角色
var permissionRolesForUserOps  = []int32{CHAIN_ADMIN}

type UserInfo struct {
	Address common.Address  `json:"address"`
	Name string				`json:"name"`
	MiscInfo string			`json:"miscInfo"`
	Version uint32			`json:"version"`
}

func (u *UserInfo) encode() ([]byte, error) {
	return rlp.EncodeToBytes(u)
}

//decode
func RetrieveUserInfo(data []byte) (*UserInfo, error) {
	var ui UserInfo
	if err := rlp.DecodeBytes(data, &ui); nil != err {
		return nil, err
	}

	return &ui, nil
}

func (u *UserManagement) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.UserManagementGas
}

// export function
func (u *UserManagement) addUser(userInfo string) ([]byte, error){
	if !u.callerPermissionCheck() {
		return nil, ErrNoPermission
	}

	info := &UserInfo{}

	if err := json.Unmarshal([]byte(userInfo), info); err != nil{
		return nil,err
	}

	if err := u.setUserInfo(info); err != nil{
		return nil, err
	}

	if err := u.addUserList(info.Address); err != nil{
		return nil, err
	}

	return nil,nil
}
func (u *UserManagement) updateUser(userInfo string)  ([]byte, error){
	if !u.callerPermissionCheck() {
		return nil, ErrNoPermission
	}

	info := &UserInfo{}

	if err := json.Unmarshal([]byte(userInfo), info); err != nil{
		return nil,err
	}

	if err := u.setUserInfo(info); err != nil{
		return nil, err
	}

	return nil, nil
}
func (u *UserManagement) getUser(addr common.Address) ([]byte, error){
	user, err := u.getUserInfo(addr)
	if err != nil{
		return nil, err
	}

	data, err := json.Marshal(user)
	if err != nil{
		return nil, err
	}
	return data, nil
}

func (u *UserManagement) getAllUser() []*UserInfo{
	var (
		addrs []common.Address
		users []*UserInfo
		err error
	)
	addrs, err = u.getUserList()
	if err != nil{
		return nil
	}

	for  _,v := range addrs{
		info, err := u.getUserInfo(v)
		if err != nil{
			return nil
		}
		users = append(users, info)
	}

	return users
}

//func (u *UserManagement) registerUserInfo(userInfo string) {}
//func (u *UserManagement) approveUserInfo(userAddress string){}


//func (u *UserManagement) getUsersByRole()

//internal function
func (u *UserManagement) setUserInfo(info *UserInfo) error{
	addr := info.Address
	key1 := append(addr[:], []byte(UserInfoKey)...)
	data, err := rlp.EncodeToBytes(info)
	if err != nil{
		return err
	}
	u.setState(key1, data)

	return nil
}
func (u *UserManagement) getUserInfo(addr common.Address) (*UserInfo, error){
	key := append(addr[:], []byte(UserInfoKey)...)
	data := u.getState(key)

	info := &UserInfo{}
	if err := rlp.DecodeBytes(data, info); err != nil{
		return nil, err
	}
	return info, nil
}


func (u *UserManagement) addUserList(addr common.Address)error{
	var addrs []common.Address
	var err error
	if addrs, err = u.getUserList(); err != nil {
		return err
	}

	for _,v := range addrs{
		if v == addr {
			return nil
		}
	}
	addrs = append(addrs, addr)

	if err = u.setUserList(addrs); err != nil{
		return err
	}
	return nil
}

func (u *UserManagement) delUserList(addr common.Address)error{
	var (
		addrs []common.Address
		err error
	)
	if addrs, err = u.getUserList(); err!=nil{
		return err
	}

	pos := -1
	for i,v := range addrs{
		if v == addr {
			pos = i
			break
		}
	}
	if pos != -1{
		addrs = append(addrs[:pos], addrs[pos+1:]...)
		if err := u.setUserList(addrs); err != nil{
			return err
		}
	}
	return nil
}

func (u *UserManagement) getUserList()([]common.Address, error){
	var addrs []common.Address

	key := []byte(UserListKey)
	data := u.getState(key)

	if err := rlp.DecodeBytes(data, addrs); err != nil{
		return nil, err
	}

	return addrs, nil
}

func (u *UserManagement) setUserList(addrs []common.Address)error{
	data, err := rlp.EncodeToBytes(addrs)
	if err != nil{
		return err
	}

	key := []byte(UserListKey)
	u.setState(key, data)
	return nil
}

func (u *UserManagement) callerPermissionCheck() bool{
	caller := u.Caller()
	role, err := u.getRolesByAddress(caller)
	if err != nil{
		return false
	}
	for _, p := range permissionRolesForUserOps {
		if role. hasRole(p){
			return true
		}
	}

	return false
}


