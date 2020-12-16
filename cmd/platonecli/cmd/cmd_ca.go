package cmd

import (
	"fmt"
	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"reflect"
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
			SetRootCertCmd,
			AddIssuerCmd,
			GetCertCmd,

		},
	}

	KeyGenerateCmd = cli.Command{
		Name:      "generateKey",
		Usage:     "generateKey",
		ArgsUsage: "--file <output> --curve <curve> --target <target> --format <format>",
		Action:    generateKey,
		Flags:     KeyGenerateFlags,
		Description: `
		platonecli ca generateKey`,
	}

	CSRGenerateCmd = cli.Command{
		Name:      "generateCSR",
		Usage:     "generateCSR",
		ArgsUsage: "--file <file> --keyfile <keyfile> --organization <organization> --commonName <commonName> --dgst <dgst>",
		Action:    generateCSR,
		Flags:     CSRGenerateFlags,
		Description: `
		platonecli ca generateCSR`,
	}

	SelfCAGenerateCmd = cli.Command{
		Name:      "genSelfSignCert",
		Usage:     "genSelfSignCert",
		ArgsUsage: "--file <file> --private <private> --organization <organization> --commonName <commonName> -- serial <serial> --dgst <dgst>",
		Action:    genSelfSignCert,
		Flags:     SelfCAGenerateFlags,
		Description: `
		platonecli ca selfCert`,
	}

	CaCreateCmd = cli.Command{
		Name:      "create",
		Usage:     "create",
		ArgsUsage: "--file <file> --private <private> --csr <csr file> -- serial <serial> --dgst <dgst>",
		Action:    generateCert,
		Flags:     CaCreateFlags,
		Description: `
		platonecli ca create`,
	}

	CaVerfyCmd = cli.Command{
		Name:      "verify",
		Usage:     "verify",
		ArgsUsage: "--file <file> --keyfile <keyfile> --organization <organization> --commonName <commonName> -- serial <serial> --dgst <dgst>",
		Action:    verifyCert,
		Flags:     CaVerfyFlags,
		Description: `
		platonecli ca verify`,
	}

	SetRootCertCmd = cli.Command{
		Name:      "setRootCert",
		Usage:     "setRootCert",
		ArgsUsage: " --ca",
		Action:    setRootCert,
		Flags:     SetRootCertFlags,
		Description: `
		platonecli ca setRootCert`,
	}

	AddIssuerCmd = cli.Command{
		Name:      "addIssuer",
		Usage:     "addIssuer",
		ArgsUsage: "--ca",
		Action:    addIssuer,
		Flags:     AddIssuerFlags,
		Description: `
		platonecli ca addIssuer`,
	}

	GetCertCmd = cli.Command{
		Name:      "getCert",
		Usage:     "getCert",
		ArgsUsage: "--all --root --subject",
		Action:    getCert,
		Flags:     GetCertFlags,
		Description: `
		platonecli ca getCert`,
	}
)

func parseFlags (c *cli.Context) (string, string, string, string, string, string, string, string, string,string,int64, string){
	curve := c.String(CurveFlag.Name)
	file := c.String(OutFileFlag.Name)
	target := c.String(TargetFlag.Name)
	format := c.String(FormatFlag.Name)
	private := c.String(PrivateFlag.Name)
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
	return curve, file,target, format, private, organization, commonName, csr, ca, cert,serialNumber, signatureAlg
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
	} else {
		panic("format error")
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

func genSelfSignCert(c *cli.Context) {
	_, file, _, _, keyfile, organization, commonName, _,_,_,serialNumber, signatureAlg := parseFlags(c)
	generateSelfSignCert(file, keyfile, organization, commonName, signatureAlg, serialNumber)
}

func generateSelfSignCert(file, keyfile, organization, commonName, signatureAlg string, serialNumber int64) {
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

func generateCert(c *cli.Context) {
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

func verifyCert(c *cli.Context) {
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

func setRootCert (c *cli.Context) {
	_, _, _, _, _, _, _, _, cafile, _,_, _ := parseFlags(c)
	cert:= readFromFile(cafile)
	funcParams := cmd_common.CombineFuncParams(cert)
	result := contractCall(c, funcParams, "setRootCert", precompile.CAManagementAddress)
	fmt.Printf("%v\n", result)
}

func addIssuer (c *cli.Context) {
	_, _, _, _, _, _, _, _, cafile, _,_, _ := parseFlags(c)
	cert:= readFromFile(cafile)
	funcParams := cmd_common.CombineFuncParams(cert)
	result := contractCall(c, funcParams, "addIssuer", precompile.CAManagementAddress)
	fmt.Printf("%v\n", result)
}

func getCert (c *cli.Context) {
	all := c.Bool(ShowAllFlags.Name)
	if all {
		result := contractCall(c, nil, "getAllCert", precompile.CAManagementAddress)
		strResult := PrintJson([]byte(result.(string)))
		fmt.Printf("result:\n%s\n", strResult)
		return
	}
	root := c.Bool(RootCertFlags.Name)

	if root {
		result := contractCall(c, nil, "getRootCert", precompile.CAManagementAddress)

		strResult := result.(string)

		fmt.Printf("result:\n%s\n", strResult)
		return
	}

	subject := c.String(SubjectFlag.Name)
	funcParams := cmd_common.CombineFuncParams(subject)
	result := contractCall(c, funcParams, "getCert", precompile.CAManagementAddress)
	strResult := PrintJson([]byte(result.(string)))
	fmt.Printf("result:\n%s\n", strResult)
}

func CertToString(res interface{}) interface{} {
	value := reflect.TypeOf(res)

	switch value.Kind() {
	case reflect.Uint64:

		return strconv.FormatUint(res.(uint64), 10)

	case reflect.Uint32:
		return strconv.FormatUint(uint64(res.(uint32)), 10)

	case reflect.String:
		fmt.Printf("string")
		return res

	default:
		panic("not support, please add the corresponding type")
	}
}


