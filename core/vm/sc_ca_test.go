package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"testing"
)

func TestCa_setRootCA(t *testing.T) {
	db := newMockStateDB()
	addr := syscontracts.CAManagementAddress
	c := CAManager{stateDB:db, contractAddr:addr}
	c.setRootCA("/home/night/go/src/github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/cmd/selfCA.PEM")
	c.addIssuer("/home/night/go/src/github.com/PlatONEnetwork/PlatONE-Go/release/linux/bin/targetCA.PEM")
	list, _ := c.getList()
	//rootCA, _ := c.getRootCA()
	//rootCAString, _ := rootCA.GetPEM()

	println("%s", list[0])
	//addr1 := syscontracts.UserManagementAddress
	//caller := common.HexToAddress("0x62fb664c49cfa4fa35931760c704f9b3ab664666")
	//um := UserManagement{stateDB: db, caller: caller, contractAddr: addr1, blockNumber: big.NewInt(100)}
	//um.setSuperAdmin()
	//um.addChainAdminByAddress(caller)
	//p := ParamManager{contractAddr: &addr, stateDB: db, caller: caller, blockNumber: big.NewInt(100)}
	//p.setGasContractName("abc")
	//ret, err := p.getGasContractName()
	//if nil != err {
	//	t.Error(err)
	//	return
	//}
	//t.Logf("%s", ret)
}

func TestCa_GetCA(t *testing.T) {
	db := newMockStateDB()
	addr := syscontracts.CAManagementAddress
	c := CAManager{stateDB:db, contractAddr:addr}
	c.setRootCA("/home/night/go/src/github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/cmd/selfCA.PEM")
	//list, _ := c.getList()
	subject := "/C=CN/O=wxbc/CN=test1"
	res, _ := c.getCA(subject)
	cert, _ := res.GetPEM()
	//rootCA, _ := c.getRootCA()
	//rootCAString, _ := rootCA.GetPEM()

	println(cert)
	//addr1 := syscontracts.UserManagementAddress
	//caller := common.HexToAddress("0x62fb664c49cfa4fa35931760c704f9b3ab664666")
	//um := UserManagement{stateDB: db, caller: caller, contractAddr: addr1, blockNumber: big.NewInt(100)}
	//um.setSuperAdmin()
	//um.addChainAdminByAddress(caller)
	//p := ParamManager{contractAddr: &addr, stateDB: db, caller: caller, blockNumber: big.NewInt(100)}
	//p.setGasContractName("abc")
	//ret, err := p.getGasContractName()
	//if nil != err {
	//	t.Error(err)
	//	return
	//}
	//t.Logf("%s", ret)
}
