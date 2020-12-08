package vm

import (
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"log"

	"math/big"
	"sort"
	"strings"
)

const(
	keyOfCAList = "CA-list"
)

var(
	errCANotFound   = errors.New("ca not found")
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

func (ca *CAManager) getState(key string) []byte{
	return ca.stateDB.GetState(ca.contractAddr, []byte(key))
}

func isPem(cert string) bool{
	return strings.HasSuffix(cert, "PEM")
}

func (ca *CAManager) setRootCert(cert string) error{

	if !isPem(cert){
		return errParamsInvalid
	}

	pem := readFromFile(cert)
	rootCert, err := gmssl.NewCertificateFromPEM(pem)
	if nil != err {
		//todo add event and new error type
		return errParamsInvalid
	}
	subject, err := rootCert.Cert.GetSubject()
	root := "root"
	if ca.getState(root) != nil {
		//todo add event and new error type
		return errParamsInvalid
	}

	ca.setState(root, []byte(subject))
	ca.setState(subject, []byte(pem))

	var caList []string
	err = ca.appendToCaList(caList, subject)
	if err != nil {
		return err
	}
	return nil
}

func (ca *CAManager) appendToCaList(caList []string, subject string) error {

	caList = append(caList, subject)
	sort.Strings(caList)
	encodedCaList, err := rlp.EncodeToBytes(caList)
	if err != nil {
		return err
	}
	ca.setState(keyOfCAList, encodedCaList)
	return nil
}

func (ca *CAManager) getList() ([]string, error ) {
	bin := ca.getState(keyOfCAList)
	if len(bin) == 0 {
		return nil, errCANotFound
	}

	var list []string
	err := rlp.DecodeBytes(bin, &list)
	if err != nil {
		return nil, err
	}
	return list, err
}

func (ca *CAManager) addIssuer(cert string) error{
	if !isPem(cert){
		return errParamsInvalid
	}
	pem := readFromFile(cert)
	issuerCert, err := gmssl.NewCertificateFromPEM(pem)
	if nil != err {
		//todo add event and new error type
		return errParamsInvalid
	}

	issuer, err := issuerCert.Cert.GetSubject()
	if err != nil {
		panic(err)
	}

	rootCASub, err := issuerCert.Cert.GetIssuer()
	if err != nil {
		panic(err)
	}

	root := "root"

	rootSub := string(ca.getState(root))
	if rootSub != rootCASub {
		//todo add event
		return errNoPermission
	}

	rootPem := ca.getPem(rootSub)
	rootCA, err := gmssl.NewCertificateFromPEM(rootPem)

	if err != nil {
		panic(err)
	}

	result, err := gmssl.Verify(rootCA, issuerCert)
	if err != nil {
		panic(err)
	}
	if !result {
		//todo add event
		return errNoPermission
	}

	list, err := ca.getList()

	if err != nil {
		if err != errCANotFound {
			return err
		}
		list = []string{}
	}

	if ca.isIssuerExist(list, issuer) {
		return errors.New("issuer exist")
	}
	err = ca.appendToCaList(list, issuer)
	if err != nil {
		return err
	}

	ca.setState(issuer, []byte(pem))

	return nil

}

func (ca *CAManager) isIssuerExist(list []string, issuer string) bool {
	index := sort.SearchStrings(list, issuer)
	//not found
	if index < len(list) && list[index] == issuer {
		return true
	}

	return false
}

func (ca *CAManager) getPem(sub string) string{
	return string(ca.getState(sub))
}

func (ca *CAManager) getCert(subject string) (*gmssl.Certificate, error){

	caPem := ca.getState(subject)
	caDoc, err := gmssl.NewCertificateFromPEM(string(caPem))
	if err != nil {
		return nil, err
	}
	return caDoc, nil
}

func (ca *CAManager) getAllCert()  ([]string, error){
	certDocList := make([]string, 0)
	caList, _ := ca.getList()
	for _, v := range caList{
		pem := ca.getState(v)
		certDocList = append(certDocList, string(pem))
	}
	return certDocList, nil
}

func (ca *CAManager) getAllCertificate()  ([]*gmssl.Certificate, error){
	certDocList := make([]*gmssl.Certificate, 0)
	caList, _ := ca.getList()
	for _, v := range caList{
		pem := ca.getState(v)
		ca, err := gmssl.NewCertificateFromPEM(string(pem))
		if err != nil{
			return nil, err
		}
		certDocList = append(certDocList, ca)
	}
	return certDocList, nil
}


func (ca *CAManager) getRootCert() (*gmssl.Certificate, error){
	log.Println("123")
	rootSub := ca.getState("root")
	rootCA := ca.getState(string(rootSub))
	rootCa, err := gmssl.NewCertificateFromPEM(string(rootCA))
	if err != nil{
		return nil, err
	}
	return rootCa, nil
}
