package vm

import (
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"strings"

	"math/big"
	"sort"
)

const (
	keyOfCAList = "CA-list"
	root        = "root"
	revoke      = "revokeList"
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

func (ca *CAManager) isRootCertExists() bool {
	if ca.getState(root) != nil {
		return true
	} else {
		return false
	}
}

func (ca *CAManager) doSetRoot(subject, cert string) {
	ca.setState(root, []byte(subject))
	ca.setState(subject, []byte(cert))
}

func (ca *CAManager) setRootCert(cert string) error {
	//todo string 过长问题
	if ca.isRootCertExists() {
		ca.emitNotifyEventInCA("SetRootCert", RootCertExist, fmt.Sprintf("root cert exists."))
		return errAlreadySetRootCert
	}

	if !hasCaOpPermission(ca.stateDB, ca.caller) {
		return errNoPermission
	}

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
	ca.doSetRoot(subject, cert)

	err = ca.appendToCaList(subject)
	if err != nil {
		ca.emitNotifyEventInCA("SetRootCert", RootCertExist, fmt.Sprintf("root cert exists."))
		return err
	}
	ca.emitNotifyEventInCA("SetRootCert", SetRootCertSuccess, fmt.Sprintf("set root cert success."))

	return nil
}

func (ca *CAManager) addIssuer(cert string) error {

	if !hasCaOpPermission(ca.stateDB, ca.caller) {
		return errNoPermission
	}

	cert = ca.ParseCert(cert)
	issuerCert, err := gmssl.NewCertificateFromPEM(cert)
	if nil != err {
		ca.emitNotifyEventInCA("AddIssuer", NewCertFailure, fmt.Sprintf("new cert fail"))
		return errNewCert
	}
	//todo 提取出两个方法
	issuerSubject, err := issuerCert.Cert.GetSubject()

	if nil != err {
		ca.emitNotifyEventInCA("AddIssuer", GetSubjectFailure, fmt.Sprintf("get subject fail."))
		return errGetSubject
	}

	rootCASub, err := issuerCert.Cert.GetIssuer()
	if nil != err {
		ca.emitNotifyEventInCA("AddIssuer", GetSubjectFailure, fmt.Sprintf("get issuer fail."))
		return errGetSubject
	}

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

	if ca.isIssuerExist(issuerSubject) {

		ca.emitNotifyEventInCA("AddIssuer", IssuerCertExist, fmt.Sprintf("issuer exsits."))
		return errAlreadySetIssuerCert
	}
	err = ca.appendToCaList(issuerSubject)
	if err != nil {

		return err
	}

	ca.setState(issuerSubject, []byte(cert))
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

func (ca *CAManager) AuthJudg(cer *gmssl.Certificate) bool {
	issuer, _ := cer.Cert.GetIssuer()

	if ca.isIssuerExist(issuer) {

		ca.emitNotifyEventInCA("AddIssuer", IssuerCertExist, fmt.Sprintf("issuer exsits."))
		return false
	}
	pk, err := cer.Cert.GetPublicKey()
	if err != nil {
		return false
	}
	pkHex, err := pk.GetHex()
	print(pkHex)
	if err != nil {
		return false

	}
	pkBytes := []byte(pkHex)
	authAddr := common.BytesToAddress(crypto.Keccak256(pkBytes[1:])[12:])
	if authAddr != ca.caller{
		return false
	}
	return true
}

func (ca *CAManager) ifIsConsensusNode(certPem *gmssl.Certificate) bool {
	public, _ := certPem.Cert.GetPublicKey()
	publicHex, _ := public.GetHex()

	err := judgeConsensus(ca.stateDB, publicHex)
	if err != nil {
		return false
	}
	return true
}

func (ca *CAManager) revoke(cert string) error {

	certPem, err := gmssl.NewCertificateFromPEM(cert)
	if err != nil {
		return err
	}
	subject, err := certPem.Cert.GetSubject()
	if err != nil {
		return err
	}
	if ca.isRootCert(subject) {
		return errParamsInvalid
	}
	//权限判断
	if !ca.AuthJudg(certPem) {
		return errNoPermission
	}

	//can not revoke consensus node
	if ca.ifIsConsensusNode(certPem) {
		return errNoPermission
	}
	err = ca.appendToRevokeList(subject)
	if err != nil {
		return err
	}
	return nil
}

func (ca *CAManager) isRevoked(subject string) bool {
	list, err := ca.getList(revoke)

	if err != nil {
		if err != errCANotFound {
			return false
		}
		list = []string{}
	}
	index := sort.SearchStrings(list, subject)

	if index < len(list) && list[index] == subject {
		return true
	}
	return false
}

func (ca *CAManager) getAllCert() ([]string, error) {
	certDocList := make([]string, 0)
	caList, _ := ca.getList(keyOfCAList)
	for _, v := range caList {
		pem := ca.getState(v)
		certDocList = append(certDocList, string(pem))
	}
	return certDocList, nil
}

func (ca *CAManager) getAllCertificate() ([]*gmssl.Certificate, error) {

	caList, _ := ca.getList(keyOfCAList)
	len := len(caList)
	certDocList := make([]*gmssl.Certificate, len)
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
	rootSub := ca.getState(root)
	rootCA := ca.getState(string(rootSub))
	rootCa, err := gmssl.NewCertificateFromPEM(string(rootCA))
	if err != nil {
		return nil, err
	}
	return rootCa, nil
}

func (ca *CAManager) isRootCert(subject string) bool {
	rootSub := ca.getState(root)
	return subject == string(rootSub)
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

func (ca *CAManager) appendToCaList(subject string) error {
	list, err := ca.getList(keyOfCAList)

	if err != nil {
		if err != errCANotFound {
			return err
		}
		list = []string{}
	}
	list = append(list, subject)
	sort.Strings(list)
	encodedCaList, err := rlp.EncodeToBytes(list)
	if err != nil {
		return err
	}
	ca.setState(keyOfCAList, encodedCaList)
	return nil
}

func (ca *CAManager) appendToRevokeList(subject string) error {
	revokeList, err := ca.getList(revoke)
	if err != nil {
		revokeList = []string{}
	}
	revokeList = append(revokeList, subject)
	sort.Strings(revokeList)
	encodedCaList, err := rlp.EncodeToBytes(revokeList)
	if err != nil {
		return err
	}
	ca.setState(revoke, encodedCaList)
	return nil
}

func (ca *CAManager) getList(key string) ([]string, error) {
	bin := ca.getState(key)
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

func (ca *CAManager) isIssuerExist(issuer string) bool {
	list, err := ca.getList(keyOfCAList)

	if err != nil {
		if err != errCANotFound {
			return false
		}
		list = []string{}
	}

	index := sort.SearchStrings(list, issuer)

	if index < len(list) && list[index] == issuer {
		return true
	}
	return false
}

func (ca *CAManager) getPem(sub string) string {
	return string(ca.getState(sub))
}
