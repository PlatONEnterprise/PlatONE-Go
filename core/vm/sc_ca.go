package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"math/big"
	"strings"
)

type CAManager struct {
	stateDB      StateDB
	contractAddr common.Address
	caller       common.Address
	blockNumber  *big.Int
}

func NewCAManager(db StateDB) *CAManager {
	return &CAManager{stateDB: db, contractAddr: syscontracts.CAManagementAddress, blockNumber: big.NewInt(0)}
}

func (ca *CAManager) setState(key string, value []byte){
	ca.stateDB.SetState(ca.contractAddr, []byte(key), value)

}
func (ca *CAManager) setRootCA(cert string) error{
	if !strings.HasSuffix(cert, "PEM") {
		//todo add event
		return errParamsInvalid
	}
	pem := readFromFile(cert)
	rootCert, err := gmssl.NewCertificateFromPEM(pem)
	if nil != err {
		//todo add event and new error type
		return errParamsInvalid
	}
	subject, err := rootCert.Cert.GetSubject()
	//var bin []byte
	ca.setState(subject, []byte(pem))
	return nil
}


func (ca *CAManager) addIssuer(cert string) error{
	return nil

}

func (ca *CAManager) getCA(commonName string)  (string, error){
	return "", nil

}

func (ca *CAManager) getAllCA()  ([]*gmssl.Certifacate, error){
	certList := make([]*gmssl.Certifacate, 0)
	return certList, nil
}

func (ca *CAManager) getRootCA()  (string, error){
	return "", nil

}
