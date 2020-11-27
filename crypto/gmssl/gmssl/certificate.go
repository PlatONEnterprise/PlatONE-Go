/*
 * Copyright (c) 2017 - 2018 The GmSSL Project.  All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in
 *    the documentation and/or other materials provided with the
 *    distribution.
 *
 * 3. All advertising materials mentioning features or use of this
 *    software must display the following acknowledgment:
 *    "This product includes software developed by the GmSSL Project.
 *    (http://gmssl.org/)"
 *
 * 4. The name "GmSSL Project" must not be used to endorse or promote
 *    products derived from this software without prior written
 *    permission. For written permission, please contact
 *    guanzhi1980@gmail.com.
 *
 * 5. Products derived from this software may not be called "GmSSL"
 *    nor may "GmSSL" appear in their names without prior written
 *    permission of the GmSSL Project.
 *
 * 6. Redistributions of any form whatsoever must retain the following
 *    acknowledgment:
 *    "This product includes software developed by the GmSSL Project
 *    (http://gmssl.org/)"
 *
 * THIS SOFTWARE IS PROVIDED BY THE GmSSL PROJECT ``AS IS'' AND ANY
 * EXPRESSED OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED.  IN NO EVENT SHALL THE GmSSL PROJECT OR
 * ITS CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
 * NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
 * STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
 * OF THE POSSIBILITY OF SUCH DAMAGE.
 */

/* +build cgo */
package gmssl

