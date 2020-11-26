package cmd

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"strings"
)
var (
	CaCmd = cli.Command{
		Name:      "ca",
		Usage:     "Manage CA",
		ArgsUsage: "",
		Category:  "CA management",
		Description: `
	`,
		Subcommands: []cli.Command{
			KeyGenerateCmd,

		},
	}

	KeyGenerateCmd = cli.Command{
		Name:      "generateKey",
		Usage:     "generateKey",
		ArgsUsage: "--file <output> --curve <curve> --target <target> --format <format>",
		Action:    generateKey,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca generateKey`,
	}
	CSRGenerateCmd = cli.Command{
		Name:      "generateCSR",
		Usage:     "generateCSR",
		ArgsUsage: "--file <output> --curve <curve> --target <target> --format <format>",
		Action:    generateCSR,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca generateCSR`,
	}
)

func parseFlags (c *cli.Context) (string, string, string, string, string, string, string, string){
	curve := c.String(CurveFlag.Name)
	file := c.String(OutFileFlag.Name)
	target := c.String(TargetFlag.Name)
	format := c.String(FormatFlag.Name)
	keyfile := c.String(KeyFileFlag.Name)
	organization := c.String(OrganizationFlags.Name)
	commonName := c.String(CommonNameFlag.Name)
	signatureAlg := c.String(SignatureAlgFlag.Name)
	return curve, file,target, format, keyfile, organization, commonName, signatureAlg
}
func readFromFile(keyfile string) string {
	res, err := ioutil.ReadFile(keyfile)
	if err != nil {
		fmt.Println("read fail", err)
	}
	return string(res)
}

func generateKeyPair(curve, file, format string){
	privatefile := "private-" + file
	publicfile := "public-" + file
	generatePrivateKey(curve, privatefile, format)
	generatePublicKey(publicfile, format, privatefile)
}
func generatePublicKey(file, format, keyfile string)  {
	if keyfile == "" {
		panic("need private key")
	}else {
		if !strings.HasSuffix(keyfile, "PEM"){
			panic("private invalid")
		}else {
			privateString := readFromFile(keyfile)
			privateKey, err := gmssl.NewPrivateKeyFromPEM(privateString)
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
				publicKey, err := gmssl.NewPublicKeyFromPEM(public)
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
	prv, err := gmssl.GenerateECPrivateKey(curve)
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

func generateKey(c *cli.Context) {
	curve, file,target, format, keyfile, _, _, _ := parseFlags(c)
	switch target {
	case "public":
			generatePublicKey(file, format, keyfile)
	case "private":
			generatePrivateKey(curve, file, format )
	case "both":
			generateKeyPair(curve, file, format)

	}
}

func generateCSR(c *cli.Context) {
	_, file, _, _, keyfile, organization, commonName, signatureAlg := parseFlags(c)
	generateCsr(file, keyfile, organization, commonName, signatureAlg)
}

func generateCsr(file, keyfile, organization, commonName, signatureAlg string){
	privateString := readFromFile(keyfile)
	privateKey, err := gmssl.NewPrivateKeyFromPEM(privateString)
	if nil != err {
		panic(err)
	}
	csr, err := gmssl.CreateCertRequest(privateKey, signatureAlg, organization, commonName)
	if nil != err {
		panic(err)
	}
	res, err := csr.GetPEM()
	if nil != err {
		panic(err)
	}
	if file == "" {
		fmt.Println(csr)
	}else {
		ioutil.WriteFile(file, []byte(res), os.ModeCharDevice)
	}
}

