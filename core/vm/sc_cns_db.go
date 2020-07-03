package vm

import (
	"strconv"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

const (
	cnsName   = "cnsManager"
	cnsTotal  = "total"
	cnsLatest = "latest"
)

type cnsMap struct {
	StateDB
	codeAddr *common.Address
}

func NewCnsMap(db StateDB, addr *common.Address) *cnsMap {
	return &cnsMap{db, addr}
}

func (c *cnsMap) setState(key, value []byte) {
	c.SetState(*c.codeAddr, key, value)
}

func (c *cnsMap) getState(key []byte) []byte {
	return c.GetState(*c.codeAddr, key)
}

func (c *cnsMap) getKey(index int) []byte {
	indexStr := strconv.Itoa(index)
	value := c.getState(wrapper(indexStr))

	return value
}

func (c *cnsMap) find(key []byte) *ContractInfo {
	value := c.getState(key)
	if value == nil {
		return nil
	}

	cnsInfo, _ := decodeCnsInfo(value)
	return cnsInfo
}

func (c *cnsMap) get(index int) *ContractInfo {
	value := c.getKey(index)
	if value == nil {
		return nil
	}

	return c.find(value)
}

func (c *cnsMap) total() int {
	value := c.getState(totalWrapper())

	if value == nil || len(value) == 0 {
		return 0
	}

	totalStr := string(value)
	total, _ := strconv.Atoi(totalStr)
	return total
}

func (c *cnsMap) insert(key, value []byte) {
	total := c.total()
	index := strconv.Itoa(total)

	c.setState(key, value)
	c.setState(wrapper(index), key)

	update := strconv.Itoa(total + 1)
	c.setState(totalWrapper(), []byte(update))
}

func (c *cnsMap) update(key, value []byte) {
	c.setState(key, value)
}

func wrapper(str string) []byte {
	return []byte(cnsName + str)
}

func totalWrapper() []byte {
	return []byte(cnsName + cnsTotal)
}

func (cMap *cnsMap) isNameRegByOthers(name, origin string) bool {
	for index := 0; index < cMap.total(); index++ {
		cnsInfo := cMap.get(index)
		if cnsInfo.Name == name && cnsInfo.Origin != origin {
			return true
		}
	}

	return false
}

func (c *cnsMap) isNameRegByOthers_Method2(name, origin string) bool {
	for index := 0; index < c.total(); index++ {
		key := c.getKey(index)
		existedName := strings.Split(string(key), ":")[0]
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

func (c *cnsMap) getLargestVersion(name string) string {
	tempVersion := "0.0.0.0"

	for index := 0; index < c.total(); index++ {
		cnsInfo := c.get(index)
		if cnsInfo.Name == name {
			if verCompare(cnsInfo.Version, tempVersion) == 1 {
				tempVersion = cnsInfo.Version
			}
		}
	}

	return tempVersion
}

func (c *cnsMap) getLargestVersion_Method2(name string) string {
	tempVersion := "0.0.0.0"

	for index := 0; index < c.total(); index++ {
		key := c.getKey(index)
		existedName := strings.Split(string(key), ":")[0]
		existedVersion := strings.Split(string(key), ":")[1]
		if existedName == name {
			if verCompare(existedVersion, tempVersion) == 1 {
				tempVersion = existedVersion
			}
		}
	}

	return tempVersion
}

func latestWrapper(name string) []byte {
	return []byte(cnsName + cnsLatest + name)
}

func (c *cnsMap) getLatestVer(name string) string {
	ver := c.getState(latestWrapper(name))
	return string(ver)
}

func (c *cnsMap) updateLatestVer(name, ver string) {
	c.setState(latestWrapper(name), []byte(ver))
}
