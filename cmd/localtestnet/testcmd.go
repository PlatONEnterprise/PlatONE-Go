package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	testnetCommand = cli.Command{
		//Action:   utils.MigrateFlags(testnetChain),
		Action:   testnetChain,
		Name:     "start",
		Usage:    "start platone testnet [flags]",
		Category: "TESTNET COMMANDS",
		Description: `
testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).`,
		Flags: []cli.Flag{
			TestnetNodeNumberFlag,
			BinaryDirFlag,
			DataDirFlag,
			P2PPortFlag,
			RPCPortFlag,
			WSPortFlag,
			GCModeFlag,
		},
	}
)

var curPath string

func init() {
	var err error
	curPath, err = os.Getwd()
	if nil != err {
		panic(err)
	}
}

func testnetChain(ctx *cli.Context) error {
	platoneBin := filepath.Join(curPath, ctx.String(BinaryDirFlag.Name), "./platone")
	//platonecliBin := filepath.Join(ctx.String(PlatONEDirFlag.Name), "platonecli")

	nodeNumber := ctx.Int(TestnetNodeNumberFlag.Name)
	for i := 0; i < nodeNumber; i++ {
		conf := buildNodeConfig(i, ctx)
		initNodeEnv(conf)
		if err := startNode(platoneBin, conf); nil != err {
			panic(err)
		}
	}

	return nil
}

func buildNodeConfig(nodeNumber int, ctx *cli.Context) *startNodeConfig {
	rpcPortBase := ctx.Int(RPCPortFlag.Name)
	p2pPortBase := ctx.Int(P2PPortFlag.Name)
	wsPortBase := ctx.Int(WSPortFlag.Name)
	gcmode := ctx.String(GCModeFlag.Name)

	dataDirBase := filepath.Join(curPath, ctx.String(DataDirFlag.Name))

	conf := newStartNodeConfig(p2pPortBase, rpcPortBase, wsPortBase, gcmode, dataDirBase)

	conf.WSPort += nodeNumber
	conf.RPCPort += nodeNumber
	conf.P2PPort += nodeNumber
	conf.DataDir = fmt.Sprintf("%s/node-%d", conf.DataDir, nodeNumber)

	return conf
}

func initNodeEnv(conf *startNodeConfig) {
	if err := os.MkdirAll(conf.DataDir, os.ModePerm); nil != err {
		panic(err)
	}

	initDefaultStartNodeEnv(conf.DataDir)
}

func startNode(platoneBin string, conf *startNodeConfig) error {
	args := conf.ToFlag()

	logs := fmt.Sprintf(" 1>/dev/null 2>%s/%s/platone_error.log &", conf.DataDir, defaultSNFlag.logsDir)
	cmd := exec.Command("nohup", platoneBin, args, logs)
	ret, err := cmd.CombinedOutput()
	if nil != err {
		log.Println("failed to exec cmd:", cmd.String())
		panic(err)
	}
	log.Println("cmd:", platoneBin, args, "ret:", string(ret))

	return nil
}
