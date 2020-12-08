package cmd

import (
	"encoding/json"
	//"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	//"reflect"
	"testing"
)

func TestGenerateECPrivateKey(t *testing.T) {
	type args struct {
		curve string
	}
	tests := []struct {
		name    string
		args    args
		want    *gmssl.PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{args:args{"SM2"}},
		{args:args{"secp256k1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generatePrivateKey("SM2", "private.PEM", "PEM" )
			//fmt.Println(got.GetText())

			//if (err != nil) != tt.wantErr {
			//	t.Errorf("GenerateECPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GenerateECPrivateKey() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
func TestGenerateCSR(t *testing.T) {
	type args struct {
		//curve string
		file string
		keyfile string
		organization string
		commonName string
		signatureAlg string
	}
	tests := []struct {
		name    string
		args    args
		//want    *gmssl.PrivateKey
		//wantErr bool
	}{
		// TODO: Add test cases.
		{args:args{"csr.PEM", "private.PEM", "wxbc", "test", "SHA256"}},
		//{args:args{"secp256k1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//generatePrivateKey("SM2", "private.PEM", "PEM" )
			generateCsr(tt.args.file, tt.args.keyfile, tt.args.organization, tt.args.commonName, tt.args.signatureAlg)
			//fmt.Println(got.GetText())

			//if (err != nil) != tt.wantErr {
			//	t.Errorf("GenerateECPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GenerateECPrivateKey() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestGenSelfCA(t *testing.T) {
	type args struct {
		//curve string
		file string
		keyfile string
		organization string
		commonName string
		signatureAlg string
		serialNumber int64
	}
	tests := []struct {
		name    string
		args    args
		//want    *gmssl.PrivateKey
		//wantErr bool
	}{
		// TODO: Add test cases.
		{args:args{"selfCA.PEM", "private.PEM", "wxbc", "test1", "SHA256", 1}},
		//{args:args{"secp256k1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//generatePrivateKey("SM2", "private.PEM", "PEM" )
			generateSelfSignCert(tt.args.file, tt.args.keyfile, tt.args.organization, tt.args.commonName, tt.args.signatureAlg, tt.args.serialNumber)
			cafile := readFromFile("selfCA.PEM")
			ca, err := gmssl.NewCertificateFromPEM(cafile)
			if err != nil{
				panic(err)
			}
			subject, err := ca.Cert.GetSubject()
			if err != nil{
				panic(err)
			}
			issuer, err := ca.Cert.GetIssuer()
			if err!=nil{
				panic(err)
			}
			println(issuer)
			b ,err := json.Marshal(subject)
			println(subject)

			println(b)
			//fmt.Println(got.GetText())

			//if (err != nil) != tt.wantErr {
			//	t.Errorf("GenerateECPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GenerateECPrivateKey() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestGenerateKeyPair(t *testing.T) {
	type args struct {
		curve string
		file string
		format string
	}
	tests := []struct {
		name    string
		args    args
		//want    *gmssl.PrivateKey
		//wantErr bool
	}{
		// TODO: Add test cases.
		{args:args{"SM2", "key111.PEM", "PEM"}},
		//{args:args{"secp256k1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//generatePrivateKey("SM2", "private.PEM", "PEM" )
			generateKeyPair(tt.args.curve, tt.args.file, tt.args.format)
			//fmt.Println(got.GetText())

			//if (err != nil) != tt.wantErr {
			//	t.Errorf("GenerateECPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GenerateECPrivateKey() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		curve string
		file string
		format string
	}
	tests := []struct {
		name    string
		args    args
		//want    *gmssl.PrivateKey
		//wantErr bool
	}{
		{args:args{"SM2", "key111.PEM", "PEM"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//generateSelfSignCA("test-selfCA.PEM", "private.PEM", "wxbc", "ljj", "sm3", 12)
			//generateCsr("test-csr.PEM", "private.PEM", "wxbc", "user", "sm3")
			//generateCAForCRS("test-CA.PEM", "private.PEM", "test-csr.PEM", "test-selfCA.PEM", 123, "sm3")
			cafile := readFromFile("test-selfCA.PEM")
			certfile := readFromFile("test-CA.PEM")
			ca, _ := gmssl.NewCertificateFromPEM(cafile)
			cert, _ := gmssl.NewCertificateFromPEM(certfile)
			bool, _ := gmssl.Verify(ca, cert)
			println(bool)
		})
	}
}