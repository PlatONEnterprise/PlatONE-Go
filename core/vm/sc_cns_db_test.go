package vm

import (
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"

	"github.com/stretchr/testify/assert"
)

func TestCnsManager_cMap(t *testing.T) {
	assert.Equal(t, key[1], cns.base.cMap.getKeyByIndex(1), "cns getKey FAILED")
	assert.Equal(t, testCases[0], cns.base.cMap.find(key[0]), "cns find() FAILED")
	assert.Equal(t, uint64(len(testCases)), cns.base.cMap.total(), "cns total() FAILED")
}

// cnsTestInitial prepares the data for the unit test
func cnsTestInitial() {
	db := newMockStateDB()
	addr := common.HexToAddress("")

	base := &CnsManager{
		cMap:        NewCnsMap(db, addr),
		caller:      testCaller,
		origin:      testOrigin,
		isInit:      -1,
		blockNumber: big1,
	}
	cns.base = base

	for _, data := range testCases {

		k := getSearchKey(data.Name, data.Version)
		cns.base.cMap.insert(k, data)
		cns.base.cMap.setCurrentVer(data.Name, data.Version)

		key = append(key, k)
	}
}
