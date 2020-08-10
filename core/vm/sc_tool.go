package vm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common/bcwasmutil"
	"reflect"
	"regexp"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	nameRegPattarn = `^[a-zA-Z0-9_\p{Han}]{1,128}$`
	emailRegPattarn = `\w[-\w.+]*@([A-Za-z0-9][-A-Za-z0-9]+\.)+[A-Za-z]{2,14}`
	ipRegPattarn = `(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)`
	telePhonePattarn = "[0-9-()（）]{7,18}"
	mobilePhonePattarn = "[0-9]{3,13}"
)

var (
	errFuncNotFoundInExportFuncs = errors.New("the func not found in export function set")
	errParamsNumInvalid          = errors.New("the number of params is invalid")
)

var (
	errUnsupportedRole      = errors.New("Unsupported role ")
	errNoPermission         = errors.New("No Permission ")
	errAlreadySetSuperAdmin = errors.New("Already Set SuperAdmin ")
	errParamInvalid         = errors.New("param is invalid")
	errEncodeFailure        = errors.New("encode failure ")
	errContactNameNotExist  = errors.New("contract name not exist")

	errNameInvalid    = errors.New("[CNS] name format is invalid")
	errVersionInvalid = errors.New("[CNS] version format is invalid")
	errAddressInvalid = errors.New("[CNS] address format is invalid")
	errNotOwner       = errors.New("[CNS] not owner of registered contract")
	errEmptyValue     = errors.New("Empty value")
)

var (
	ZeroAddress = common.Address{}
)

var fwErrNotOwner = errors.New("FW : error, only contract owner can set firewall setting")

func execSC(input []byte, fns SCExportFns) (string,[]byte, error) {
	txType, fnName, fn, params, err := retrieveFnAndParams(input, fns)
	if nil != err {
		log.Error("failed to retrieve func name and params.", "error", err, "function", fnName)
		return fnName, nil, err
	}

	//execute system contract method
	//all the export method of system contracts must return two results,
	//first result type is: primitive type, second result type: error
	result := reflect.ValueOf(fn).Call(params)
	if err, ok := result[1].Interface().(error); ok {
		log.Error("execute system contract failed.", "error", err)
	}

	//vm run successfully, so return nil
	return fnName, toContractReturnValueType(txType, result[0]), nil
}

func toContractReturnValueType(txType int, val reflect.Value) []byte {
	defer func() {
		if e := recover(); nil != e {
			err := fmt.Errorf("toContractReturnValueType:%+v", e)
			log.Error("toContractReturnValueType", "error", err, "value type", val.Kind())
		}
	}()

	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return toContractReturnValueUintType(txType, val.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return toContractReturnValueIntType(txType, val.Int())
	case reflect.String:
		return toContractReturnValueStringType(txType, []byte(val.String()))
	case reflect.Slice:
		return toContractReturnValueStringType(txType, []byte(val.Bytes()))
		//case reflect.Bool:
		//case reflect.Float64, reflect.Float32:
		// case reflect.Array
	}
	panic("unsupported type")
}

func retrieveFnAndParams(input []byte, fns SCExportFns) (txType int, fnName string, fn SCExportFn, fnParams []reflect.Value, err error) {
	defer func() {
		if e := recover(); nil != e {
			fn, fnParams, err = nil, nil, fmt.Errorf("parse tx data failed:%+v", e)
			log.Error("Failed to parse tx data", "error", err, "input", input)
		}
	}()

	var args [][]byte
	if err := rlp.DecodeBytes(input, &args); nil != err {
		log.Error("Failed to verify input of system contract,Decode rlp input failed", "error", err)
		return 0, "", nil, nil, err
	}
	txType = int(common.BytesToInt64(args[0]))
	fnName = string(args[1])

	var ok bool
	if fn, ok = fns[fnName]; !ok {
		return 0, fnName, nil, nil, errFuncNotFoundInExportFuncs
	}

	fnType := reflect.TypeOf(fn)
	paramNum := fnType.NumIn()
	if paramNum != len(args)-2 {
		log.Warn("params number invalid. ", "expected:", paramNum, "got:", len(args)-2)
		return 0, fnName, nil, nil, errParamsNumInvalid
	}

	fnParams = make([]reflect.Value, paramNum)
	for i := 0; i < paramNum; i++ {
		targetType := fnType.In(i).String()
		inputByte := args[i+2]
		fnParams[i] = byteutil.ConvertBytesTo(inputByte, targetType)
	}

	return txType, fnName, fn, fnParams, nil
}

func CheckPublicKeyFormat(pub string) error {
	b, err := hex.DecodeString(pub)
	if err != nil {
		return err
	}
	// nodeid = pubkey[1:]
	b = append([]byte{4}, b...)
	_, err = crypto.UnmarshalPubkey(b)
	if err != nil {
		return err
	}

	return nil
}
// Name Format
// Length: 2~128
// `^[a-zA-Z0-9_]\w{1,127}$`
func checkNameFormat(name string) (bool, error) {
	b, err := regexp.Match(nameRegPattarn, []byte(name))
	if err != nil{
		return false, err
	}
	return b, nil
}

func checkIpFormat(ip string) (bool, error){
	b, err := regexp.Match(ipRegPattarn, []byte(ip))
	if err != nil{
		return false, err
	}
	return b, nil
}
// Email Format
// xxx@xxx.xxx
// total length <= 64
func checkEmailFormat(email string) (bool, error){
	b, err := regexp.Match(emailRegPattarn, []byte(email))
	if err != nil{
		return false, err
	}

	if len(email) > 64{
		return false, nil
	}
	return b, nil
}

func checkPhoneFormat(phone string) (bool, error) {
	b1, err1 := regexp.Match(telePhonePattarn, []byte(phone))
	b2, err2 := regexp.Match(mobilePhonePattarn, []byte(phone))

	if b1 || b2{
		return true, nil
	}

	if !b1 {
		return b1, err1
	}
	return b2, err2
}

// generate state key compatible with bcwasm state key
func generateStateKey(key string) []byte {
	return bcwasmutil.SerilizString(key)
}

func recoveryStateKey(key []byte) string {
	return bcwasmutil.DeserilizeString(key)
}

