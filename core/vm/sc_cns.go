package vm

import (
	"encoding/json"
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"strings"
	"time"
)

const (
	SUCCESS = 0
	FAILURE = 1

	UNREGISTERED = 0
	REGISTERED = 1

	MSG_OK = "ok"
	MSG_ERR = "input is invalid"

	CODE_OK = 0
	CODE_ERR = 1
)

type CnsManager struct {
	callerAddr 	*common.Address	// callerAddr = Contract.CallerAddress
	cMap	 	*cnsMap			// cMap = NewCnsMap(StateDB, Contract.CodeAddr)
}

type ContractInfo struct {
	Name		string
	Version 	string
	Address 	string
	Origin  	string
	TimeStamp 	time.Time
	Enabled 	bool
}

type returnMsg struct {
	Code	int
	Msg     string
	Array	[]*ContractInfo
}

func newContractInfo(name, version, address, origin string) *ContractInfo {
	return &ContractInfo{
		Name:		name,
		Version:	version,
		Address:	address,
		Origin:		origin,
		TimeStamp:  time.Now(),
		Enabled:	true,
	}
}

func newReturnMsg(code int, msg string, arrary []*ContractInfo) *returnMsg {
	return &returnMsg{
		Code:  code,
		Msg:   msg,
		Array: arrary,
	}
}

func (c *ContractInfo) encode() ([]byte, error) {
	return rlp.EncodeToBytes(c)
}

//decode
func decodeCnsInfo(data []byte) (*ContractInfo, error) {
	var ci ContractInfo
	if err := rlp.DecodeBytes(data, &ci); nil != err {
		return nil, err
	}

	return &ci, nil
}

func (u *CnsManager) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.CnsManagerGas
}

// Run runs the precompiled contract
func (u *CnsManager) Run(input []byte) ([]byte, error) {
	return execSC(input, u.AllExportFns())
}

// for access control
func (u *CnsManager) AllExportFns() SCExportFns {
	return SCExportFns{
		"cnsRegister": 				u.cnsRegister,
		"cnsUnregister":			u.cnsUnregister,
		"getContractAddress":		u.getContractAddress,
		"ifRegisteredByAddress":	u.ifRegisteredByAddress,
		"ifRegisteredByName":		u.ifRegisteredByName,
		"getRegisteredContracts":	u.getRegisteredContracts,
	}
}

/*
func (u *CnsManager) setState(key, value []byte) {
	u.Evm.StateDB.SetState(*u.Contract.CodeAddr, key, value)
}

func (u *CnsManager) getState(key []byte) []byte {
	return u.Evm.StateDB.GetState(*u.Contract.CodeAddr, key)
}*/

func (cns *CnsManager) isOwner(contractAddr common.Address) bool {	// TODO return type int64?
	callerAddr := cns.callerAddr
	contractOwnerAddr := cns.cMap.GetContractCreator(contractAddr)

	if callerAddr.Hex() == contractOwnerAddr.Hex() {
		return true
	} else {
		return false
	}
}

// not applicable
func (cns *CnsManager) cnsRegisterFromInit(name, version, address string) ([]byte, error) {
	return nil, nil
}

func (cns *CnsManager) cnsRegister(name, version, address string) ([]byte, error) {

	// TODO paramCheck: name, version, address
	// common.IsHexAddress(address)

	contractAddr := common.HexToAddress(address)

	// check the owner
	if !cns.isOwner(contractAddr) {
		return nil, errors.New("not Owner?")
	}

	return cns.doCnsRegister(name, version, address)
}

func (cns *CnsManager) doCnsRegister(name, version, address string) ([]byte, error) {
	key := getSearchKey(name, version)

	// do cns Register
	value := cns.cMap.find(key)
	if value != nil {
		return nil, errors.New("[CNS] Name and version is already registered and activated in CNS")
	}

	ori := cns.callerAddr.String()

	// check name unique
	if cns.cMap.isNameDuplicated(name, ori) {
		return nil, errors.New("[CNS] Name is already registered")
	}

	// check version
	latestVersion := cns.cMap.getLatestVersion(name)
	/// latestVersion := cMap.getLatestVer(name)
	if verCompare(version, latestVersion) != 1 {
		return nil, errors.New("[CNS] Version must be larger than current latest version")
	}

	cnsInfo := newContractInfo(name, version, address, ori)
	cBytes, _ := cnsInfo.encode()
	cns.cMap.insert(key, cBytes)

	/// cMap.updateLatestVer(name, version)

	return nil, nil		// todo
}

func verCompare(ver1, ver2 string) int {
	ver1Arr := strings.Split(ver1, ".")
	ver2Arr := strings.Split(ver2, ".")

	if len(ver1Arr) != len(ver2Arr) { // Todo panic?
	}

	for i := 0; i < len(ver1Arr); i++ {
		if ver1Arr[i] > ver2Arr[i] {
			return 1
		} else if ver1Arr[i] > ver2Arr[i] {
			return -1
		} else {
			continue
		}
	}

	return 0
}

func (cns *CnsManager) cnsUnregister(name, version string) ([]byte, error) {

	//cMap := newCnsMap()

	// paramCheck: name, version

	if strings.EqualFold(version, "latest") {
		version = cns.cMap.getLatestVersion(name)
		/// version = cMap.getLatestVer(name)
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return nil, errors.New("[CNS] Name and version didn't register before")
	}

	// isOwner
	contractAddr := common.HexToAddress(cnsInfo.Address)
	if !cns.isOwner(contractAddr) {
		return nil, errors.New("[CNS] Not owner of registered contract")
	}

	cnsInfo.Enabled = false
	cBytes, _ := cnsInfo.encode()
	cns.cMap.update(key, cBytes)

	return common.Int64ToBytes(SUCCESS), nil // TODO
}

func (cns *CnsManager) getContractAddress(name, version string) ([]byte, error) {
	//cMap := newCnsMap()

	// paramCheck: name, version

	if strings.EqualFold(version, "latest") {
		version = cns.cMap.getLatestVersion(name)
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return nil, errors.New("is not existed")
	}

	// if valid
	if !cnsInfo.Enabled {
		return nil, errors.New("not valid")
	}

	return []byte(cnsInfo.Address), nil
}

func (cns *CnsManager) ifRegisteredByName(name string) ([]byte, error) {
	//cMap := newCnsMap()

	// TODO paramCheck name

	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name && cnsInfo.Enabled{
			return common.Int64ToBytes(REGISTERED), nil
		}
	}

	return common.Int64ToBytes(UNREGISTERED), nil
}

func (cns *CnsManager) ifRegisteredByAddress(address string) ([]byte, error) {
	//cMap := newCnsMap()

	// TODO paramCheck name

	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == address && cnsInfo.Enabled{
			return common.Int64ToBytes(REGISTERED), nil
		}
	}

	return common.Int64ToBytes(UNREGISTERED), nil
}

func getSearchKey(name, version string) []byte {
	return []byte(name + ":" + version)
}

func (cns *CnsManager) getRegisteredContracts() ([]byte, error) {
	//cMap := newCnsMap()
	cnsInfoArray := make([]*ContractInfo, 0)

	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Enabled {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return serializeCnsInfo(CODE_OK, MSG_OK, cnsInfoArray)
}

func serializeCnsInfo(code int, msg string, array []*ContractInfo) ([]byte, error) {
	data := newReturnMsg(code, msg, array)
	return json.Marshal(data)
}