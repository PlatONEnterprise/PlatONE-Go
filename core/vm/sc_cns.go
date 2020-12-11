package vm

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/rlp"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
)

const (
	cnsSuccess         CodeType = 0
	cnsInvalidCall     CodeType = 1
	cnsNoPermission    CodeType = 2
	cnsInvalidArgument CodeType = 3
	cnsRegErr          CodeType = 4
)

const (
	cnsUnregistered CodeType = 0
	cnsRegistered   CodeType = 1
)

const (
	cnsMigSuccess CodeType = 0
	cnsMigFailed  CodeType = 1
)

const (
	/// namePattern    = `^[a-zA-Z]\w{2,15}$`  // alice
	versionRegPattern = `^([\d]{1,3}\.){3}[\d]{1,3}$` // 0.0.0.1
)

var (
	/// regName = regexp.MustCompile(namePattern)
	regVer = regexp.MustCompile(versionRegPattern)
)
//
//var (
//	cnsSysContractsMap = map[string]common.Address{
//		"__sys_ParamManager": syscontracts.ParameterManagementAddress,
//		"__sys_NodeManager":  syscontracts.NodeManagementAddress,
//		"__sys_UserManager":  syscontracts.UserManagementAddress,
//		"cnsManager":         syscontracts.CnsManagementAddress,
//	}
//)

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
	Name      string         `json:"name"`
	Version   string         `json:"version"`
	Address   common.Address `json:"address"`
	Origin    common.Address `json:"origin"`
	TimeStamp uint64         `json:"create_time"`
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

