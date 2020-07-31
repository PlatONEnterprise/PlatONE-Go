package vm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/rlp"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

const (
	cnsName    = "cnsManager"
	cnsTotal   = "total"
	cnsCurrent = "current"
)

const seperateChar = ":"

type cnsMap struct {
	StateDB
	contractAddr common.Address
}

func NewCnsMap(stateDB StateDB, contractAddr common.Address) *cnsMap {
	return &cnsMap{stateDB, contractAddr}
}

func (c *cnsMap) setState(key, value interface{}) {

	keyBytes, err := rlp.EncodeToBytes(key)
	if err != nil {
		panic(fmt.Sprintf("setState encode key error: %v, key: %v", err, key))
	}

	valueBytes, err := rlp.EncodeToBytes(value)
	if err != nil {
		panic(fmt.Sprintf("setState encode value error: %v, value: %v", err, value))
	}

	c.SetState(c.contractAddr, keyBytes, valueBytes)
}

//
func (c *cnsMap) getState(key, value interface{}) {

	keyBytes, err := rlp.EncodeToBytes(key)
	if err != nil {
		panic(fmt.Sprintf("getState encode key error: %v, key: %v", err, key))
	}

	valueBytes := c.GetState(c.contractAddr, keyBytes)
	if len(valueBytes) == 0 {
		return
	}

	err = rlp.DecodeBytes(valueBytes, value)
	if err != nil {
		panic(fmt.Sprintf("getState dencode value error: %v, value: %v", err, valueBytes))
	}
}

func (c *cnsMap) getKeyByIndex(index uint64) string {
	var key string
	c.getState(indexWrapper(index), &key)
	return key
}

func (c *cnsMap) find(key string) *ContractInfo {
	var cInfo *ContractInfo
	c.getState(key, &cInfo)
	return cInfo
}

func (c *cnsMap) total() uint64 {
	var total uint64
	// when there is no corresponding value to the <totalWrapper()> key,
	// the total is 0
	c.getState(totalWrapper(), &total)
	return total
}

func (c *cnsMap) get(index uint64) *ContractInfo {
	key := c.getKeyByIndex(index)
	if key == "" {
		return nil
	}

	return c.find(key)
}

func (c *cnsMap) insert(key string, value *ContractInfo) {
	total := c.total()
	c.setState(key, value)
	c.setState(indexWrapper(total), key)
	c.setState(totalWrapper(), total+1)
}

func (c *cnsMap) update(key, value []byte) {
	c.setState(key, value)
}

func (c *cnsMap) getCurrentVer(name string) string {
	var curVersion = "0.0.0.0"
	c.getState(currentVerWrapper(name), &curVersion)
	return curVersion
}

func (c *cnsMap) setCurrentVer(name, ver string) {
	c.setState(currentVerWrapper(name), []byte(ver))
}

func currentVerWrapper(name string) []byte {
	return []byte(cnsName + cnsCurrent + name)
}

func indexWrapper(index uint64) string {
	return cnsName + strconv.FormatUint(index, 10)
}

func totalWrapper() string {
	return cnsName + cnsTotal
}

func getSearchKey(name, version string) string {
	return name + seperateChar + version
}

/*
func (c *cnsMap) isNameRegByOthers_Old(name string, origin common.Address) bool {
	var index uint64

	for index = 0; index < c.total(); index++ {
		cnsInfo := c.get(index)
		if cnsInfo.Name == name && cnsInfo.Origin != origin {
			return true
		}
	}

	return false
}*/

func (c *cnsMap) isNameRegByOthers(name string, origin common.Address) bool {
	var index uint64

	for index = 0; index < c.total(); index++ {
		key := c.getKeyByIndex(index)
		existedName := strings.Split(key, seperateChar)[0]
		if existedName == name {
			cnsInfo := c.find(key)
			if cnsInfo.Origin != origin {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

/*
func (c *cnsMap) getLargestVersion_Old(name string) string {
	tempVersion := "0.0.0.0"
	var index uint64

	for index = 0; index < c.total(); index++ {
		cnsInfo := c.get(index)
		if cnsInfo.Name == name {
			if verCompare(cnsInfo.Version, tempVersion) == 1 {
				tempVersion = cnsInfo.Version
			}
		}
	}

	return tempVersion
}*/

func (c *cnsMap) getLargestVersion(name string) string {
	var index uint64 = 0
	tempVersion := "0.0.0.0"

	for ; index < c.total(); index++ {
		// no possibility for the key to be null
		// therefore, there is no param checking
		key := c.getKeyByIndex(index)

		ary := strings.Split(string(key), seperateChar)
		existedName := ary[0]
		existedVersion := ary[1]

		if existedName == name {
			if verCompare(existedVersion, tempVersion) == 1 {
				tempVersion = existedVersion
			}
		}
	}

	return tempVersion
}
