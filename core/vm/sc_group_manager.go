package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

var (
	ErrRepeatedGroupID = errors.New("Repeated GroupID ")
)

const (
	groupKey  = "c9373997b64ef7ab34be47746b83f3bbad9ed86e"
	groupList = "b78adefaefdbf2ace61af534bbfe5e6d2e58682d"
)

type GroupManagement struct {
	stateDB      StateDB
	caller       common.Address // msg.From()	contract.caller
	blockNumber  *big.Int
	contractAddr common.Address
}

type GroupInfo struct {
	Creator      string   `json:"creator"`
	GroupID      uint64   `json:"groupID"`
	CreatorEnode string   `json:"creatorEnode"`
	BootNodes    []string `json:"bootNodes"`
}

func (g GroupInfo) String() string {
	data, _ := json.Marshal(g)
	return string(data)
}

func (g *GroupManagement) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.UserManagementGas
}

// Run runs the precompiled contract
func (g *GroupManagement) Run(input []byte) ([]byte, error) {
	fnName, ret, err := execSC(input, g.AllExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
		g.emitEvent(fnName, operateFail, err.Error())
	}
	return ret, nil
}

func (g *GroupManagement) setState(key, value []byte) {
	g.stateDB.SetState(g.contractAddr, key, value)
}
func (g *GroupManagement) getState(key []byte) []byte {
	value := g.stateDB.GetState(g.contractAddr, key)
	return value
}

func (g *GroupManagement) Caller() common.Address {
	return g.caller
}

//for access control
func (g *GroupManagement) AllExportFns() SCExportFns {
	return SCExportFns{
		"hasGroupOpPermission": g.hasGroupOpPermission,
		"createGroup":          g.createGroup,
		"getAllGroups":         g.getAllGroups,
		"getGroupByID":         g.getGroupByID,
		"updateBootNodes":      g.updateBootNodes,
		"addBootNode":          g.addBootNode,
		"delBootNode":          g.delBootNode,
	}
}

// export functions
func (g *GroupManagement) hasGroupOpPermission() (int32, error) {
	if hasGroupCreatePermission(g.stateDB, g.caller) {
		return 1, nil
	}
	return 0, nil
}

