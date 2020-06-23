package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/math"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"math/big"
)

func toContractReturnValueIntType(txType int, res int64) []byte {
	if txType == common.CALL_CANTRACT_FLAG {
		return utils.Int64ToBytes(res)
	}

	bigRes := new(big.Int)
	bigRes.SetInt64(res)
	finalRes := utils.Align32Bytes(math.U256(bigRes).Bytes())
	return finalRes
}

func toContractReturnValueUintType(txType int, res uint64) []byte {
	if txType == common.CALL_CANTRACT_FLAG {
		return utils.Uint64ToBytes(res)
	}

	finalRes := utils.Align32Bytes(utils.Uint64ToBytes((res)))
	return finalRes
}

func toContractReturnValueStringType(txType int, res []byte) []byte {
	if txType == common.CALL_CANTRACT_FLAG || txType == common.TxTypeCallSollCompatibleWasm {
		return res
	}

	var dataRealSize = len(res)
	if (dataRealSize % 32) != 0 {
		dataRealSize = dataRealSize + (32 - (dataRealSize % 32))
	}
	dataByt := make([]byte, dataRealSize)
	copy(dataByt[0:], res)

	strHash := common.BytesToHash(common.Int32ToBytes(32))
	sizeHash := common.BytesToHash(common.Int64ToBytes(int64((len(res)))))

	finalData := make([]byte, 0)
	finalData = append(finalData, strHash.Bytes()...)
	finalData = append(finalData, sizeHash.Bytes()...)
	finalData = append(finalData, dataByt...)

	return finalData
}
