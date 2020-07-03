package vm

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	msgOk = "ok"
	// MSG_ERR = "input is invalid"

	codeOk = 0
	// CODE_ERR = 1

	success = 0
	failure = 1

	undefined    = -1
	unregistered = 0
	registered   = 1

	namePattern    = `^[a-zA-Z]\w{2,15}$` // alice
	versionPattern = `^([\d]\.){3}[\d]$`  // 0.0.0.1
)

var (
	errNameInvalid    = errors.New("[CNS] name format is invalid")
	errVersionInvalid = errors.New("[CNS] version format is invalid")
	errAddressInvalid = errors.New("[CNS] address format is invalid")
	errNotOwner       = errors.New("[CNS] not owner of registered contract")
)

var (
	regName = regexp.MustCompile(namePattern)
	regVer  = regexp.MustCompile(versionPattern)
)

var (
	cnsSysContractsMap = map[string]common.Address{
		"__sys_ParamManager": syscontracts.ParameterManagementAddress,
		"__sys_NodeManager":  syscontracts.NodeManagementAddress,
		"__sys_UserManager":  syscontracts.UserManagementAddress,
	}
)

// CnsManager
type CnsManager struct {
	callerAddr common.Address // callerAddr = Contract.CallerAddress
	cMap       *cnsMap        // cMap 	  = NewCnsMap(StateDB, Contract.CodeAddr)
	isInit     int            // isInit 	  = evm.InitEntryID
	origin     common.Address // origin 	  = evm.Context.Origin // todo necessary?
}

// ContractInfo stores cns registration info of a contract address
type ContractInfo struct {
	Name      string
	Version   string
	Address   string
	Origin    string
	TimeStamp uint64
	// Enabled 	bool			// deprecated
}

type returnMsg struct {
	Code  int
	Msg   string
	Array []*ContractInfo
}

func newContractInfo(name, version, address, origin string) *ContractInfo {
	return &ContractInfo{
		Name:      name,
		Version:   version,
		Address:   address,
		Origin:    origin,
		TimeStamp: uint64(time.Now().Unix()),
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

// decodeCnsInfo decodes rlp bytes to ContractInfo struct
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
		"cnsRegisterFromInit":             u.cnsRegisterFromInit,
		"cnsRegister":                     u.cnsRegister,
		"cnsRedirect":                     u.cnsRedirect, // cnsUnregister is deprecated, replaced by cnsRedirect
		"getContractAddress":              u.getContractAddress,
		"ifRegisteredByAddress":           u.ifRegisteredByAddress,
		"ifRegisteredByName":              u.ifRegisteredByName,
		"getRegisteredContracts":          u.getRegisteredContractsByRange,
		"getRegisteredContractsByName":    u.getRegisteredContractsByName, // getHistoryContractsByName -> getRegisteredContractsByName
		"getRegisteredContractsByAddress": u.getRegisteredContractsByAddress,
		"getRegisteredContractsByOrigin":  u.getRegisteredContractsByOrigin, // getContractInfoByAddress -> getRegisteredContractsByOrigin
	}
}

// isOwner checks if the caller is the owner of the contract to be registerd
func (cns *CnsManager) isOwner(contractAddr common.Address) bool {
	callerAddr := cns.origin
	contractOwnerAddr := cns.cMap.GetContractCreator(contractAddr)

	if callerAddr.Hex() == contractOwnerAddr.Hex() {
		return true
	} else {
		return false
	}
}

// isFromInit checks if the method is called from init()
func (cns *CnsManager) isFromInit() bool {
	if cns.isInit != -1 {
		return true
	} else {
		return false
	}
}

func (cns *CnsManager) cnsRegisterFromInit(name, version string) (int, error) {
	if !cns.isFromInit() {
		return failure, errors.New("[CNS] cnsRegisterFromInit can only be called from init()")
	}

	address := cns.callerAddr.Hex()
	return cns.doCnsRegister(name, version, address)
}

func (cns *CnsManager) cnsRegister(name, version, address string) (int, error) {

	if cns.isFromInit() {
		return failure, errors.New("[CNS] cnsRegister can't be called from init()")
	}

	contractAddr := common.HexToAddress(address)

	// check the owner
	if !cns.isOwner(contractAddr) {
		return failure, errNotOwner
	}

	return cns.doCnsRegister(name, version, address)
}

func (cns *CnsManager) doCnsRegister(name, version, address string) (int, error) {
	if !common.IsHexAddress(address) {
		return failure, errAddressInvalid
	}

	if !regName.MatchString(name) {
		return failure, errNameInvalid
	}

	if !regVer.MatchString(version) {
		return failure, errVersionInvalid
	}

	// check if registered
	key := getSearchKey(name, version)
	value := cns.cMap.find(key)
	if value != nil {
		return failure, errors.New("[CNS] name and version is already registered and activated in CNS")
	}

	// check is name unique
	ori := cns.origin.Hex()
	if cns.cMap.isNameRegByOthers(name, ori) {
		return failure, errors.New("[CNS] Name is already registered")
	}

	// check is version valid
	largestVersion := cns.cMap.getLargestVersion(name)
	if verCompare(version, largestVersion) != 1 {
		return failure, errors.New("[CNS] Version must be larger than previous version")
	}

	// new a contractInfo struct and serialize it
	cnsInfo := newContractInfo(name, version, address, ori)
	cBytes, _ := cnsInfo.encode()

	// record the info to stateDB
	cns.cMap.insert(key, cBytes)

	// update the latest version of the cns name
	cns.cMap.updateLatestVer(name, version)

	return success, nil
}

