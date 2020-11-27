package cmd

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
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
			CSRGenerateCmd,
			SelfCAGenerateCmd,
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
		ArgsUsage: "--file <file> --keyfile <keyfile> --organization <organization> --commonName <commonName> --signatureAlg <signatureAlg>",
		Action:    generateCSR,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca generateCSR`,
	}

	SelfCAGenerateCmd = cli.Command{
		Name:      "genSelfSignCA",
		Usage:     "genSelfSignCA",
		ArgsUsage: "--file <file> --keyfile <keyfile> --organization <organization> --commonName <commonName> -- serialNumber <serialNumber> --signatureAlg <signatureAlg>",
		Action:    genSelfSignCA,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca genSelfSignCA`,
	}
)

func parseFlags (c *cli.Context) (string, string, string, string, string, string, string, int64, string){
	curve := c.String(CurveFlag.Name)
	file := c.String(OutFileFlag.Name)
	target := c.String(TargetFlag.Name)
	format := c.String(FormatFlag.Name)
	keyfile := c.String(KeyFileFlag.Name)
	organization := c.String(OrganizationFlags.Name)
	commonName := c.String(CommonNameFlag.Name)
	var serialNumber int64
	var err error
	if c.String(SerialNumberFlag.Name) != ""{
		serialNumber, err = strconv.ParseInt(c.String(SerialNumberFlag.Name), 10, 64)
		if nil != err {
			panic(err)
		}
	}else {
		serialNumber = -1
	}

	signatureAlg := c.String(SignatureAlgFlag.Name)
	return curve, file,target, format, keyfile, organization, commonName, serialNumber, signatureAlg
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
			syscall.Umask(0)
			os.Chmod(keyfile, 0666)
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
					fmt.Println(public)
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
						fmt.Println(public)
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
					fmt.Println(result)
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
			fmt.Println(pem)
			ioutil.WriteFile(file, []byte(pem), os.ModeCharDevice)
		}
	}
}

func generateKey(c *cli.Context) {
	curve, file, target, format, keyfile, _, _, _, _ := parseFlags(c)
	switch target {
	case "public":
			generatePublicKey(file, format, keyfile)
	case "private":
			generatePrivateKey(curve, file, format )
	case "pair":
			generateKeyPair(curve, file, format)
	default:
		panic("param invalid")

	}
}

func generateCSR(c *cli.Context) {
	_, file, _, _, keyfile, organization, commonName,_, signatureAlg := parseFlags(c)
	generateCsr(file, keyfile, organization, commonName, signatureAlg)
}

func generateCsr(file, keyfile, organization, commonName, signatureAlg string){
	syscall.Umask(0)
	os.Chmod(keyfile, 0666)
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
		fmt.Println(res)
	}else {
		fmt.Println(res)
		ioutil.WriteFile(file, []byte(res), os.ModeCharDevice)
	}
}

func genSelfSignCA(c *cli.Context) {
	_, file, _, _, keyfile, organization, commonName, serialNumber, signatureAlg := parseFlags(c)
	generateSelfSignCA(file, keyfile, organization, commonName, signatureAlg, serialNumber)
}

func generateSelfSignCA(file, keyfile, organization, commonName, signatureAlg string, serialNumber int64) {
	syscall.Umask(0)
	os.Chmod(keyfile, 0666)
	privateString := readFromFile(keyfile)
	privateKey, err := gmssl.NewPrivateKeyFromPEM(privateString)
	if nil != err {
		panic(err)
	}
	publicKey := privateKey.GetPublicKey()
	if serialNumber == -1 {
		panic("serialNumber must > 0")
	}

	selfCA, err := gmssl.CreateCerficate(privateKey, publicKey, signatureAlg, serialNumber, organization, commonName)
	if nil != err {
		panic(err)
	}
	res, err := selfCA.GetPEM()
	if nil != err {
		panic(err)
	}
	if file == "" {
		fmt.Println(res)
	}else {
		fmt.Println(res)
		ioutil.WriteFile(file, []byte(res), os.ModeCharDevice)
	}
}

//func generateCA(c *cli.Context)
