package utils

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
)

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

	return nil, errors.New("address convert failed: invalid address")
}
