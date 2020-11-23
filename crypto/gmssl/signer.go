package gmssl


func SignWithSM2(sk *PrivateKey, msg []byte) ([]byte, error){
	return sk.SignWithSM2(msg)
}

func VerifyWithSM2(pk *PublicKey, msg, sig []byte) bool {
	return pk.VerifyWithSM2(msg, sig)
}

func SignWithECDSASHA256(sk *PrivateKey, msg []byte) ([]byte, error) {
	return sk.SignWithECDSASHA256(msg)
}

func VerifyWithECDSASHA256(pk *PublicKey, msg, sig []byte) bool {
	return pk.VerifyWithECDSASHA256(msg, sig)
}
