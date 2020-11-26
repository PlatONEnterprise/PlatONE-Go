package cmd

import (
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/gmssl"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
)
var (
	CaCmd = cli.Command{
		Name:      "account",
		Usage:     "Manage accounts",
		ArgsUsage: "",
		Category:  "account",
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
)

func parseFlags (c *cli.Context) (string, string, string, string, string){
	curve := c.String(CurveFlag.Name)
	file := c.String(OutFileFlag.Name)
	target := c.String(TargetFlag.Name)
	format := c.String(FormatFlag.Name)
	keyfile := c.String(KeyFileFlag.Name)

	return curve, file,target, format, keyfile
}

func generateKey(c *cli.Context) {
	curve, file,target, format, _ := parseFlags(c)

	prv, err := gmssl.GenerateECPrivateKey(curve)
	if err != nil{
		panic(err)
	}

	if format == "PEM" && target == "private" {
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


