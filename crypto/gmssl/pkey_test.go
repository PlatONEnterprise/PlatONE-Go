package gmssl

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestGenerateECPrivateKey(t *testing.T) {
	type args struct {
		curve string
	}
	tests := []struct {
		name    string
		args    args
		want    *PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{args:args{"SM2"}},
		{args:args{"secp256k1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//got, err := GenerateECPrivateKey(tt.args.curve)
			//fmt.Println(got.GetText())
			//
			//if (err != nil) != tt.wantErr {
			//	t.Errorf("GenerateECPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GenerateECPrivateKey() got = %v, want %v", got, tt.want)
			//}
			//generatePrivateKey("SM2", "prv.PEM", "PEM")
			//generatePublicKey("pub.PEM", "PEM", "prv.PEM")
			rootPem := readFromFile("selfCA.PEM")
			certPem := readFromFile("targetCA.PEM")
			root, _ := NewCertificateFromPEM(rootPem)
			cert, _ := NewCertificateFromPEM(certPem)
			pub, _ := root.Cert.GetPublicKey()
			hex, _ := pub.GetHex()
			res, _ := Verify(root, cert)
			println(res)
			println(hex)
		})
	}
}
func readFromFile(keyfile string) string {
	res, err := ioutil.ReadFile(keyfile)
	if err != nil {
		fmt.Println("read fail", err)
	}
	return string(res)
}
func generatePublicKey(file, format, keyfile string)  {
	if keyfile == "" {
		panic("need private key")
	}else {
		if !strings.HasSuffix(keyfile, "PEM"){
			panic("private invalid")
		}else {
			privateString := readFromFile(keyfile)
			privateKey, err := NewPrivateKeyFromPEM(privateString)
			if nil != err {
				panic(err)
			}
			switch format {
			case "HEX":
				public, err :=privateKey.GetPublicKeyHex()
				if nil != err {
					panic(err)
				}
				if file == "" {
					fmt.Println(public)
				}else {
					ioutil.WriteFile(file, []byte(public), os.ModeCharDevice)
				}
			case "PEM":
				if !strings.HasSuffix(file, "PEM"){
					panic(err)
				}else {
					public, err :=privateKey.GetPublicKeyPEM()
					if nil != err {
						panic(err)
					}
					if file == "" {
						fmt.Println(public)
					}else {
						ioutil.WriteFile(file, []byte(public), os.ModeCharDevice)
					}
				}
			case "TXT":
				public, err :=privateKey.GetPublicKeyPEM()
				if nil != err {
					panic(err)
				}
				publicKey, err := NewPublicKeyFromPEM(public)
				if nil != err {
					panic(err)
				}
				result, err := publicKey.GetText()
				if nil != err {
					panic(err)
				}
				if file == "" {
					fmt.Println(result)
				}else {
					ioutil.WriteFile(file, []byte(result), os.ModeCharDevice)
				}


			}
		}
	}

}

func generatePrivateKey(curve, file, format string)  {
	prv, err := GenerateECPrivateKey(curve)
	if err != nil{
		panic(err)
	}

	if format == "PEM" {
		pem, err := prv.GetPEM()
		if err != nil {
			panic(err)
		}

		if file == "" {
			fmt.Println(pem)
		}else {
			ioutil.WriteFile(file, []byte(pem), os.ModeCharDevice)
		}
	}
}




