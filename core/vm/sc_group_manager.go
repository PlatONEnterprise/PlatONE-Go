package vm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

var (
	ErrRepeatedGroupID = errors.New("Repeated GroupID")
)

const (
	groupKey  = "c9373997b64ef7ab34be47746b83f3bbad9ed86e"
	groupList = "b78adefaefdbf2ace61af534bbfe5e6d2e58682d"
)

type GroupManagement struct {
	state   StateDB
	caller  common.Address
	address common.Address
}

type GroupInfo struct {
	Creator   string `json:"creator"`
	GroupID   uint64 `json:"groupID"`
	BootNodes string `json:"bootNodes"`
}

func (g *GroupManagement) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.UserManagementGas
}

// Run runs the precompiled contract
func (g *GroupManagement) Run(input []byte) ([]byte, error) {
	return execSC(input, g.AllExportFns())
}

func (g *GroupManagement) setState(key, value []byte) {
	g.state.SetState(g.address, key, value)
}
func (g *GroupManagement) getState(key []byte) []byte {
	value := g.state.GetState(g.address, key)
	return value
}

func (g *GroupManagement) Caller() common.Address {
	return g.caller
}

//for access control
func (g *GroupManagement) AllExportFns() SCExportFns {
	return SCExportFns{}
}

// export functions
func (g *GroupManagement) hasGroupOpPermisson() (int32, error) {
	if hasGroupCreatePermmision(g.state, g.caller) {
		return 1, nil
	}
	return 0, nil
}

func (g *GroupManagement) createGroup(groupINfo string) ([]byte, error) {
	if ok, _ := g.hasGroupOpPermisson(); ok != 1 {
		return nil, ErrNoPermission
	}
	group := &GroupInfo{}
	err := json.Unmarshal([]byte(groupINfo), group)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (g *GroupManagement) getAllGroups(groupINfo string) (string, error) {

	return "", nil
}

func (g *GroupManagement) getGroupByGroupID(groupID uint64) (string, error) {

	return "", nil
}

// internal functions

func (g *GroupManagement) addGroup(info GroupInfo) error {
	if p, _ := g.hasGroupOpPermisson(); p != 1 {
		return ErrNoPermission
	}

	groups, err := g.getGroupList()
	if err != nil {
		return err
	}

	for _, g := range groups {
		if g.GroupID == info.GroupID {
			return ErrRepeatedGroupID
		}
	}

	g.storeGroupInfo(&info)
	g.addGroupToList(&info)

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
	g.storeGroupList(groups)
	return nil
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
		g.storeGroupList(groups)
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
