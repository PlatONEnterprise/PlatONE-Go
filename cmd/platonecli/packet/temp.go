package packet

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

// todo: handle error
// ExtractContractData extract the role info from the contract return result
func ExtractContractData(result, role string) string {
	var inter = make([]interface{}, 0)
	var count int

	r, _ := ParseSysContractResult([]byte(result))
	data := r.Data.([]interface{})

	length := len(data)
	for i := 0; i < length; i++ {
		temp, _ := json.Marshal(data[0])
		if strings.Contains(string(temp), role) {
			inter = append(inter, data[i])
			count++
		}
	}

	if count == 0 {
		return fmt.Sprintf("no %s in registration\n", role)
	} else {
		r.Data = inter
		newContractData, _ := json.Marshal(r)
		return string(newContractData)
	}
}

//========================== StringToType ==============================

func SolInputTypeConversionV2(input abi.ArgumentMarshaling, v string) (interface{}, error) {
	switch {
	case strings.HasPrefix(input.Type, "address"):
		return common.HexToAddress(v), nil
	case strings.HasPrefix(input.Type, "int") || strings.HasPrefix(input.Type, "uint"):
		parts := regexp.MustCompile(`(u)?int([0-9]*)`).FindStringSubmatch(input.Type)
		switch parts[2] {
		case "8":
			return abi.SolInputStringTOInt(v, 8, parts[1] == "")
		case "16":
			return abi.SolInputStringTOInt(v, 16, parts[1] == "")
		case "32":
			return abi.SolInputStringTOInt(v, 32, parts[1] == "")
		case "64":
			return abi.SolInputStringTOInt(v, 64, parts[1] == "")
		case "128", "256":
			if parts[1] != "" && strings.Contains(v, "-") {
				return nil, fmt.Errorf("value does not match type: Unsigned type passes negative number")
			}
			value, ok := big.NewInt(0).SetString(v, 10)
			bit, _ := strconv.Atoi(parts[2])
			if !ok || !common.IsSafeNumber(v, bit, parts[1] != "") {
				return nil, fmt.Errorf("paring big int string error")
			}
			return value, nil
		}
		return nil, errors.New("parse input type int has err bitsize")
	case strings.HasPrefix(input.Type, "bool"):
		if v == "false" {
			return false, nil
		} else if v == "true" {
			return true, nil
		} else {
			return false, errors.New("parse bool type error")
		}
	case strings.HasPrefix(input.Type, "string"):
		return v, nil
	case strings.HasPrefix(input.Type, "tuple"):
		var argsTup = make([]interface{}, 0)
		tupleArray := utils.GetFuncParams(v)

		// todo: refactor the code
		for i, vTup := range tupleArray {
			argTup, err := SolInputTypeConversionV2(input.Components[i], vTup)
			if err != nil {
				return nil, err
			}

			argsTup = append(argsTup, argTup)
		}

		return argsTup, nil
	default:
		return nil, errors.New("sol input type error")

		//case strings.HasPrefix(t, "bytes"):
		//	if len(v) < 3 {
		//		return nil, fmt.Errorf("input format error: %s", v)
		//	}
		//
		//	v = v[1 : len(v)-1]
		//	vs := strings.Split(v, ",")
		//	var res []byte
		//	for _, value := range vs {
		//		intV, err := strconv.Atoi(value)
		//		if err != nil || intV > 255 {
		//			return nil, fmt.Errorf("bytes input strconv a to i error || value > 255 : %s", value)
		//		}
		//		res = append(res, byte(intV))
		//	}
		//	return res, nil
		//todo: 反生成数组形式
		//parts := regexp.MustCompile(`bytes([0-9]*)`).FindStringSubmatch(t)
		//if parts[1] != "" {
		//	length, err := strconv.Atoi(parts[1])
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	if len(v) < 3 {
		//		return nil, fmt.Errorf("input format error: %s", v)
		//	}
		//
		//	v = v[1 : len(v)-1]
		//	vs := strings.Split(v, ",")
		//	if len(vs) != length {
		//		return nil, fmt.Errorf("input format error: %s", v)
		//	}
		//}
	}
}
