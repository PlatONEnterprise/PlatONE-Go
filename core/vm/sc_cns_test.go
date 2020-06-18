package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestLatestVersion (t *testing.T) {
	testCase := []struct {
		ver1 string
		ver2 string
	}{
		{"0.0.0.0", "0.0.2.0"},
		{"1.0.0.0","0.0.0.1"},
	}

	for _, data := range testCase {
		if verCompare(data.ver1, data.ver2) == 1 {
			t.Logf("ver1 %s is larger than ver2 %s\n", data.ver1, data.ver2)
		} else {
			t.Logf("ver1 %s is smaller than ver2 %s\n", data.ver1, data.ver2)
		}
	}
}

//func TestMain(m *testing.M) {
	// evm := NewEVM()
//}

func TestSerializeCnsInfo(t *testing.T) {
	cnsInfoArray := make([]*ContractInfo, 0)
	cnsInfo := newContractInfo("tofu", "0.0.0.1", "0x123", "0x123")
	cnsInfoArray = append(cnsInfoArray, cnsInfo, cnsInfo, cnsInfo)

	sBytes, err := serializeCnsInfo(CODE_OK, MSG_OK, cnsInfoArray)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s\n", sBytes)
}

/*
func TestCnsFunc(t *testing.T) {
	fnNameInput := "cnsRegister"
	var input= MakeInput(fnNameInput, "aaa", "0.0.0.1", "0x00002")
	ret, err := execSC(input, (&CnsManager{}).AllExportFns())
	if nil != err {
		t.Error(err)
		return
	}

	assert.Equal(t, []byte("aaabbb"), ret)
}*/

func TestCnsManager_cMap(t *testing.T) {
	db := newMockStateDB()
	addr := common.HexToAddress("")

	cns := &CnsManager{
		cMap: NewCnsMap(db, &addr),
	}

	testCases := []*ContractInfo{
		{
			Name: 		"tofu",
			Version: 	"0.0.0.1",
			Address:	"0x123",
			Origin:		"0x000",
			Enabled:	true,
		},
		{
			Name: 		"tofu",
			Version: 	"0.0.0.2",
			Address:	"0x456",
			Origin:		"0x000",
			Enabled:	true,
		},
		{
			Name: 		"tofu",
			Version: 	"0.0.0.3",
			Address:	"0x789",
			Origin:		"0x000",
			Enabled:	true,
		},
		{
			Name: 		"tofu",
			Version: 	"0.0.0.4",
			Address:	"0x102",
			Origin:		"0x000",
			Enabled:	false,
		},
		{
			Name: 		"bob",
			Version: 	"0.0.0.1",
			Address:	"0x123",
			Origin:		"0x000",
			Enabled:	false,
		},
	}

	var key = make([][]byte, 0)

	for _, data := range testCases {
		value, _ := rlp.EncodeToBytes(data)
		k := getSearchKey(data.Name, data.Version)
		cns.cMap.insert(k, value)
		key = append(key, k)
	}

	assert.Equal(t, key[1], cns.cMap.getKey(1),  "cns getKey equal")
	assert.Equal(t, testCases[0], cns.cMap.find(key[0]),  "cns find() equal")
	assert.Equal(t, len(testCases), cns.cMap.total(),  "cns total() equal")
	//cns.cMap.update()
	//cns.cMap.get()

	nameT := testCases[0].Name
	verT := testCases[0].Version
	assert.Equal(t, verT, cns.cMap.getLatestVersion(nameT),  "getLatestVersion equal")

	//fmt.Println(db.mockDB)
}

func newMockStateDB() *mockStateDB{
	return &mockStateDB{
		mockDB:		make(map[common.Address]map[string][]byte),
	}
}

type mockStateDB struct {
	mockDB	map[common.Address]map[string][]byte
}

func (m *mockStateDB) GetState(addr common.Address, key []byte) []byte {
	// panic("implement me")

	//temp := m.mockDB[addr]
	//fmt.Println(temp)

	//return temp[string(key)]

	return m.mockDB[addr][string(key)]
}

func (m *mockStateDB) SetState(addr common.Address, key []byte, value []byte) {
	//panic("implement me")
	//fmt.Printf("----------------------%v %v,%v, %v\n", m.mockDB, addr, key, value)

	if m.mockDB[addr] == nil {
		m.mockDB[addr] = make(map[string][]byte)
	}

	//tempMap := make(map[string][]byte)
	m.mockDB[addr][string(key)] = value
	//m.mockDB[addr] = tempMap
}

func (m *mockStateDB) CreateAccount(common.Address) {
	panic("implement me")
}

func (m *mockStateDB) SubBalance(common.Address, *big.Int) {
	panic("implement me")
}

func (m *mockStateDB) AddBalance(common.Address, *big.Int) {
	panic("implement me")
}

func (m *mockStateDB) GetBalance(common.Address) *big.Int {
	panic("implement me")
}

func (m *mockStateDB) GetNonce(common.Address) uint64 {
	panic("implement me")
}

func (m *mockStateDB) SetNonce(common.Address, uint64) {
	panic("implement me")
}

func (m *mockStateDB) GetCodeHash(common.Address) common.Hash {
	panic("implement me")
}

func (m *mockStateDB) GetCode(common.Address) []byte {
	panic("implement me")
}

func (m *mockStateDB) SetCode(common.Address, []byte) {
	panic("implement me")
}

func (m *mockStateDB) GetCodeSize(common.Address) int {
	panic("implement me")
}

func (m *mockStateDB) GetAbiHash(common.Address) common.Hash {
	panic("implement me")
}

func (m *mockStateDB) GetAbi(common.Address) []byte {
	panic("implement me")
}

func (m *mockStateDB) SetAbi(common.Address, []byte) {
	panic("implement me")
}

func (m *mockStateDB) AddRefund(uint64) {
	panic("implement me")
}

func (m *mockStateDB) SubRefund(uint64) {
	panic("implement me")
}

func (m *mockStateDB) GetRefund() uint64 {
	panic("implement me")
}

func (m *mockStateDB) GetCommittedState(common.Address, []byte) []byte {
	panic("implement me")
}

func (m *mockStateDB) Suicide(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) HasSuicided(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) Exist(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) Empty(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) RevertToSnapshot(int) {
	panic("implement me")
}

func (m *mockStateDB) Snapshot() int {
	panic("implement me")
}

func (m *mockStateDB) AddLog(*types.Log) {
	panic("implement me")
}

func (m *mockStateDB) AddPreimage(common.Hash, []byte) {
	panic("implement me")
}

func (m *mockStateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {
	panic("implement me")
}

func (m *mockStateDB) FwAdd(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) FwClear(contractAddr common.Address, action state.Action) {
	panic("implement me")
}

func (m *mockStateDB) FwDel(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) FwSet(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) SetFwStatus(contractAddr common.Address, status state.FwStatus) {
	panic("implement me")
}

func (m *mockStateDB) GetFwStatus(contractAddr common.Address) state.FwStatus {
	panic("implement me")
}

func (m *mockStateDB) SetContractCreator(contractAddr common.Address, creator common.Address) {
	panic("implement me")
}

func (m *mockStateDB) GetContractCreator(contractAddr common.Address) common.Address {
	panic("implement me")
}

func (m *mockStateDB) OpenFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m *mockStateDB) CloseFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m *mockStateDB) IsFwOpened(contractAddr common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) FwImport(contractAddr common.Address, data []byte) error {
	panic("implement me")
}
