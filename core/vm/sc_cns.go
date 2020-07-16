package vm

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	cnsSuccess         CodeType = 0
	cnsInvalidCall     CodeType = 1
	cnsNoPermission    CodeType = 2
	cnsInvalidArgument CodeType = 3
	cnsRegDuplicate    CodeType = 4
	cnsRegNotExist     CodeType = 5
)

const (
	success int32 = 0
	failure int32 = 1

	undefined    int32 = -1
	unregistered int32 = 0
	registered   int32 = 1
)

const (
	namePattern    = `^[a-zA-Z]\w{2,15}$` // alice
	versionPattern = `^([\d]\.){3}[\d]$`  // 0.0.0.1
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
		"cnsManager":         syscontracts.CnsManagementAddress,
	}
)

// CnsManager
type CnsManager struct {
	caller      common.Address // caller = Contract.CallerAddress
	cMap        *cnsMap        // cMap 	= NewCnsMap(StateDB, Contract.CodeAddr)
	isInit      int            // isInit = evm.InitEntryID
	origin      common.Address // origin = evm.Context.Origin
	blockNumber *big.Int       // blockNumber = evm.BlockNumber
}

// ContractInfo stores cns registration info of a contract address
type ContractInfo struct {
	Name      string
	Version   string
	Address   common.Address
	Origin    common.Address
	TimeStamp uint64
}

func newContractInfo(name, version string, address, origin common.Address) *ContractInfo {
	return &ContractInfo{
		Name:      name,
		Version:   version,
		Address:   address,
		Origin:    origin,
		TimeStamp: uint64(time.Now().Unix()),
	}
}

func newCnsManager(stateDB StateDB) *CnsManager {
	return &CnsManager{
		cMap: NewCnsMap(stateDB, syscontracts.CnsManagementAddress),
	}
}

func (ci *ContractInfo) encode() ([]byte, error) {
	return rlp.EncodeToBytes(ci)
}

// todo: deprecated
// decodeCnsInfo decodes rlp bytes to ContractInfo struct
func decodeCnsInfo(data []byte) (*ContractInfo, error) {
	var ci ContractInfo
	if err := rlp.DecodeBytes(data, &ci); nil != err {
		return nil, err
	}

	return &ci, nil
}

func (cns *CnsManager) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.CnsManagerGas
}

// Run runs the precompiled contract
func (cns *CnsManager) Run(input []byte) ([]byte, error) {
	return execSC(input, cns.AllExportFns())
}

// for access control
func (cns *CnsManager) AllExportFns() SCExportFns {
	return SCExportFns{
		"cnsRegisterFromInit":             cns.cnsRegisterFromInit,
		"cnsRegister":                     cns.cnsRegister,
		"cnsRedirect":                     cns.cnsRedirect, // cnsUnregister is deprecated, replaced by cnsRedirect
		"getContractAddress":              cns.getContractAddress,
		"ifRegisteredByAddress":           cns.ifRegisteredByAddress,
		"ifRegisteredByName":              cns.ifRegisteredByName,
		"getRegisteredContracts":          cns.getRegisteredContractsByRange,
		"getRegisteredContractsByName":    cns.getRegisteredContractsByName, // getHistoryContractsByName -> getRegisteredContractsByName
		"getRegisteredContractsByAddress": cns.getRegisteredContractsByAddress,
		"getRegisteredContractsByOrigin":  cns.getRegisteredContractsByOrigin, // getContractInfoByAddress -> getRegisteredContractsByOrigin
	}
}

// isOwner checks if the caller is the owner of the contract to be registered
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

func (cns *CnsManager) cnsRegisterFromInit(name, version string) (int32, error) {
	if !cns.isFromInit() {
		cns.emitNotifyEvent(cnsInvalidCall, "[CNS] cnsRegisterFromInit can only be called from init()")
		return failure, errors.New("[CNS] cnsRegisterFromInit can only be called from init()")
	}

	address := cns.caller
	return cns.doCnsRegister(name, version, address)
}

func (cns *CnsManager) cnsRegister(name, version string, contractAddr common.Address) (int32, error) {

	if cns.isFromInit() {
		cns.emitNotifyEvent(cnsInvalidCall, "[CNS] cnsRegister can't be called from init()")
		return failure, errors.New("[CNS] cnsRegister can't be called from init()")
	}

	// check the owner
	if !cns.isOwner(contractAddr) {
		cns.emitNotifyEvent(cnsNoPermission, "[CNS] not owner of registered contract")
		return failure, errNotOwner
	}

	return cns.doCnsRegister(name, version, contractAddr)
}

