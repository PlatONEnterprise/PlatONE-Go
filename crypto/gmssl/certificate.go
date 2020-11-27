package gmssl

import (
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl/gmssl"
)

type Certifacate struct{
	cert *gmssl.Certificate
}
type CertificateRequest struct{
	req *gmssl.CertificateRequest
}

func CreateCerficate(pkey *PrivateKey,pub *PublicKey, digestAlg string,serialNumber int64, orgnizationName, commonName string) (*Certifacate, error){
	cert, err := gmssl.CreateCertificate(pkey.sk, pub.pk, digestAlg,serialNumber, orgnizationName,commonName)
	if err != nil{
		return nil, err
	}
	return &Certifacate{cert:cert}, nil
}

func CreateCertificateForReq(prv *PrivateKey, req *CertificateRequest, ca *Certifacate, digestAlg string, serialNumber int64) (*Certifacate, error){
	cert, err := gmssl.CreateCertificateForReq(prv.sk,req.req, ca.cert,digestAlg, serialNumber)
	if err != nil{
		return nil, err
	}
	return &Certifacate{cert:cert}, nil
}

func (cert *Certifacate) GetPEM() (string, error) {
	return cert.cert.GetPEM()
}

func (req *CertificateRequest) GetPEM() (string, error) {
	return req.req.GetPEM()
}

func NewCertificateFromPEM(pem string) (*Certifacate, error){
	cert, err := gmssl.NewCertificateFromPEM(pem,"")
	if err != nil{
		return nil, err
	}
	return &Certifacate{cert:cert}, nil
}

func CreateCertRequest(prv *PrivateKey, digestAlg string,orgnizationName, commonName string) (*CertificateRequest, error){
	req, err := gmssl.CreateCertRequest(prv.sk, digestAlg,orgnizationName,commonName)
	if err != nil{
		return nil, err
	}

	return &CertificateRequest{req}, nil
}

func NewCertRequestFromPEM(pem string) (*CertificateRequest, error){
	req, err :=  gmssl.NewCertRequestFromPEM(pem)
	if err != nil{
		return nil, err
	}
	return &CertificateRequest{req}, nil
}

func Verify(ca, cert *Certifacate) (bool, error) {
	return gmssl.VerifyCertificate(ca.cert, cert.cert)
}