// decodeCnsInfo decodes rlp bytes to ContractInfo struct
func decodeCnsInfo(data []byte) (*ContractInfo, error) {
	var ci ContractInfo
	if err := rlp.DecodeBytes(data, &ci); nil != err {
		return nil, err
	}

	return &ci, nil
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

func (cns *CnsManager) importOldCnsManagerData(cnsInfos []*ContractInfo) {
	// todo: check validation of  import data?

	for _, data := range cnsInfos {
		key := getSearchKey(data.Name, data.Version)
		cns.cMap.insert(key, data)

		// set the largest version to CURRENT version
		curVersion := cns.cMap.getCurrentVer(data.Name)
		if verCompare(data.Version, curVersion) == 1 {
			cns.cMap.setCurrentVer(data.Name, data.Version)
		}
	}

	cns.emitNotifyEvent(cnsSuccess, "[CNS] cns data migration succeed")
}

func (cns *CnsManager) cnsRegisterFromInit(name, version string) error {
	if !cns.isFromInit() {
		cns.emitNotifyEvent(cnsInvalidCall, errInvalidCallNotFromInit.Error())
		return errInvalidCallNotFromInit
	}

	address := cns.caller
	return cns.doCnsRegister(name, version, address)
}

func (cns *CnsManager) cnsRegister(name, version string, contractAddr common.Address) error {

	if cns.isFromInit() {
		cns.emitNotifyEvent(cnsInvalidCall, errInvalidCallFromInit.Error())
		return errInvalidCallFromInit
	}

	// check the owner
	if !cns.isOwner(contractAddr) {
		cns.emitNotifyEvent(cnsNoPermission, errNotOwner.Error())
		return errNotOwner
	}

	return cns.doCnsRegister(name, version, contractAddr)
}

func (cns *CnsManager) doCnsRegister(name, version string, address common.Address) error {

	if ok, _ := checkNameFormat(name); !ok {
		cns.emitNotifyEvent(cnsInvalidArgument, errNameInvalid.Error())
		return errNameInvalid
	}

	if !regVer.MatchString(version) {
		cns.emitNotifyEvent(cnsInvalidArgument, errVersionInvalid.Error())
		return errVersionInvalid
	}

	// check if registered
	key := getSearchKey(name, version)
	value := cns.cMap.find(key)
	if value != nil {
		cns.emitNotifyEvent(cnsRegErr, errNameAndVerReg.Error())
		return errNameAndVerReg
	}

	// check is name unique
	ori := cns.origin
	if cns.cMap.isNameRegByOthers(name, ori) {
		cns.emitNotifyEvent(cnsRegErr, errNameReg.Error())
		return errNameReg
	}

	// check is version valid
	largestVersion := cns.cMap.getLargestVersion(name)
	if verCompare(version, largestVersion) != 1 {
		cns.emitNotifyEvent(cnsInvalidArgument, errLowRegVersion.Error())
		return errLowRegVersion
	}

	// new a contractInfo struct
	cnsInfo := newContractInfo(name, version, address, ori)

	// record the info to stateDB
	cns.cMap.insert(key, cnsInfo)

	// update the current version of the cns name
	cns.cMap.setCurrentVer(name, version)

	cns.emitNotifyEvent(cnsSuccess, "[CNS] cns register succeed")
	return nil
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
func (cns *CnsManager) cnsRedirect(name, version string) error {

	if ok, _ := checkNameFormat(name); !ok {
		cns.emitNotifyEvent(cnsInvalidArgument, errNameInvalid.Error())
		return errNameInvalid
	}

	if !regVer.MatchString(version) {
		cns.emitNotifyEvent(cnsInvalidArgument, errVersionInvalid.Error())
		return errVersionInvalid
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		cns.emitNotifyEvent(cnsRegErr, errNameAndVerUnReg.Error())
		return errNameAndVerUnReg
	}

	// check is Owner
	contractAddr := cnsInfo.Address
	if !cns.isOwner(contractAddr) {
		cns.emitNotifyEvent(cnsNoPermission, errNotOwner.Error())
		return errNotOwner
	}

	cns.cMap.setCurrentVer(name, version)

	cns.emitNotifyEvent(cnsSuccess, "[CNS] cns redirect succeed")
	return nil
}

// getContractAddress returns the address of a cns name at specific version
func (cns *CnsManager) getContractAddress(name, version string) (common.Address, error) {
	//if sysCon, ok := cnsSysContractsMap[name]; ok {
	//	return sysCon, nil
	//}

	if strings.EqualFold(version, "latest") {
		version = cns.cMap.getCurrentVer(name)
	}

	if ok, _ := checkNameFormat(name); !ok {
		return common.Address{}, errNameInvalid
	}

	if !regVer.MatchString(version) {
		return common.Address{}, errVersionInvalid
	}

	key := getSearchKey(name, version)

	// get cnsInfo
	cnsInfo := cns.cMap.find(key)
	if cnsInfo == nil {
		return common.Address{}, errors.New("[CNS] name and version is not registered in CNS")
	}

	return cnsInfo.Address, nil
}

func (cns *CnsManager) ifRegisteredByName(name string) (bool, error) {
	var index uint64

	if ok, _ := checkNameFormat(name); !ok {
		return false, errNameInvalid
	}

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (cns *CnsManager) ifRegisteredByAddress(address common.Address) (bool, error) {
	var index uint64

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == address {
			return true, nil
		}
	}

	return false, nil
}

// getRegisteredContractsByRange returns the cnsContractInfo within the ranges specified
// the size 0 means returning the cnsContractInfo from head to the end
func (cns *CnsManager) getRegisteredContractsByRange(head, size int) ([]*ContractInfo, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var tail int

	total := int(cns.cMap.total())

	// check the head and size are valid numbers
	invalidRange := head < 0 || size < 0
	if invalidRange {
		return nil, errors.New("[CNS] invalid range")
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

	return cnsInfoArray, nil
}

// before: getHistoryContractsByName -> after refactory: getRegisteredContractsByName
func (cns *CnsManager) getRegisteredContractsByName(name string) ([]*ContractInfo, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var index uint64

	if ok, _ := checkNameFormat(name); !ok {
		return nil, errNameInvalid
	}

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Name == name {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return cnsInfoArray, nil
}

func (cns *CnsManager) getRegisteredContractsByAddress(addr common.Address) ([]*ContractInfo, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var index uint64

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Address == addr {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return cnsInfoArray, nil
}

// before: getContractInfoByAddress -> after refactory: getRegisteredContractsByOrigin
func (cns *CnsManager) getRegisteredContractsByOrigin(origin common.Address) ([]*ContractInfo, error) {
	var cnsInfoArray = make([]*ContractInfo, 0)
	var index uint64

	for index = 0; index < cns.cMap.total(); index++ {
		cnsInfo := cns.cMap.get(index)
		if cnsInfo.Origin == origin {
			cnsInfoArray = append(cnsInfoArray, cnsInfo)
		}
	}

	return cnsInfoArray, nil
}

func getCnsAddress(stateDB StateDB, name, version string) (common.Address, error) {
	cns := newCnsManager(stateDB)
	return cns.getContractAddress(name, version)
}

func getRegisterStatusByName(stateDB StateDB, name string) (bool, error) {
	cns := newCnsManager(stateDB)
	return cns.ifRegisteredByName(name)
}

func (cns *CnsManager) emitNotifyEvent(code CodeType, msg string) {
	topic := "[CNS] Notify"
	cns.emitEvent(topic, code, msg)
}

func (cns *CnsManager) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(cns.cMap.contractAddr, cns.cMap, cns.blockNumber.Uint64(), topic, code, msg)
}
