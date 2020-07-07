package utils

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"math/big"
	"strconv"
	"strings"
)

// ChainParamConvert convert the string to chain defined type
func ChainParamConvert(param, paramName string) interface{} {
	var err error
	var i interface{}

	switch paramName {
	case "value", "gasPrice":
		i, err = IntValueConvert(param)
	case "gas":
		i, err = UintValueConvert(param)
	case "address", "to", "from":
		i, err = AddressConvert(param)
	default:
		i, err = param, nil //TODO
	}

	if err != nil {
		utils.Fatalf(ErrParamParseFormat, paramName, err.Error())
	}

	return i
}

//TODO optimize ?
func IntValueConvert(value string) (string, error) {
	var err error
	var intValue int64

	if value == "" {
		return "", nil
	}

	//TODO
	if !strings.HasPrefix(value, "0x") {
		intValue, err = strconv.ParseInt(value, 10, 64)
	} else {
		intValue, err = strconv.ParseInt(value, 0, 64)
	}

	value = hexutil.EncodeBig(big.NewInt(intValue))

	return value, err
}

func UintValueConvert(value string) (string, error) {
	var err error
	var intValue uint64

	if value == "" {
		return "", nil
	}

	if !strings.HasPrefix(value, "0x") {
		intValue, err = strconv.ParseUint(value, 10, 64)
	} else {
		intValue, err = strconv.ParseUint(value, 0, 64)
	}

	value = hexutil.EncodeUint64(intValue)

	return value, err
}

// AddressConvert converts hex string format address to byte format address
// when the input is null, the output address is [0 0 ... 0]
func AddressConvert(address string) (interface{}, error) {
	//ParamValid(address, "to")
	if address == "" || IsMatch(address, "address") {
		return common.HexToAddress(address), nil
	}

	return nil, fmt.Errorf("TODO")
}
