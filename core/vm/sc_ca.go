package vm

import (
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"strings"

	"math/big"
	"sort"
)

const (
	keyOfCAList = "CA-list"
)

const (
	SetRootCertSuccess CodeType = 0
	NewCertFailure     CodeType = 1
	GetSubjectFailure  CodeType = 2
	RootCertExist      CodeType = 3
	NoPermission       CodeType = 4
	IssuerCertExist    CodeType = 5
)

var (
	errCANotFound = errors.New("ca not found")
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

func (ca *CAManager) setState(key string, value []byte) {
	ca.stateDB.SetState(ca.contractAddr, []byte(key), value)
}

func (ca *CAManager) getState(key string) []byte {
	return ca.stateDB.GetState(ca.contractAddr, []byte(key))
}

func isPem(cert string) bool {
	return strings.HasSuffix(cert, "PEM")
}

func (ca *CAManager) returnSuccess(topic string) (int32, error) {
	ca.emitEvent(topic, operateSuccess, "Success")
	return int32(operateSuccess), nil
}

func (ca *CAManager) returnFail(topic string, err error) (int32, error) {
	ca.emitEvent(topic, operateFail, err.Error())
	returnErr := err
	// todo: in some cases, returnErr = nil
	return int32(operateFail), returnErr
}

func (ca *CAManager) setRootCert(cert string) error {

	//if !isPem(cert){
	//	return errParamsInvalid
	//}
	//
	cert = ca.ParseCert(cert)

	rootCert, err := gmssl.NewCertificateFromPEM(cert)
	if nil != err {
		ca.emitNotifyEventInCA("SetRootCert", NewCertFailure, fmt.Sprintf("new cert fail"))
		return errNewCert
	}
	subject, err := rootCert.Cert.GetSubject()
	if nil != err {
		ca.emitNotifyEventInCA("SetRootCert", GetSubjectFailure, fmt.Sprintf("get subject fail."))
		return errGetSubject
	}
	root := "root"
	if ca.getState(root) != nil {
		ca.emitNotifyEventInCA("SetRootCert", RootCertExist, fmt.Sprintf("root cert exists."))

		return errAlreadySetRootCert
	}

	ca.setState(root, []byte(subject))
	ca.setState(subject, []byte(cert))

	var caList []string
	err = ca.appendToCaList(caList, subject)
	if err != nil {
		ca.emitNotifyEventInCA("SetRootCert", RootCertExist, fmt.Sprintf("root cert exists."))
		return err
	}
	ca.emitNotifyEventInCA("SetRootCert", SetRootCertSuccess, fmt.Sprintf("set root cert success."))

	return nil
}

func (ca *CAManager) addIssuer(cert string) error {
	cert = ca.ParseCert(cert)
	issuerCert, err := gmssl.NewCertificateFromPEM(cert)
	if nil != err {
		ca.emitNotifyEventInCA("AddIssuer", NewCertFailure, fmt.Sprintf("new cert fail"))
		return errNewCert
	}

	issuer, err := issuerCert.Cert.GetSubject()
	if nil != err {
		ca.emitNotifyEventInCA("AddIssuer", GetSubjectFailure, fmt.Sprintf("get subject fail."))
		return errGetSubject
	}

	rootCASub, err := issuerCert.Cert.GetIssuer()
	if nil != err {
		ca.emitNotifyEventInCA("AddIssuer", GetSubjectFailure, fmt.Sprintf("get issuer fail."))
		return errGetSubject
	}

	root := "root"

	rootSub := string(ca.getState(root))
	if rootSub != rootCASub {
		ca.emitNotifyEventInCA("AddIssuer", NoPermission, fmt.Sprintf("root cert wrong."))
		return errNoPermission
	}

	rootPem := ca.getPem(rootSub)
	rootCA, err := gmssl.NewCertificateFromPEM(rootPem)

	if err != nil {
		ca.emitNotifyEventInCA("AddIssuer", NewCertFailure, fmt.Sprintf("new root cert fail"))
		return errNewCert
	}

	result, _ := gmssl.Verify(rootCA, issuerCert)

	if !result {
		ca.emitNotifyEventInCA("AddIssuer", NoPermission, fmt.Sprintf("root cert verify fail."))
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

		ca.emitNotifyEventInCA("AddIssuer", IssuerCertExist, fmt.Sprintf("issuer exsits."))
		return errAlreadySetIssuerCert
	}
	err = ca.appendToCaList(list, issuer)
	if err != nil {

		return err
	}

	ca.setState(issuer, []byte(cert))
	ca.emitNotifyEventInCA("AddIssuer", SetRootCertSuccess, fmt.Sprintf("set issuer cert success."))

	return nil
}

func (ca *CAManager) getCert(subject string) (*gmssl.Certificate, error) {

	caPem := ca.getState(subject)
	caDoc, err := gmssl.NewCertificateFromPEM(string(caPem))
	if err != nil {
		return nil, err
	}
	return caDoc, nil
}

func (ca *CAManager) getAllCert() ([]string, error) {
	certDocList := make([]string, 0)
	caList, _ := ca.getList()
	for _, v := range caList {
		pem := ca.getState(v)
		certDocList = append(certDocList, string(pem))
	}
	return certDocList, nil
}

func (ca *CAManager) getAllCertificate() ([]*gmssl.Certificate, error) {
	certDocList := make([]*gmssl.Certificate, 0)
	caList, _ := ca.getList()
	for _, v := range caList {
		pem := ca.getState(v)
		ca, err := gmssl.NewCertificateFromPEM(string(pem))
		if err != nil {
			return nil, err
		}
		certDocList = append(certDocList, ca)
	}
	return certDocList, nil
}

func (ca *CAManager) getRootCert() (*gmssl.Certificate, error) {
	rootSub := ca.getState("root")
	rootCA := ca.getState(string(rootSub))
	rootCa, err := gmssl.NewCertificateFromPEM(string(rootCA))
	if err != nil {
		return nil, err
	}
	return rootCa, nil
}

func (ca *CAManager) emitNotifyEventInCA(topic string, code CodeType, msg string) {
	emitEvent(ca.contractAddr, ca.stateDB, ca.blockNumber.Uint64(), topic, code, msg)
}

func (ca *CAManager) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(ca.contractAddr, ca.stateDB, ca.blockNumber.Uint64(), topic, code, msg)
}

func (ca *CAManager) ParseCert(cert string) string {
	oldBegin := "BEGIN"
	newBegin := "BEGIN "
	oldEnd := "END"
	newEnd := "END "
	cert = strings.Replace(cert, oldBegin, newBegin, -1)
	cert = strings.Replace(cert, oldEnd, newEnd, -1)
	return cert
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

func (ca *CAManager) getList() ([]string, error) {
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

func (ca *CAManager) isIssuerExist(list []string, issuer string) bool {
	index := sort.SearchStrings(list, issuer)
	//not found
	if index < len(list) && list[index] == issuer {
		return true
	}

	return false
}

func (ca *CAManager) getPem(sub string) string {
	return string(ca.getState(sub))
}
