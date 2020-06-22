package vm

import (
	"encoding/json"
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"regexp"
	"strings"
	"time"
)

const (
	MSG_OK = "ok"
	MSG_ERR = "input is invalid"

	CODE_OK = 0
	CODE_ERR = 1
)

type CnsManager struct {
	callerAddr 	*common.Address	// callerAddr = Contract.CallerAddress
	cMap	 	*cnsMap			// cMap = NewCnsMap(StateDB, Contract.CodeAddr)
	isInit		int				// isInit = evm.InitEntryID
	origin		*common.Address	// origin = evm.Context.Origin // todo necessary?
}

type ContractInfo struct {
	Name		string
	Version 	string
	Address 	string
	Origin  	string
	TimeStamp 	int64
	// Enabled 	bool
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
		TimeStamp:  time.Now().Unix(),
		// Enabled:	true,
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

// decode
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
		"cnsRegisterFromInit": 		u.cnsRegisterFromInit,
		"cnsRegister": 				u.cnsRegister,
		// "cnsUnregister":			u.cnsUnregister,
		"cnsRecall": 				u.cnsRecall,
		"getContractAddress":		u.getContractAddress,
		"ifRegisteredByAddress":	u.ifRegisteredByAddress,
		"ifRegisteredByName":		u.ifRegisteredByName,
		"getRegisteredContracts":			u.getRegisteredContractsByRange,
		"getRegisteredContractsByName":		u.getRegisteredContractsByName,			// getHistoryContractsByName -> getRegisteredContractsByName
		"getRegisteredContractsByAddress":	u.getRegisteredContractsByAddress,
		"getRegisteredContractsByOrigin":	u.getRegisteredContractsByOrigin,		// getContractInfoByAddress -> getRegisteredContractsByOrigin
	}
}

func (cns *CnsManager) isOwner(contractAddr common.Address) bool {
	callerAddr := cns.callerAddr
	contractOwnerAddr := cns.cMap.GetContractCreator(contractAddr)

	if callerAddr.Hex() == contractOwnerAddr.Hex() {
		return true
	} else {
		return false
	}
}

func (cns *CnsManager) isFromInit() bool {
	if cns.isInit != -1 {
		return true
	} else {
		return false
	}
}

func (cns *CnsManager) cnsRegisterFromInit(name, version string) (bool, error) {
	if !cns.isFromInit() {
		return false, errors.New("[CNS] cnsRegisterFromInit can only be called from init()")
	}

	address := cns.callerAddr.Hex()

	return cns.doCnsRegister(name, version, address)
}

func (cns *CnsManager) cnsRegister(name, version, address string) (bool, error) {

	// todo is it necessary?
	if cns.isFromInit() {
		return false, errors.New("[CNS] cnsRegister can't be called from init()")
	}

	contractAddr := common.HexToAddress(address)

	// check the owner
	if !cns.isOwner(contractAddr) {
		return false, errors.New("[CNS] not owner of registered contract")
	}

	return cns.doCnsRegister(name, version, address)
}

// todo precompiled regMatch
// regMatch check if string matches the pattern by regular expression
func regMatch(param, pattern string) bool {
	result, _ := regexp.MatchString(pattern, param)
	return result
}

const (
	NAME_PATTERN = `^[a-zA-Z]\w{2,15}$`		// alice
	VERSION_PATTERN = `^([\d]\.){3}[\d]$`	// 0.0.0.1
)

func (cns *CnsManager) doCnsRegister(name, version, address string) (bool, error) {
	if !common.IsHexAddress(address) {
		return false, errors.New("[CNS] address format is invalid")
	}

	if !regMatch(name, NAME_PATTERN) {
		return false, errors.New("[CNS] name format is invalid")
	}

	if regMatch(version, VERSION_PATTERN) {
		return false, errors.New("[CNS] version format is invalid")
	}

	key := getSearchKey(name, version)

	value := cns.cMap.find(key)
	if value != nil {
		return false, errors.New("[CNS] name and version is already registered and activated in CNS")
	}

	ori := cns.origin.Hex()
	//ori := cns.callerAddr.String()

	// check name unique
	if cns.cMap.isNameRegByOthers(name, ori) {
		return false, errors.New("[CNS] Name is already registered")
	}

	// check version
	latestVersion := cns.cMap.getLargestVersion(name)
	if verCompare(version, latestVersion) != 1 {
		return false, errors.New("[CNS] Version must be larger than current latest version")
	}

	cnsInfo := newContractInfo(name, version, address, ori)
	cBytes, _ := cnsInfo.encode()
	cns.cMap.insert(key, cBytes)

	cns.cMap.updateLatestVer(name, version)

	return true, nil
}

