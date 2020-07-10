package vm

import (
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"

	"github.com/stretchr/testify/assert"
)

func TestCnsManager_cMap(t *testing.T) {
	assert.Equal(t, key[1], cns.cMap.getKey(1), "cns getKey FAILED")
	assert.Equal(t, testCases[0], cns.cMap.find(key[0]), "cns find() FAILED")
	assert.Equal(t, len(testCases), cns.cMap.total(), "cns total() FAILED")
}

// cnsTestInital prepares the data for the unit test
func cnsTestInital() {
	db := newMockStateDB()
	addr := common.HexToAddress("")

	cns = &CnsManager{
		cMap:       NewCnsMap(db, addr),
		callerAddr: common.HexToAddress(testCaller),
		origin:     common.HexToAddress(testOrigin),
		isInit:     -1,
	}

	for _, data := range testCases {
		value, err := data.encode()
		if err != nil {
			// m.Fatalf(err.Error())
		}

		k := getSearchKey(data.Name, data.Version)
		cns.cMap.insert(k, value)
		cns.cMap.updateLatestVer(data.Name, data.Version)

		key = append(key, k)
	}
}
