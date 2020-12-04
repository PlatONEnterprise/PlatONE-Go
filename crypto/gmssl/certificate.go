package gmssl

import (
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl/gmssl"
)

type Certificate struct{
	Cert *gmssl.Certificate
}
type CertificateRequest struct{
	req *gmssl.CertificateRequest
}

func CreateCerficate(pkey *PrivateKey,pub *PublicKey, digestAlg string,serialNumber int64, orgnizationName, commonName string) (*Certificate, error){
	cert, err := gmssl.CreateCertificate(pkey.sk, pub.pk, digestAlg,serialNumber, orgnizationName,commonName)
	if err != nil{
		return nil, err
	}
	return &Certificate{Cert: cert}, nil
}

func CreateCertificateForReq(prv *PrivateKey, req *CertificateRequest, ca *Certificate, digestAlg string, serialNumber int64) (*Certificate, error){
	cert, err := gmssl.CreateCertificateForReq(prv.sk,req.req, ca.Cert,digestAlg, serialNumber)
	if err != nil{
		return nil, err
	}
	return &Certificate{Cert: cert}, nil
}

func (cert *Certificate) GetPEM() (string, error) {
	return cert.Cert.GetPEM()
}

func (req *CertificateRequest) GetPEM() (string, error) {
	return req.req.GetPEM()
}

func NewCertificateFromPEM(pem string) (*Certificate, error){
	cert, err := gmssl.NewCertificateFromPEM(pem,"")
	if err != nil{
		return nil, err
	}
	return &Certificate{Cert: cert}, nil
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

func Verify(ca, cert *Certificate) (bool, error) {
	return gmssl.VerifyCertificate(ca.Cert, cert.Cert)
}

