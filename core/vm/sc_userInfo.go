package vm

import (
	"encoding/json"
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
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

var (
	ZeroAddress = common.Address{}
)

var (
	errUsernameUnsupported = errors.New("Unsupported Username")
	errUserNameAlreadyExist = errors.New(" UserName Already Exist")
	errAlreadySetUserName = errors.New("Already Set UserName")
	errNoUserInfo = errors.New("No UserInfo")
)

// 具备操作user权限的角色
var permissionRolesForUserOps  = []int32{CHAIN_ADMIN}

type UserInfo struct {
	Address common.Address  	`json:"address"`  	// 地址，不可变更
	Authorizer common.Address 	`json:"authorizer"`	// 授权者，不可变更
	Name string					`json:"name"`		// 用户名，不可变更

	DescInfo string				`json:"descInfo"`	// 描述信息，可变更
	Version uint32				`json:"version"`	// 可变更
}

type DescInfoV1 struct{
	Email        string 	`json:"email"`
	Organization string 	`json:"organization"`
	Phone string 			`json:"phone"`
}

func (src *DescInfoV1) update(dest *DescInfoV1){
	if src.Email != dest.Email{
		src.Email = dest.Email
	}
	if src.Organization != dest.Organization {
		src.Organization = dest.Organization
	}
	if src.Phone != dest.Phone{
		src.Phone = dest.Phone
	}
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

// export function
// 管理员操作
func (u *UserManagement) addUser(userInfo string) ([]byte, error){
	if !u.callerPermissionCheck() {
		return nil, ErrNoPermission
	}

	info := &UserInfo{}

	if err := json.Unmarshal([]byte(userInfo), info); err != nil{
		return nil, err
	}

	if info.Name == "" {
		return nil, errUsernameUnsupported
	}

	if u.getAddrByName(info.Name) != ZeroAddress {
		return nil, errUserNameAlreadyExist
	}

	if u.getNameByAddr(info.Address) != ""{
		return nil, errAlreadySetUserName
	}

	descInfo := &DescInfoV1{}
	if err := json.Unmarshal([]byte(info.DescInfo), descInfo); err!= nil{
		return nil,err
	}

	info.Authorizer = u.Caller()

	if err := u.setUserInfo(info); err != nil{
		return nil, err
	}
	u.addAddrNameMap(info.Address, info.Name)
	u.addNameAddrMap(info.Address, info.Name)

	if err := u.addUserList(info.Address); err != nil{
		return nil, err
	}

	return nil,nil
}

// 管理员操作，可以更新用户信息中的DescInfo字段
func (u *UserManagement) updateUserDescInfo(addr common.Address, descInfo string)  ([]byte, error){
	if !u.callerPermissionCheck() {
		return nil, ErrNoPermission
	}

	info := &DescInfoV1{}
	if err := json.Unmarshal([]byte(descInfo), info); err != nil{
		return nil,err
	}

	userInfo, err := u.getUserInfo(addr)
	if err != nil{
		return nil, err
	}

	infoOnChain := &DescInfoV1{}
	if err := json.Unmarshal([]byte(userInfo.DescInfo), infoOnChain); err != nil{
		return nil,err
	}
	infoOnChain.update(info)

	ser, err := json.Marshal(infoOnChain)
	userInfo.DescInfo = string(ser)

	if err := u.setUserInfo(userInfo); err != nil{
		return nil, err
	}

	return nil, nil
}

// 查询用户信息，任意用户可查
func (u *UserManagement) getUserByAddress(addr common.Address) ([]byte, error){
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
func (u *UserManagement) getUserByName(name string) ([]byte, error){
	addr := u.getAddrByName(name)
	if addr == ZeroAddress {
		return nil,errNoUserInfo
	}

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

// 查询登记的所有用户，任意用户可查
func (u *UserManagement) getAllUsers() ([]byte, error){
	var (
		addrs []common.Address
		users []*UserInfo
		err error
	)
	addrs, err = u.getUserList()
	if err != nil{
		return nil, err
	}

	for  _,v := range addrs{
		info, err := u.getUserInfo(v)
		if err != nil{
			return nil, err
		}
		users = append(users, info)
	}

	return json.Marshal(users)
}
//func (u *UserManagement) registerUserInfo(userInfo string) {}
//func (u *UserManagement) approveUserInfo(userAddress string){}

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
	if len(data) == 0{
		return nil ,errNoUserInfo
	}

	info := &UserInfo{}
	if err := rlp.DecodeBytes(data, info); err != nil{
		return nil, err
	}
	return info, nil
}

func (u *UserManagement) addAddrNameMap(addr common.Address, name string){
	key := append(addr[:], []byte(AddressUserMapKey)...)
	u.setState(key, []byte(name))
}
func (u *UserManagement) getNameByAddr(addr common.Address) string {
	key := append(addr[:], []byte(AddressUserMapKey)...)
	data := u.getState(key)
	return string(data)
}
func (u *UserManagement) addNameAddrMap(addr common.Address, name string){
	key := append([]byte(name), []byte(UserAddressMapKey)...)
	u.setState(key, addr[:])
}
func (u *UserManagement) getAddrByName(name string) common.Address{
	key := append([]byte(name), []byte(UserAddressMapKey)...)
	data := u.getState(key)
	addr := common.Address{}
	if len(data) == 20 {
		copy(addr[:], data)
	}
	return addr
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

	rolesStr := []string  {}
	json.Unmarshal(role, rolesStr)

	ur := &UserRoles{}
	ur.FromStrings(rolesStr)

	for _, p := range permissionRolesForUserOps {
		if ur. hasRole(p){
			return true
		}
	}

	return false
}


