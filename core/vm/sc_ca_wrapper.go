package vm

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

const (
	success = 0
	fail = 1
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
	return params.CnsManagerGas
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
		//ca.base.emitEvent(fnName, operateFail, err.Error())
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
	return 0, nil
}

func (ca *CAWrapper) getCA(commonName string)  (string, error){
	castring, err := ca.base.getCA(commonName)
	if nil != err {
		return "", err
	}
	return castring, nil

}

func (ca *CAWrapper) getAllCA()  (string, error){
	caList, err := ca.base.getAllCA()
	if nil != err {
		return "", err
	}
	return newSuccessResult(caList).String(), nil

}

func (ca *CAWrapper) getRootCA()  (string, error){
	return "", nil

}