/*
#cgo  LDFLAGS: -Llib -lcrypto -lssl

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <openssl/ec.h>
#include <openssl/sm2.h>
#include <openssl/bio.h>
#include <openssl/evp.h>
#include <openssl/pem.h>
#include <openssl/engine.h>
#include <openssl/objects.h>
#include <openssl/opensslconf.h>
#include <openssl/x509v3.h>

extern long _BIO_get_mem_data(BIO *b, char **pp);
extern void _OPENSSL_free(void *addr);
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

type Certificate struct {
	x509 *C.X509
}

type CertificateRequest struct {
	req *C.X509_REQ
}

func (req * CertificateRequest) VerifySig() bool {
	var pub *C.EVP_PKEY
	pub = C.X509_REQ_get0_pubkey(req.req)

	i := C.X509_REQ_verify(req.req, pub)

	if i <=  0{
		return false
	}
	return true
}

func (req * CertificateRequest) GetPEM() (string, error) {
	bio := C.BIO_new(C.BIO_s_mem())
	if bio == nil {
		return "", GetErrors()
	}
	defer C.BIO_free(bio)

	if 1 != C.PEM_write_bio_X509_REQ(bio,req.req) {
		return  "", GetErrors()
	}

	var p *C.char
	len := C._BIO_get_mem_data(bio, &p)
	if len <= 0 {
		return "", GetErrors()
	}
	return C.GoString(p)[:len], nil
}

func (req * CertificateRequest) GetPublicKey() (*PublicKey, error) {
	var pub *C.EVP_PKEY
	pub = C.X509_REQ_get0_pubkey(req.req);

	return &PublicKey{pub}, nil
}

func NewCertRequestFromPEM(pem string) (*CertificateRequest, error){
	cpem := C.CString(pem)
	defer C.free(unsafe.Pointer(cpem))

	bio := C.BIO_new_mem_buf(unsafe.Pointer(cpem), -1)
	if bio == nil {
		return nil, GetErrors()
	}
	defer C.BIO_free(bio)

	req := C.PEM_read_bio_X509_REQ(bio, nil, nil, nil)
	if req == nil {
		return nil, GetErrors()
	}
	ret := &CertificateRequest{req}
	runtime.SetFinalizer(ret, func(ret *CertificateRequest) {
		C.X509_REQ_free(ret.req)
	})

	return ret, nil
}
func CreateCertRequest(prv *PrivateKey, digestAlg string,orgnizationName, commonName string) (*CertificateRequest, error) {
	var (
		req *C.X509_REQ
	)

	req = C.X509_REQ_new();
	if req == nil{
		return nil, GetErrors()
	}

	C.X509_REQ_set_pubkey(req, prv.pkey);

	var name *C.X509_NAME
	name = C.X509_REQ_get_subject_name(req);

	C.X509_NAME_add_entry_by_txt(name,
		C.CString("C"),
		C.MBSTRING_ASC,
		(*C.uchar)(unsafe.Pointer(C.CString("CN"))),
		-1, -1, 0);
	C.X509_NAME_add_entry_by_txt(name,
		C.CString("O"),
		C.MBSTRING_ASC,
		(*C.uchar)(unsafe.Pointer(C.CString(orgnizationName))),
		-1, -1, 0);
	C.X509_NAME_add_entry_by_txt(name,
		C.CString("CN"),
		C.MBSTRING_ASC,
		(*C.uchar)(unsafe.Pointer(C.CString(commonName))),
		-1, -1, 0);

	C.X509_REQ_set_subject_name(req, name);

	var md *C.EVP_MD
	switch digestAlg {
	case "sm3": md = C.EVP_sm3();
	case "SM3": md = C.EVP_sm3();
	case "sha256": md = C.EVP_sha256();
	case "SHA256": md = C.EVP_sha256();
	}

	C.X509_REQ_sign(req, prv.pkey, md);

	return &CertificateRequest{req:req}, nil
}
func CreateCertificateForReq(prv *PrivateKey, req *CertificateRequest, ca *Certificate, digestAlg string, serialNumber int64) (*Certificate, error) {
	var (
		x509ss *C.X509
	)
	pkey := prv.pkey

	x509ss = C.X509_new()
	if x509ss == nil {
		return nil, errors.New("x509_new error")
	}

	C.ASN1_INTEGER_set(C.X509_get_serialNumber(x509ss), C.long(serialNumber));
	C.X509_gmtime_adj(C.X509_get_notBefore(x509ss), 0); // 0: now
	C.X509_gmtime_adj(C.X509_get_notAfter(x509ss), 365*10*24*3600); // 10 years from now

	var pub *C.EVP_PKEY
	pub = C.X509_REQ_get0_pubkey(req.req);
	i := C.X509_REQ_verify(req.req, pub)
	if i <= 0{
		return nil, GetErrors()
	}

	C.X509_set_pubkey(x509ss, pub);

	var name *C.X509_NAME
	name = C.X509_REQ_get_subject_name(req.req)
	C.X509_set_subject_name(x509ss, name)

	caName :=  C.X509_get_subject_name(ca.x509);
	C.X509_set_issuer_name(x509ss, caName);

	var md *C.EVP_MD
	switch digestAlg {
		case "sm3": md = C.EVP_sm3();
		case "SM3": md = C.EVP_sm3();
		case "sha256": md = C.EVP_sha256();
		case "SHA256": md = C.EVP_sha256();
	}
	C.X509_sign(x509ss, pkey, md);

	return &Certificate{x509:x509ss}, nil
}

func CreateCertificate(prv *PrivateKey, pub *PublicKey, digestAlg string,serialNumber int64, orgnizationName, commonName string) (*Certificate, error) {
	var (
		x509ss *C.X509
	)
	pkey := prv.pkey

	x509ss = C.X509_new()
	if x509ss == nil {
		return nil, errors.New("x509_new error")
	}

	C.ASN1_INTEGER_set(C.X509_get_serialNumber(x509ss), C.long(serialNumber));
	C.X509_gmtime_adj(C.X509_get_notBefore(x509ss), 0); // 0: now
	C.X509_gmtime_adj(C.X509_get_notAfter(x509ss), 365*10*24*3600); // 10 years from now

	C.X509_set_pubkey(x509ss, pub.pkey);

	var name *C.X509_NAME
	name = C.X509_get_subject_name(x509ss);

	C.X509_NAME_add_entry_by_txt(name,
									C.CString("C"),
									C.MBSTRING_ASC,
									(*C.uchar)(unsafe.Pointer(C.CString("CN"))),
									-1, -1, 0);
	C.X509_NAME_add_entry_by_txt(name,
								C.CString("O"),
								C.MBSTRING_ASC,
								(*C.uchar)(unsafe.Pointer(C.CString(orgnizationName))),
								-1, -1, 0);
	C.X509_NAME_add_entry_by_txt(name,
								C.CString("CN"),
								C.MBSTRING_ASC,
								(*C.uchar)(unsafe.Pointer(C.CString(commonName))),
								-1, -1, 0);

	C.X509_set_issuer_name(x509ss, name);

	// TODO: specify hash function, EVP_sha256 or EVP_sm3

	var md *C.EVP_MD
	switch digestAlg {
		case "sm3": md = C.EVP_sm3();
		case "SM3": md = C.EVP_sm3();
		case "sha256": md = C.EVP_sha256();
		case "SHA256": md = C.EVP_sha256();
	}
	C.X509_sign(x509ss, pkey, md);

	return &Certificate{x509:x509ss}, nil
}

func NewCertificateFromPEM(pem string, pass string) (*Certificate, error) {
	cpem := C.CString(pem)
	defer C.free(unsafe.Pointer(cpem))
	//cpass := C.CString(pass)
	//defer C.free(unsafe.Pointer(cpass))

	bio := C.BIO_new_mem_buf(unsafe.Pointer(cpem), -1)
	if bio == nil {
		return nil, GetErrors()
	}
	defer C.BIO_free(bio)

	//x509 := C.PEM_read_bio_X509(bio, nil, nil, unsafe.Pointer(cpass))
	x509 := C.PEM_read_bio_X509(bio, nil, nil, nil)
	if x509 == nil {
		return nil, GetErrors()
	}
	ret := &Certificate{x509}
	runtime.SetFinalizer(ret, func(ret *Certificate) {
		C.X509_free(ret.x509)
	})

	return ret, nil
}

func (cert *Certificate) GetPEM() (string, error) {
	bio := C.BIO_new(C.BIO_s_mem())
	if bio == nil {
		return "", GetErrors()
	}
	defer C.BIO_free(bio)

	if 1 != C.PEM_write_bio_X509(bio,cert.x509) {
		return  "", GetErrors()
	}

	var p *C.char
	len := C._BIO_get_mem_data(bio, &p)
	if len <= 0 {
		return "", GetErrors()
	}
	return C.GoString(p)[:len], nil
}

func (cert *Certificate) GetSubject() (string, error) {
	name := C.X509_get_subject_name(cert.x509)
	if name == nil {
		return "", GetErrors()
	}
	str := C.X509_NAME_oneline(name, nil, 0)
	if str == nil {
		return "", GetErrors()
	}
	defer C._OPENSSL_free(unsafe.Pointer(str))
	return C.GoString(str), nil
}

func (cert *Certificate) GetIssuer() (string, error) {
	name := C.X509_get_issuer_name(cert.x509)
	if name == nil {
		return "", GetErrors()
	}
	str := C.X509_NAME_oneline(name, nil, 0)
	if str == nil {
		return "", GetErrors()
	}
	defer C._OPENSSL_free(unsafe.Pointer(str))
	return C.GoString(str), nil
}

func (cert *Certificate) GetSerialNumber() (string, error) {
	serial := C.X509_get0_serialNumber(cert.x509)
	if serial == nil {
		return "", GetErrors()
	}
	bn := C.ASN1_INTEGER_to_BN(serial, nil)
	if bn == nil {
		return "", GetErrors()
	}
	defer C.BN_free(bn)
	hex := C.BN_bn2hex(bn)
	if hex == nil {
		return "", GetErrors()
	}
	defer C._OPENSSL_free(unsafe.Pointer(hex))
	return C.GoString(hex), nil
}

func (cert *Certificate) GetPublicKey() (*PublicKey, error) {
	pkey := C.X509_get_pubkey(cert.x509)
	if pkey == nil {
		return nil, GetErrors()
	}
	ret := &PublicKey{pkey}
	runtime.SetFinalizer(ret, func(ret *PublicKey) {
		C.EVP_PKEY_free(ret.pkey)
	})
	return ret, nil
}

func (cert *Certificate) GetText() (string, error) {
	bio := C.BIO_new(C.BIO_s_mem())
	if bio == nil {
		return "", GetErrors()
	}
	defer C.BIO_free(bio)
	if 0 >= C.X509_print(bio, cert.x509) {
		return "", GetErrors()
	}
	var p *C.char
	len := C._BIO_get_mem_data(bio, &p)
	if len <= 0 {
		return "", GetErrors()
	}
	return C.GoString(p)[:len], nil
}

func (cert *Certificate) CheckPrivateKey(skey *PrivateKey) error {
	if 1 != C.X509_check_private_key(cert.x509, skey.pkey) {
		err := GetErrors()
		if err == nil {
			return fmt.Errorf("failure")
		}
		return err
	}
	return nil
}


// FIXME: 验证证书本应使用X509_verify_cert()，但是此方法在cgo调用时会出现dso报错，因此使用对比证书验证签名的方式验证
func VerifyCertificate(parent, cert *Certificate) (bool, error){

	// TODO: 时间周期

	// compare subject
	subj0, err := parent.GetSubject()
	if err != nil{
		return false, err
	}
	subj1, err := cert.GetIssuer()
	if err != nil{
		return false, err
	}
	if subj0 != subj1 {
		return false, nil
	}

	// verify signature
	pk := C.X509_get_pubkey(parent.x509)
	i := C.X509_verify(cert.x509, pk)
	if i != 1 {
		return false, GetErrors()
	}

	return true, nil
}