func (cns *CnsManager) doCnsRegister(name, version string, address common.Address) (int32, error) {

	if !regName.MatchString(name) {
		cns.emitNotifyEvent(cnsInvalidArgument, errNameInvalid.Error())
		return failure, errNameInvalid
	}

	if !regVer.MatchString(version) {
		cns.emitNotifyEvent(cnsInvalidArgument, errVersionInvalid.Error())
		return failure, errVersionInvalid
	}

	// check if registered
	key := getSearchKey(name, version)
	value := cns.cMap.find(key)
	if value != nil {
		cns.emitNotifyEvent(cnsRegDuplicate, "[CNS] name and version is already registered and activated in CNS")
		return failure, errors.New("[CNS] name and version is already registered and activated in CNS")
	}

	// check is name unique
	ori := cns.origin
	if cns.cMap.isNameRegByOthers(name, ori) {
		cns.emitNotifyEvent(cnsRegDuplicate, "[CNS] Name is already registered")
		return failure, errors.New("[CNS] Name is already registered")
	}

	// check is version valid
	largestVersion := cns.cMap.getLargestVersion(name)
	if verCompare(version, largestVersion) != 1 {
		cns.emitNotifyEvent(cnsInvalidArgument, "[CNS] Version must be larger than previous version")
		return failure, errors.New("[CNS] Version must be larger than previous version")
	}

	// new a contractInfo struct
	cnsInfo := newContractInfo(name, version, address, ori)

	// record the info to stateDB
	cns.cMap.insert(key, cnsInfo)

	// update the current version of the cns name
	cns.cMap.setCurrentVer(name, version)

	cns.emitNotifyEvent(cnsSuccess, "[CNS] cns register succeed")
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
// cnsRedirect selects a specific version of a cns name and set it to current version
func (cns *CnsManager) cnsRedirect(name, version string) (int32, error) {

	if !regName.MatchString(name) {
		cns.emitNotifyEvent(cnsInvalidArgument, errNameInvalid.Error())
		return failure, errNameInvalid
	}

	if !regVer.MatchString(version) {
		cns.emitNotifyEvent(cnsInvalidArgument, errVersionInvalid.Error())
		return failure, errVersionInvalid
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		cns.emitNotifyEvent(cnsRegNotExist, "[CNS] Name and version didn't register before")
		return failure, errors.New("[CNS] Name and version didn't register before")
	}

	// check is Owner
	contractAddr := cnsInfo.Address
	if !cns.isOwner(contractAddr) {
		cns.emitNotifyEvent(cnsNoPermission, errNotOwner.Error())
		return failure, errNotOwner
	}

	cns.cMap.setCurrentVer(name, version)

	cns.emitNotifyEvent(cnsSuccess, "[CNS] cns redirect succeed")
	return success, nil
}

// getContractAddress returns the address of a cns name at specific version
func (cns *CnsManager) getContractAddress(name, version string) (string, error) {
	if sysCon, ok := cnsSysContractsMap[name]; ok {
		return sysCon.String(), nil
	}

	if strings.EqualFold(version, "latest") {
		version = cns.cMap.getCurrentVer(name)
	}

	if !regName.MatchString(name) {
		return "", errNameInvalid
	}

	if !regVer.MatchString(version) {
		return "", errVersionInvalid
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return "", errors.New("[CNS] name and version is not registered in CNS")
	}

	return cnsInfo.Address.String(), nil
}

func (cns *CnsManager) ifRegisteredByName(name string) (int32, error) {
	var index uint64

	if !regName.MatchString(name) {
		return undefined, errNameInvalid
	}

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name {
			return registered, nil
		}
	}

	return unregistered, nil
}

func (cns *CnsManager) ifRegisteredByAddress(address common.Address) (int32, error) {
	var index uint64

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == address {
			return registered, nil
		}
	}

	return unregistered, nil
}

// getRegisteredContractsByRange returns the cnsContractInfo within the ranges specified
// the size 0 means returning the cnsContractInfo from head to the end
func (cns *CnsManager) getRegisteredContractsByRange(head, size int) (string, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var tail int

	total := int(cns.cMap.total())

	// check the head and size are valid numbers
	invalidRange := head >= total || size < 0
	if invalidRange {
		return newInternalErrorResult(errors.New("invalid range")).String(), nil
	}

	// make sure the head + size does not exceed the total numbers of cnsContractInfo
	if size == 0 || head+size > total {
		tail = total
	} else {
		tail = head + size
	}

	for index := head; index < tail; index++ {
		cnsInfo := cns.cMap.get(uint64(index))
		cnsInfoArray = append(cnsInfoArray, cnsInfo)
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

// before: getHistoryContractsByName -> after refactory: getRegisteredContractsByName
func (cns *CnsManager) getRegisteredContractsByName(name string) (string, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var index uint64

	if !regName.MatchString(name) {
		return newInternalErrorResult(errNameInvalid).String(), nil
	}

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

func (cns *CnsManager) getRegisteredContractsByAddress(addr common.Address) (string, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var index uint64

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == addr {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

// before: getContractInfoByAddress -> after refactory: getRegisteredContractsByOrigin
func (cns *CnsManager) getRegisteredContractsByOrigin(origin common.Address) (string, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var index uint64

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Origin == origin {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

func getCnsAddress(stateDB StateDB, name, version string) (string, error) {
	cns := newCnsManager(stateDB)
	return cns.getContractAddress(name, version)
}

func (cns *CnsManager) emitNotifyEvent(code CodeType, msg string) {
	topic := "Notify"
	emitEvent(cns.cMap.contractAddr, cns.cMap, cns.blockNumber.Uint64(), topic, code, msg)
}
