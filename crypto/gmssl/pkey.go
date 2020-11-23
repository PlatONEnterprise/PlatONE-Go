package gmssl

import (
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl/gmssl"
)

var keyGenArgs =  map[string][][2]string{
	"SM2":{
		{"ec_paramgen_curve", "sm2p256v1"},
		{"ec_param_enc", "named_curve"},
	},
	"sm2":{
		{"ec_paramgen_curve", "sm2p256v1"},
		{"ec_param_enc", "named_curve"},
	},
	"secp256k1":{
		{"ec_paramgen_curve", "secp256k1"},
		{"ec_param_enc", "named_curve"},
	},
	//"secp256r1":{
	//	{"ec_paramgen_curve", "secp256r1"},
	//	{"ec_param_enc", "named_curve"},
	//},
}


type PrivateKey struct{
	sk *gmssl.PrivateKey
}

type PublicKey struct {
	pk *gmssl.PublicKey
}

func (prv *PrivateKey) GetPublicKey() *PublicKey{
	return &PublicKey{pk: prv.sk.GetPublicKey()}
}

func (prv *PrivateKey) GetPEM() (string, error){
	return prv.sk.GetPEM("", "")
}

func (prv *PrivateKey) GetText() (string, error){
	return prv.sk.GetText()
}

func (prv *PrivateKey) GetPublicKeyPEM() (string, error){
	return prv.sk.GetPublicKeyPEM()
}

func (prv *PrivateKey) GetPublicKeyHex() (string, error) {
	return prv.sk.GetECPublicKeyHex()
}

func (prv *PrivateKey)GetPrivateKeyHex() (string, error) {
	return prv.sk.GetECPrivateKeyHex()
}

func (pub *PublicKey) GetPublicKeyPEM() (string, error){
	return pub.pk.GetPEM()
}

func (pub *PublicKey) GetPublicKeyHex()(string, error){
	return pub.pk.GetHex()
}

func (pub *PublicKey) GetText() (string, error){
	return pub.pk.GetText()
}

func NewPublicKeyFromPEM(pem string) (*PublicKey, error) {
	pub, err := gmssl.NewPublicKeyFromPEM(pem)
	if err != nil{
		return nil,err
	}
	return &PublicKey{pk: pub}, nil
}

func NewPrivateKeyFromPEM(pem string) (*PrivateKey, error) {
	pk, err := gmssl.NewPrivateKeyFromPEM(pem, "")
	if err != nil{
		return nil, err
	}

	return &PrivateKey{sk: pk}, nil
}

func GenerateECPrivateKey(curve string) (*PrivateKey, error) {
	args, ok := keyGenArgs[curve]
	if !ok {
		return nil, errors.New("unsupported curve")
	}

	pk, err :=  gmssl.GeneratePrivateKey("EC", args, nil)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{sk: pk}, nil
}

func (prv *PrivateKey)SignWithSM2(msg []byte) ([]byte, error) {
	sm2zid, _ := prv.sk.ComputeSM2IDDigest("1234567812345678")
	sm3ctx, _ := gmssl.NewDigestContext("SM3")
	sm3ctx.Reset()
	sm3ctx.Update(sm2zid)
	sm3ctx.Update(msg)
	tbs, _ := sm3ctx.Final()
	return prv.sk.Sign("sm2sign", tbs,nil)
}

func (pub *PublicKey)VerifyWithSM2(msg, sig []byte) bool{
	sm2zid, _ := pub.pk.ComputeSM2IDDigest("1234567812345678")
	sm3ctx, _ := gmssl.NewDigestContext("SM3")
	sm3ctx.Reset()
	sm3ctx.Update(sm2zid)
	sm3ctx.Update(msg)
	tbs, _ := sm3ctx.Final()
	return nil == pub.pk.Verify("tbs", tbs, sig, nil)
}

func (prv *PrivateKey)SignWithECDSASHA256(msg []byte) ([]byte, error) {
	dgst := HashWithSha256(msg)
	return prv.sk.Sign("ecdsa-with-SHA256", dgst,nil)
}

func (pub *PublicKey)VerifyWithECDSASHA256(msg, sig []byte) bool{
	dgst := HashWithSha256(msg)
	return nil == pub.pk.Verify("ecdsa-with-SHA256", dgst, sig, nil)
}

