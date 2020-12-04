package vm

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

const (
	success = 0
	fail = -1
)
type CAWrapper struct {
	base *CAManager
}
func newCAWrapper(db StateDB) *CAWrapper {
	return &CAWrapper{NewCAManager(db)}
}

func (ca *CAWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.CAManagerGas
}

func (ca *CAWrapper) Run(input []byte) ([]byte, error) {
	defer func() {
		if e := recover(); e != nil {
			err := fmt.Errorf("[CA] running error: %+v", e.(string))
			log.Error("[CA] running", "error", err)
		}
	}()

	fnName, ret, err := execSC(input, ca.AllExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
	}

	return ret, nil
}

func (ca *CAWrapper) AllExportFns() SCExportFns {
	return SCExportFns{
		"setRootCA":             ca.setRootCA,
		"addIssuer":             ca.addIssuer,
		"getCA":                 ca.getCA,
		"getAllCA":              ca.getAllCA,
		"getRootCA":             ca.getRootCA,
		//"getAllCertificate":     ca.getAllCertificate,
	}
}

func (ca *CAWrapper) setRootCA(cert string)  (int32, error){
	err := ca.base.setRootCA(cert)
	if nil != err {
		return fail, err
	}
	return success, nil
}


func (ca *CAWrapper) addIssuer(cert string)  (int32, error){
	err := ca.base.addIssuer(cert)
	if nil != err {
		return fail, err
	}
	return success, nil
}

func (ca *CAWrapper) getCA(subject string)  (string, error){
	caStruct, err := ca.base.getCA(subject)
	if nil != err {
		return "", err
	}
	res, err := caStruct.GetPEM()
	if nil != err {
		return "", err
	}
	return res, nil

}

func (ca *CAWrapper) getAllCA()  (string, error){
	caList, err := ca.base.getAllCA()
	if nil != err {
		return newInternalErrorResult(err).String(), err
	}
	return newSuccessResult(caList).String(), nil

}

func (ca *CAWrapper) getAllCertificate() ([]*gmssl.Certifacate, error){
	caList, err := ca.base.getAllCA()
	if nil != err {
		return nil, err
	}
	return caList, nil
}

func (ca *CAWrapper) getRootCA()  (string, error){
	rootCa, err := ca.base.getRootCA()
	if nil != err {
		return "", err
	}
	res, _ := rootCa.GetPEM()
	return res, nil

}