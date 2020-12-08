package vm

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
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
		"setRootCert":             ca.setRootCert,
		"addIssuer":               ca.addIssuer,
		"getCert":                 ca.getCert,
		"getAllCert":              ca.getAllCert,
		"getRootCert":             ca.getRootCert,
		}
}

func (ca *CAWrapper) setRootCert(cert string)  (int32, error){
	err := ca.base.setRootCert(cert)
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

func (ca *CAWrapper) getCert(subject string)  (string, error){
	caStruct, err := ca.base.getCert(subject)
	if nil != err {
		return "", err
	}
	res, err := caStruct.GetPEM()
	if nil != err {
		return "", err
	}
	return res, nil

}

func (ca *CAWrapper) getAllCert() (string, error){
	caList, err := ca.base.getAllCert()
	if nil != err {
		return newInternalErrorResult(err).String(), err
	}
	return newSuccessResult(caList).String(), nil
}

//func (ca *CAWrapper) getAllCertificate() ([]*gmssl.Certificate, error){
//	caList, err := ca.base.getAllCert()
//	if nil != err {
//		return nil, err
//	}
//	return caList, nil
//}
func (ca *CAWrapper) test() (string, error){
	return "test", nil
}

func (ca *CAWrapper) getRootCert()  (string, error){
	rootCa, err := ca.base.getRootCert()
	if nil != err {
		return "", err
	}
	res, _ := rootCa.GetPEM()
	return newSuccessResult(res).String(), nil

	//return res, nil

}