package cmd

import (
	"fmt"
	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
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
			CaCreateCmd,
			CaVerfyCmd,
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
		ArgsUsage: "--file <file> --keyfile <keyfile> --organization <organization> --commonName <commonName> --dgst <dgst>",
		Action:    generateCSR,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca generateCSR`,
	}

	SelfCAGenerateCmd = cli.Command{
		Name:      "genSelfSignCA",
		Usage:     "genSelfSignCA",
		ArgsUsage: "--file <file> --keyfile <keyfile> --organization <organization> --commonName <commonName> -- serial <serial> --dgst <dgst>",
		Action:    genSelfSignCA,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca genSelfSignCA`,
	}

	CaCreateCmd = cli.Command{
		Name:      "create",
		Usage:     "create",
		ArgsUsage: "--file <file> --keyfile <keyfile> --organization <organization> --commonName <commonName> -- serial <serial> --dgst <dgst>",
		Action:    generateCA,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca create`,
	}

	CaVerfyCmd = cli.Command{
		Name:      "verify",
		Usage:     "verify",
		ArgsUsage: "--file <file> --keyfile <keyfile> --organization <organization> --commonName <commonName> -- serial <serial> --dgst <dgst>",
		Action:    verifyCa,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca verify`,
	}

	SetRootCACmd = cli.Command{
		Name:      "setRootCA",
		Usage:     "setRootCA",
		ArgsUsage: " --ca",
		Action:    setRootCA,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca setRootCA`,
	}

	AddIssuerCmd = cli.Command{
		Name:      "addIssuer",
		Usage:     "addIssuer",
		ArgsUsage: "--ca",
		Action:    addIssuer,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca addIssuer`,
	}

	GetCACmd = cli.Command{
		Name:      "getCA",
		Usage:     "getCA",
		ArgsUsage: "--all --root --subject",
		Action:    getCA,
		Flags:     CaCmdFlags,
		Description: `
		platonecli ca getCA`,
	}
)

func parseFlags (c *cli.Context) (string, string, string, string, string, string, string, string, string,string,int64, string){
	curve := c.String(CurveFlag.Name)
	file := c.String(OutFileFlag.Name)
	target := c.String(TargetFlag.Name)
	format := c.String(FormatFlag.Name)
	keyfile := c.String(KeyFileFlag.Name)
	organization := c.String(OrganizationFlags.Name)
	commonName := c.String(CommonNameFlag.Name)
	csr := c.String(CsrFileFlag.Name)
	ca := c.String(CaFileFlag.Name)
	cert := c.String(CertFileFlag.Name)
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
	return curve, file,target, format, keyfile, organization, commonName, csr, ca, cert,serialNumber, signatureAlg
}

func readFromFile(keyfile string) string {
	res, err := ioutil.ReadFile(keyfile)
	if err != nil {
		fmt.Println("read fail", err)
	}
	return string(res)
}

func writeToFile(outfile ,content string) {
	if outfile == "" {
		fmt.Println(content)
	} else {
		fmt.Println(content)
		err := ioutil.WriteFile(outfile, []byte(content), 0666)
		if err != nil {
			panic(err)
		}
	}
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
				writeToFile(file, public)
			case "PEM":
				if !strings.HasSuffix(file, "PEM"){
					panic(err)
				}else {
					public, err :=privateKey.GetPublicKeyPEM()
					if nil != err {
						panic(err)
					}
					writeToFile(file, public)
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
				writeToFile(file, result)
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

		writeToFile(file, pem)
	}
}

func generateKey(c *cli.Context) {
	curve, file,target, format, keyfile, _, _, _,_,_,_, _ := parseFlags(c)
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
	_, file, _, _, keyfile, organization, commonName,_,_,_,_, signatureAlg := parseFlags(c)
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
	writeToFile(file, res)
}

func genSelfSignCA(c *cli.Context) {
	_, file, _, _, keyfile, organization, commonName, _,_,_,serialNumber, signatureAlg := parseFlags(c)
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
	writeToFile(file, res)
}

func generateCA(c *cli.Context) {
	_, outfile, _, _, keyfile, _, _, csrfile, cafile,_,serialNumber, alg := parseFlags(c)
	generateCAForCRS(outfile, keyfile, csrfile, cafile, serialNumber, alg)
}

func generateCAForCRS(outfile string, keyfile string, csrfile string, cafile string, serialNumber int64, alg string) {
	caPEM := readFromFile(cafile)
	csrPEM := readFromFile(csrfile)
	keyPEM := readFromFile(keyfile)

	ca, err := gmssl.NewCertificateFromPEM(caPEM)
	if err != nil{
		panic(err)
	}

	csr, err := gmssl.NewCertRequestFromPEM(csrPEM)
	if err != nil{
		fmt.Println("NewCertRequestFromPEM FAILED")
		panic(err)
	}

	prv, err := gmssl.NewPrivateKeyFromPEM(keyPEM)

	cert, err := gmssl.CreateCertificateForReq(prv, csr, ca, alg, serialNumber)
	if err != nil{
		panic(err)
	}

	certPEM, err := cert.GetPEM()
	if err != nil{
		panic(err)
	}

	writeToFile(outfile, certPEM)
}

func verifyCa(c *cli.Context) {
	_, _, _, _, _, _, _, _, cafile, certfile,_, _ := parseFlags(c)
	verify(cafile, certfile)
}

func verify(cafile , certfile string) {
	caPEM := readFromFile(cafile)
	ca, err := gmssl.NewCertificateFromPEM(caPEM)
	if err != nil{
		panic(err)
	}

	certPEM := readFromFile(certfile)
	cert, err := gmssl.NewCertificateFromPEM(certPEM)
	if err != nil{
		panic(err)
	}

	ret, err := gmssl.Verify(ca, cert)

	if !ret {
		fmt.Println("verify failed!")
		if err != nil {
			fmt.Println(err)
		}
	}else {
		fmt.Println("verify success!")
	}
}

func setRootCA (c *cli.Context) {
	_, _, _, _, _, _, _, _, cafile, _,_, _ := parseFlags(c)
	funcParams := cmd_common.CombineFuncParams(cafile)
	result := contractCall(c, funcParams, "setRootCA", precompile.CAManagementAddress)
	fmt.Printf("%v\n", result)
}

func addIssuer (c *cli.Context) {
	_, _, _, _, _, _, _, _, cafile, _,_, _ := parseFlags(c)
	funcParams := cmd_common.CombineFuncParams(cafile)
	result := contractCall(c, funcParams, "addIssuer", precompile.CAManagementAddress)
	fmt.Printf("%v\n", result)
}

func getCA (c *cli.Context) {
	all := c.Bool(ShowAllFlags.Name)
	if all {
		result := contractCall(c, nil, "getAllCA", precompile.CAManagementAddress)
		strResult := PrintJson([]byte(result.(string)))
		fmt.Printf("result:\n%s\n", strResult)
		return
	}
	root := c.Bool(RootCAFlags.Name)
	if root {
		result := contractCall(c, nil, "getRootCA", precompile.CAManagementAddress)
		fmt.Printf("%v\n", result)
		return
	}
	subject := c.String(SubjectFlag.Name)
	funcParams := cmd_common.CombineFuncParams(subject)
	result := contractCall(c, funcParams, "getCA", precompile.CAManagementAddress)
	fmt.Printf("%v\n", result)
}




//func generateCA(c *cli.Context)