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
	fnNameInput := "setRootCert"
	file := "/home/night/go/src/github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/cmd/selfCA.PEM"
	params := readFromFile(file)
	println(params)
	c := CAWrapper{NewCAManager(db)}
	var input = MakeInput(fnNameInput, string("-----BEGINCERTIFICATE-----\nMIIBOjCB4gIBDDAKBggqgRzPVQGDdTAqMQswCQYDVQQGEwJDTjENMAsGA1UECgwE\nd3hiYzEMMAoGA1UEAwwDbGpqMB4XDTIwMTIwNzA5NDYwNloXDTMwMTIwNTA5NDYw\nNlowKjELMAkGA1UEBhMCQ04xDTALBgNVBAoMBHd4YmMxDDAKBgNVBAMMA2xqajBZ\nMBMGByqGSM49AgEGCCqBHM9VAYItA0IABJcXZ56ANChuVwLCQk+YEYBba27xqYLc\nZtLDJ97qOSjYZ8AsgctKTJTWuQ6WoUvWPvjJnHcBECMpGgR8aVVC8BIwCgYIKoEc\nz1UBg3UDRwAwRAIgQMhAPP/GyYiH2xterLDsctwXHmvMckIBooVTuRvUoAkCIEp6\nmuGGjyK22fFqmbhaDSyql5/7/cSoaNJCTHl+EqZM\n-----ENDCERTIFICATE-----\n"))
	ret, _ := c.Run(input)

	fnNameInput1 := "getAllCert"
	//params1 := "/C=CN/O=wxbc/CN=test1"

	var input1 = MakeInput(fnNameInput1)
	ret2, _ := c.Run(input1)

	log.Println(BytesToInt(ret))
	log.Println(byteutil.BytesToString(ret2))


}
//-----BEGIN CERTIFICATE-----\nMIIBOjCB4gIBDDAKBggqgRzPVQGDdTAqMQswCQYDVQQGEwJDTjENMAsGA1UECgwE\nd3hiYzEMMAoGA1UEAwwDbGpqMB4XDTIwMTIwNzA5NDYwNloXDTMwMTIwNTA5NDYw\nNlowKjELMAkGA1UEBhMCQ04xDTALBgNVBAoMBHd4YmMxDDAKBgNVBAMMA2xqajBZ\nMBMGByqGSM49AgEGCCqBHM9VAYItA0IABJcXZ56ANChuVwLCQk+YEYBba27xqYLc\nZtLDJ97qOSjYZ8AsgctKTJTWuQ6WoUvWPvjJnHcBECMpGgR8aVVC8BIwCgYIKoEc\nz1UBg3UDRwAwRAIgQMhAPP/GyYiH2xterLDsctwXHmvMckIBooVTuRvUoAkCIEp6\nmuGGjyK22fFqmbhaDSyql5/7/cSoaNJCTHl+EqZM\n-----END CERTIFICATE-----\n
//-----BEGIN CERTIFICATE-----\nMIIBQDCB5gIBATAKBggqhkjOPQQDAjAsMQswCQYDVQQGEwJDTjENMAsGA1UECgwE\nd3hiYzEOMAwGA1UEAwwFdGVzdDEwHhcNMjAxMjAyMDIxNjM4WhcNMzAxMTMwMDIx\nNjM4WjAsMQswCQYDVQQGEwJDTjENMAsGA1UECgwEd3hiYzEOMAwGA1UEAwwFdGVz\ndDEwWTATBgcqhkjOPQIBBggqgRzPVQGCLQNCAAQuyGAzLdi7JYixfAPS7zbIk+qS\nTZZnKXTkRh3Av1o4XhydkrtEitT2aNYqVVhgSlS4kNPK2bKkE1MZ++p+SZQqMAoG\nCCqGSM49BAMCA0kAMEYCIQCAkRDgsUeoiaqy1t8jHbmst3BzmMWItc6n4eQCgr0Y\njgIhAOYueVj8HE6QsbezDhRxpBPz1qYCxAUdvYvwXeEjV/ud\n-----END CERTIFICATE-----\n