func (g *GroupManagement) createGroup(groupInfo string) (int32, error) {
	if ok, _ := g.hasGroupOpPermission(); ok != 1 {
		return 0, errNoPermission
	}
	group := GroupInfo{}
	err := json.Unmarshal([]byte(groupInfo), &group)
	if err != nil {
		return -1, err
	}
	group.Creator = g.Caller().String()
	if err := g.addGroup(group); err != nil {
		return -1, err
	}
	return 0, nil
}
func (g *GroupManagement) getAllGroups() (string, error) {
	groups, err := g.getGroupList()
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(groups)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (g *GroupManagement) getGroupByID(groupID uint64) (string, error) {
	group, err := g.getGroupInfo(groupID)
	if err != nil {
		return "", err
	}
	return group.String(), nil
}

func (g *GroupManagement) updateBootNodes(groupID uint64, nodes string) (int32, error) {
	group, err := g.getGroupInfo(groupID)
	if err != nil {
		return -1, err
	}
	if group.Creator != g.Caller().String() {
		return -1, errNoPermission
	}
	var bootNodes []string
	err = json.Unmarshal([]byte(nodes), &bootNodes)
	if err != nil {
		return -1, err
	}

	group.BootNodes = bootNodes

	err = g.updateGroupInfo(*group)
	if err != nil {
		return -1, err
	}
	return 0, nil
}

func (g *GroupManagement) addBootNode(groupID uint64, node string) (int32, error) {
	group, err := g.getGroupInfo(groupID)
	if err != nil {
		return -1, err
	}
	if group.Creator != g.Caller().String() {
		return -1, errNoPermission
	}
	for _, n := range group.BootNodes {
		if n == node {
			return -1, nil
		}
	}
	group.BootNodes = append(group.BootNodes, node)

	if err := g.updateGroupInfo(*group); err != nil {
		return -1, err
	}
	return 0, nil
}
func (g *GroupManagement) delBootNode(groupID uint64, node string) (int32, error) {
	group, err := g.getGroupInfo(groupID)
	if err != nil {
		return -1, err
	}
	if group.Creator != g.Caller().String() {
		return -1, errNoPermission
	}
	pos := -1
	for i, n := range group.BootNodes {
		if n == node {
			pos = i
		}
	}
	if pos != -1 {
		group.BootNodes = append(group.BootNodes[:pos], group.BootNodes[pos+1:]...)
		if err := g.updateGroupInfo(*group); err != nil {
			return -1, err
		}
	}

	return 0, nil
}

// internal functions
func (g *GroupManagement) addGroup(info GroupInfo) error {
	groups, err := g.getGroupList()
	if err != nil {
		return err
	}

	for _, g := range groups {
		if g.GroupID == info.GroupID {
			return ErrRepeatedGroupID
		}
	}

	if err := g.storeGroupInfo(&info); err != nil {
		return err
	}
	if err := g.addGroupToList(&info); err != nil {
		return err
	}

	return nil
}

func (g *GroupManagement) updateGroupInfo(info GroupInfo) error {
	groups, err := g.getGroupList()
	if err != nil {
		return err
	}

	for _, g := range groups {
		if g.GroupID == info.GroupID {
			g.BootNodes = info.BootNodes
		}
	}

	if err := g.storeGroupInfo(&info); err != nil {
		return err
	}
	if err := g.updateGroupList(groups); err != nil {
		return err
	}

	return nil
}

func (g *GroupManagement) storeGroupInfo(info *GroupInfo) error {
	groupKey := generateGroupKey(info.GroupID)
	rawData, err := json.Marshal(info)
	if err != nil {
		return err
	}
	g.setState(groupKey, rawData)
	return nil
}

func (g *GroupManagement) getGroupInfo(id uint64) (*GroupInfo, error) {
	groupKey := generateGroupKey(id)

	rawData := g.getState(groupKey)
	group := &GroupInfo{}

	if err := json.Unmarshal(rawData, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (g *GroupManagement) storeGroupList(infos []GroupInfo) error {
	rawData, err := json.Marshal(infos)
	if err != nil {
		return err
	}
	g.setState([]byte(groupList), rawData)
	return nil
}

func generateGroupKey(id uint64) []byte {
	key := fmt.Sprintf("%s:%d", groupKey, id)
	return []byte(key)
}

func (g *GroupManagement) delGroupInfo(id uint64) error {
	if err := g.delGroupfromList(id); err != nil {
		return err
	}
	g.setState(generateGroupKey(id), nil)

	return nil
}

func (g *GroupManagement) addGroupToList(info *GroupInfo) error {
	groups, err := g.getGroupList()
	if err != nil {
		return err
	}

	for _, g := range groups {
		if g.GroupID == info.GroupID {
			return ErrRepeatedGroupID
		}
	}

	groups = append(groups, *info)
	return g.storeGroupList(groups)
}

func (g *GroupManagement) updateGroupList(groups []GroupInfo) error {
	return g.storeGroupList(groups)
}

func (g *GroupManagement) delGroupfromList(id uint64) error {
	groups, err := g.getGroupList()
	if err != nil {
		return err
	}

	pos := -1
	for i, g := range groups {
		if g.GroupID == id {
			pos = i
			break
		}
	}
	if pos > -1 {
		groups = append(groups[:pos], groups[pos+1:]...)
		if err := g.storeGroupList(groups); err != nil {
			return err
		}
	}
	return nil
}

func (g *GroupManagement) getGroupList() ([]GroupInfo, error) {
	data := g.getState([]byte(groupList))
	if len(data) == 0 {
		return nil, nil
	}

	var groups []GroupInfo
	err := json.Unmarshal(data, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (g *GroupManagement) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(syscontracts.GroupManagementAddress, g.stateDB, g.blockNumber.Uint64(), topic, code, msg)
}