// verCompare compares the versions
// 1: ver1 > ver2
// -1: ver1 < ver2
// 0: ver1 = ver2
func verCompare(ver1, ver2 string) int {
	ver1Arr := strings.Split(ver1, ".")
	ver2Arr := strings.Split(ver2, ".")

	for i := 0; i < len(ver1Arr); i++ {
		if ver1Arr[i] > ver2Arr[i] {
			return 1
		} else if ver1Arr[i] < ver2Arr[i] {
			return -1
		} else {
			continue
		}
	}

	return 0
}

// cnsUnregister is deprecated, cnsUnregister -> cnsRedirect
// cnsRedirect selects a specific version of a cns name and set it "latest"
func (cns *CnsManager) cnsRedirect(name, version string) (int, error) {

	if !regName.MatchString(name) {
		return failure, errNameInvalid
	}

	if !regVer.MatchString(version) {
		return failure, errVersionInvalid
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return failure, errors.New("[CNS] Name and version didn't register before")
	}

	// check is Owner
	contractAddr := common.HexToAddress(cnsInfo.Address)
	if !cns.isOwner(contractAddr) {
		return failure, errNotOwner
	}

	cns.cMap.updateLatestVer(name, version)

	return success, nil
}

// getContractAddress returns the address of a cns name at specific version
func (cns *CnsManager) getContractAddress(name, version string) (string, error) {
	if sysCon, ok := cnsSysContractsMap[name]; ok {
		return sysCon.String(), nil
	}
	if strings.EqualFold(version, "latest") {
		version = cns.cMap.getLatestVer(name)
	}

	if !regVer.MatchString(version) {
		return "", errVersionInvalid
	}

	if !regName.MatchString(name) {
		return "", errNameInvalid
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return "", errors.New("[CNS] name and version is not registered in CNS")
	}

	return cnsInfo.Address, nil
}

func (cns *CnsManager) ifRegisteredByName(name string) (int, error) {

	if !regName.MatchString(name) {
		return undefined, errNameInvalid
	}

	for index := 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name {
			return registered, nil
		}
	}

	return unregistered, nil
}

func (cns *CnsManager) ifRegisteredByAddress(address string) (int, error) {

	if !common.IsHexAddress(address) {
		return undefined, errAddressInvalid
	}

	for index := 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == address {
			return registered, nil
		}
	}

	return unregistered, nil
}

func getSearchKey(name, version string) []byte {
	return []byte(name + ":" + version)
}

// getRegisteredContractsByRange returns the cnsContractInfo within the ranges specified
// the size 0 means returning the cnsContractInfo from head to the end
func (cns *CnsManager) getRegisteredContractsByRange(head, size int) (string, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var tail int

	// check the head and size are valid numbers
	invalidRange := head >= cns.cMap.total() || size < 0
	if invalidRange {
		return "", errors.New("")
	}

	// make sure the head + size does not exceed the total numbers of cnsContractInfo
	if size == 0 || head+size > cns.cMap.total() {
		tail = cns.cMap.total()
	} else {
		tail = head + size
	}

	for index := head; index < tail; index++ {
		cnsInfo := cns.cMap.get(index)
		cnsInfoArray = append(cnsInfoArray, cnsInfo)
	}

	return serializeCnsInfo(codeOk, msgOk, cnsInfoArray)
}

// before: getHistoryContractsByName -> after refactory: getRegisteredContractsByName
func (cns *CnsManager) getRegisteredContractsByName(name string) (string, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)

	if !regName.MatchString(name) {
		return "", errNameInvalid
	}

	for index := 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return serializeCnsInfo(codeOk, msgOk, cnsInfoArray)
}

func (cns *CnsManager) getRegisteredContractsByAddress(addr string) (string, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)

	if !common.IsHexAddress(addr) {
		return "", errAddressInvalid
	}

	for index := 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == addr {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return serializeCnsInfo(codeOk, msgOk, cnsInfoArray)
}

// before: getContractInfoByAddress -> after refactory: getRegisteredContractsByOrigin
func (cns *CnsManager) getRegisteredContractsByOrigin(origin string) (string, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)

	if !common.IsHexAddress(origin) {
		return "", errAddressInvalid
	}

	for index := 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Origin == origin {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return serializeCnsInfo(codeOk, msgOk, cnsInfoArray)
}

func serializeCnsInfo(code int, msg string, array []*ContractInfo) (string, error) {
	data := newReturnMsg(code, msg, array)
	cBytes, err := json.Marshal(data)
	return string(cBytes), err
}
