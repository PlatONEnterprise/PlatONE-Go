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

	SUCCESS = 0
	FAILURE = 1

	UNDEFINED = -1
	UNREGISTERD = 0
	REGISTERD = 1

	NAME_PATTERN = `^[a-zA-Z]\w{2,15}$`		// alice
	VERSION_PATTERN = `^([\d]\.){3}[\d]$`	// 0.0.0.1

	ERR_NAME_INVALID = "[CNS] name format is invalid"
	ERR_VERSION_INVALID = "[CNS] version format is invalid"
	ERR_ADDRESS_INVALID = "[CNS] address format is invalid"
	ERR_NOT_OWNER = "[CNS] not owner of registered contract"
)

var (
	regName = regexp.MustCompile(NAME_PATTERN)
	regVer = regexp.MustCompile(VERSION_PATTERN)
)

type CnsManager struct {
	callerAddr 	common.Address	// callerAddr = Contract.CallerAddress
	cMap	 	*cnsMap			// cMap = NewCnsMap(StateDB, Contract.CodeAddr)
	isInit		int				// isInit = evm.InitEntryID
	origin		common.Address	// origin = evm.Context.Origin // todo necessary?
}

type ContractInfo struct {
	Name		string
	Version 	string
	Address 	string
	Origin  	string
	TimeStamp 	int64
	// Enabled 	bool	// deprecated
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
	callerAddr := cns.origin	// todo
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

func (cns *CnsManager) cnsRegisterFromInit(name, version string) (int, error) {
	if !cns.isFromInit() {
		return FAILURE, errors.New("[CNS] cnsRegisterFromInit can only be called from init()")
	}

	address := cns.callerAddr.Hex()

	return cns.doCnsRegister(name, version, address)
}

func (cns *CnsManager) cnsRegister(name, version, address string) (int, error) {

	if cns.isFromInit() {
		return FAILURE, errors.New("[CNS] cnsRegister can't be called from init()")
	}

	contractAddr := common.HexToAddress(address)

	// check the owner
	if !cns.isOwner(contractAddr) {
		return FAILURE, errors.New(ERR_NOT_OWNER)
	}

	return cns.doCnsRegister(name, version, address)
}

func (cns *CnsManager) doCnsRegister(name, version, address string) (int, error) {
	if !common.IsHexAddress(address) {
		return FAILURE, errors.New(ERR_ADDRESS_INVALID)
	}

	if !regName.MatchString(name) {
		return FAILURE, errors.New(ERR_NAME_INVALID)
	}

	if !regVer.MatchString(version) {
		return FAILURE, errors.New(ERR_VERSION_INVALID)
	}

	key := getSearchKey(name, version)

	value := cns.cMap.find(key)
	if value != nil {
		return FAILURE, errors.New("[CNS] name and version is already registered and activated in CNS")
	}

	ori := cns.origin.Hex()
	//ori := cns.callerAddr.String()

	// check name unique
	if cns.cMap.isNameRegByOthers(name, ori) {
		return FAILURE, errors.New("[CNS] Name is already registered")
	}

	// check version
	largestVersion := cns.cMap.getLargestVersion(name)
	if verCompare(version, largestVersion) != 1 {
		return FAILURE, errors.New("[CNS] Version must be larger than previous version")
	}

	cnsInfo := newContractInfo(name, version, address, ori)
	cBytes, _ := cnsInfo.encode()
	cns.cMap.insert(key, cBytes)

	cns.cMap.updateLatestVer(name, version)

	return SUCCESS, nil
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
func (cns *CnsManager) cnsRecall(name, version string) (int, error) {

	if !regName.MatchString(name) {
		return FAILURE, errors.New(ERR_NAME_INVALID)
	}

	if !regVer.MatchString(version) {
		return FAILURE, errors.New(ERR_VERSION_INVALID)
	}

	// if strings.EqualFold(version, "latest") {
	//	version = cns.cMap.getLatestVersion(name)
		/// version = cMap.getLatestVer(name)
	// }

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return FAILURE, errors.New("[CNS] Name and version didn't register before")
	}

	// isOwner
	contractAddr := common.HexToAddress(cnsInfo.Address)
	if !cns.isOwner(contractAddr) {
		return FAILURE, errors.New(ERR_NOT_OWNER)
	}

	// cBytes, _ := cnsInfo.encode()
	cns.cMap.updateLatestVer(name, version)

	return SUCCESS, nil
}

func (cns *CnsManager) getContractAddress(name, version string) (string, error) {
	if !regName.MatchString(name) {
		return "", errors.New(ERR_NAME_INVALID)
	}

	if !regVer.MatchString(version) {
		return "", errors.New(ERR_VERSION_INVALID)
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

func (cns *CnsManager) ifRegisteredByName(name string) (int, error) {

	if !regName.MatchString(name) {
		return UNDEFINED, errors.New(ERR_NAME_INVALID)
	}
	
	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name{
			return REGISTERD, nil
		}
	}

	return UNREGISTERD, nil
}

func (cns *CnsManager) ifRegisteredByAddress(address string) (int, error) {

	if !common.IsHexAddress(address) {
		return UNDEFINED, errors.New("[CNS] contract address format is invalid")
	}
	
	for index := 0; index < cns.cMap.total(); index++{
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == address{
			return REGISTERD, nil
		}
	}

	return UNREGISTERD, nil
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

	if !regName.MatchString(name) {
		return nil, errors.New(ERR_NAME_INVALID)
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
		return nil, errors.New(ERR_ADDRESS_INVALID)
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
		return nil, errors.New(ERR_ADDRESS_INVALID)
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