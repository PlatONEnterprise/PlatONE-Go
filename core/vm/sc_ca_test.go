package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"log"
	"testing"
)

func TestCa_setRootCA(t *testing.T) {
	db := newMockStateDB()
	addr := syscontracts.CAManagementAddress
	c := CAManager{stateDB:db, contractAddr:addr}
	c.setRootCert("/home/night/go/src/github.com/PlatONEnetwork/PlatONE-Go/release/linux/bin/nightout.PEM")
	c.addIssuer("/home/night/go/src/github.com/PlatONEnetwork/PlatONE-Go/release/linux/bin/outCert.PEM")
	res, _ := c.getAllCert()

	rest, _ := c.getCert("/C=CN/O=wxbc/CN=test")
	re , _ := rest.GetPEM()
	list, _ := c.getList()
	//rootCA, _ := c.getRootCA()
	//rootCAString, _ := rootCA.GetPEM()
	//resPem, _ := res.GetPEM()
	fmt.Printf(re)
	fmt.Println( list)

	fmt.Println( res)
}

//func TestCa_GetCA(t *testing.T) {
//	db := newMockStateDB()
//	addr := syscontracts.CAManagementAddress
//	c := CAManager{stateDB:db, contractAddr:addr}
//	c.setRootCert("/home/night/go/src/github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/cmd/selfCA.PEM")
//	//list, _ := c.getList()
//	//subject := "/C=CN/O=wxbc/CN=test1"
//	var result []string
//	res, _ := c.getAllCert()
//	for _, v := range res{
//
//		pem, _ := v.GetPEM()
//		//println(pem)
//		result = append(result, pem)
//	}
//	//cert, _ :=res.GetPEM()
//	//rootCA, _ := c.getRootCA()
//	//rootCAString, _ := rootCA.GetPEM()
//
//	fmt.Println(result)
//
//}
func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}
func TestWrapperCA(t *testing.T) {
	db := newMockStateDB()
	fnNameInput := "setRootCA"
	params := "/home/night/go/src/github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/cmd/test-CA.PEM"
	c := CAWrapper{NewCAManager(db)}
	var input = MakeInput(fnNameInput, string(params))
	ret, _ := c.Run(input)

	fnNameInput1 := "getRootCA"
	var input1 = MakeInput(fnNameInput1)
	ret2, _ := c.Run(input1)

	log.Println(BytesToInt(ret))
	log.Println(byteutil.BytesToString(ret2))


}