func (cns *CnsManager) isVersionDuplicated(name, ver string) bool {
	return true
}

func verCompare(ver1, ver2 string) int {
	ver1Arr := strings.Split(ver1, ".")
	ver2Arr := strings.Split(ver2, ".")

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

// cnsUnregister is deprecated, cnsUnregister -> cnsRecall
func (cns *CnsManager) cnsRecall(name, version string) (bool, error) {

	if !regMatch(name, NAME_PATTERN) {
		return false, errors.New("[CNS] name format is invalid")
	}

	if regMatch(version, VERSION_PATTERN) {
		return false, errors.New("[CNS] version format is invalid")
	}

	// if strings.EqualFold(version, "latest") {
	//	version = cns.cMap.getLatestVersion(name)
		/// version = cMap.getLatestVer(name)
	// }

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return false, errors.New("[CNS] Name and version didn't register before")
	}

	// isOwner
	contractAddr := common.HexToAddress(cnsInfo.Address)
	if !cns.isOwner(contractAddr) {
		return false, errors.New("[CNS] Not owner of registered contract")
	}

	// cBytes, _ := cnsInfo.encode()
	cns.cMap.updateLatestVer(name, version)

	return true, nil
}

func (cns *CnsManager) getContractAddress(name, version string) (string, error) {
	if !regMatch(name, NAME_PATTERN) {
		return "", errors.New("[CNS] name format is invalid")
	}

	if regMatch(version, VERSION_PATTERN) {
		return "", errors.New("[CNS] version format is invalid")
	}

	if strings.EqualFold(version, "latest") {
		version = cns.cMap.getLatestVer(name)
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return "", errors.New("[CNS] name and version is not registered in CNS")
	}

	// (deprecated)if valid
	// if !cnsInfo.Enabled {
	//	return "", errors.New("not valid")
	// }

	return cnsInfo.Address, nil
}

func (cns *CnsManager) ifRegisteredByName(name string) (bool, error) {

	if !regMatch(name, NAME_PATTERN) {
		return false, errors.New("[CNS] name format is invalid")
	}
	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name{
			return true, nil
		}
	}

	return false, nil
}

func (cns *CnsManager) ifRegisteredByAddress(address string) (bool, error) {

	if !common.IsHexAddress(address) {
		return false, errors.New("[CNS] contract address format is invalid")
	}
	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == address{
			return true, nil
		}
	}

	return false, nil
}

func getSearchKey(name, version string) []byte {
	return []byte(name + ":" + version)
}

func (cns *CnsManager) getRegisteredContractsByRange(head, size int) ([]byte, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var tail int

	invalidRange := head >= cns.cMap.total() || size < 0

	if invalidRange {
		return nil, errors.New("")
	}

	if size == 0 || head + size > cns.cMap.total() {
		tail = cns.cMap.total()
	} else {
		tail = head + size
	}

	for index := head; index < tail; index++{
		cnsInfo := cns.cMap.get(index)
		cnsInfoArray = append(cnsInfoArray, cnsInfo)
	}

	return serializeCnsInfo(CODE_OK, MSG_OK, cnsInfoArray)
}

// before: getHistoryContractsByName -> after refactory: getRegisteredContractsByName
func (cns *CnsManager) getRegisteredContractsByName(name string) ([]byte, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)

	if !regMatch(name, NAME_PATTERN) {
		return nil, errors.New("[CNS] name format is invalid")
	}

	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return serializeCnsInfo(CODE_OK, MSG_OK, cnsInfoArray)
}

func (cns *CnsManager) getRegisteredContractsByAddress(addr string) ([]byte, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)

	if !common.IsHexAddress(addr) {
		return nil, errors.New("[CNS] contract address format is invalid")
	}

	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == addr {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return serializeCnsInfo(CODE_OK, MSG_OK, cnsInfoArray)
}

// todo: return serializeCnsInfo(CODE_ERR, MSG_ERR, cnsInfoArray) ???
// before: getContractInfoByAddress -> after refactory: getRegisteredContractsByOrigin
func (cns *CnsManager) getRegisteredContractsByOrigin(origin string) ([]byte, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)

	if !common.IsHexAddress(origin) {
		return nil, errors.New("[CNS] contract owner address format is invalid")
	}

	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Origin == origin {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return serializeCnsInfo(CODE_OK, MSG_OK, cnsInfoArray)
}

func serializeCnsInfo(code int, msg string, array []*ContractInfo) ([]byte, error) {
	data := newReturnMsg(code, msg, array)
	return json.Marshal(data)
}