package vm

import (
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

/*
func (c *cnsMap) setState(key, value []byte) {
	c.SetState(c.contractAddr, key, value)
}

func (c *cnsMap) getState(key []byte) []byte {
	return c.GetState(c.contractAddr, key)
}*/

func (c *cnsMap) setState(key, value interface{}) {

	keyBytes, err := rlp.EncodeToBytes(key)
	if err != nil {
		// todo: panic?
	}

	valueBytes, err := rlp.EncodeToBytes(value)
	if err != nil {
		// todo: panic?
	}

	c.SetState(c.contractAddr, keyBytes, valueBytes)
}

func (c *cnsMap) getState(key interface{}) []byte {

	keyBytes, err := rlp.EncodeToBytes(key)
	if err != nil {
		// todo
	}

	return c.GetState(c.contractAddr, keyBytes)
}

// todo: if could optimize the getState() by go reflect
func (c *cnsMap) getKeyByIndex(index uint64) string {
	value := c.getState(indexWrapper(index))
	if len(value) == 0 {
		return ""
	}

	var result string

	err := rlp.DecodeBytes(value, &result)
	if err != nil {
		// todo: panic
	}

	return result
}

func (c *cnsMap) find(key string) *ContractInfo {
	value := c.getState(key)
	if len(value) == 0 {
		return nil
	}

	var result ContractInfo

	err := rlp.DecodeBytes(value, &result)
	if err != nil {
		// todo: panic
	}

	return &result
}

func (c *cnsMap) get(index uint64) *ContractInfo {
	value := c.getKeyByIndex(index)
	if value == "" {
		return nil
	}

	return c.find(value)
}

func (c *cnsMap) total() uint64 {
	value := c.getState(totalWrapper())
	if len(value) == 0 {
		return 0
	}

	var result uint64

	err := rlp.DecodeBytes(value, &result)
	if err != nil {
		// todo: panic
	}

	return result
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

func currentVerWrapper(name string) []byte {
	return []byte(cnsName + cnsCurrent + name)
}

func (c *cnsMap) getCurrentVer(name string) string {
	value := c.getState(currentVerWrapper(name))
	if len(value) == 0 {
		return ""
	}

	var result string

	err := rlp.DecodeBytes(value, &result)
	if err != nil {
		// todo: panic
	}

	return result
}

func (c *cnsMap) setCurrentVer(name, ver string) {
	c.setState(currentVerWrapper(name), []byte(ver))
}
