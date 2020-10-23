package main

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	testnetCommand = cli.Command{
		Action:   utils.MigrateFlags(testnetChain),
		Name:     "testnet",
		Usage:    "platone testnet [flags]",
		Category: "TESTNET COMMANDS",
		Description: `
testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).`,
		Flags: []cli.Flag{
			TestnetNodeNumberFlag,
			PlatONEDirFlag,
			PlatONECliDirFlag,
			DataDirFlag,
			P2PPortFlag,
			RPCPortFlag,
			WSPortFlag,
			GCModeFlag,
		},
	}
)

func testnetChain(ctx *cli.Context) error {
	platoneBin := filepath.Join(ctx.GlobalString(PlatONEDirFlag.Name), "platone")
	//platonecliBin := filepath.Join(ctx.GlobalString(PlatONEDirFlag.Name), "platone")
	currentPath, err := os.Getwd()
	if nil != err {
		panic(err)
	}

	rpcPortBase := ctx.GlobalInt(RPCPortFlag.Name)
	p2pPortBase := ctx.GlobalInt(P2PPortFlag.Name)
	wsPortBase := ctx.GlobalInt(WSPortFlag.Name)
	dataDirBase := filepath.Join(currentPath, ctx.GlobalString(DataDirFlag.Name))
	gcmode := ctx.GlobalString(GCModeFlag.Name)

	conf := newStartNodeConfig(p2pPortBase, rpcPortBase, wsPortBase, gcmode, dataDirBase)

	nodeNumber := ctx.GlobalInt(PlatONEDirFlag.Name)
	for i := 0; i < nodeNumber; i++ {
		if err := startNode(i, platoneBin, *conf); nil != err {
			panic(err)
		}
	}

	return nil
}

func startNode(nodeNumber int, platoneBin string, conf startNodeConfig) error {
	conf.WSPort += nodeNumber
	conf.RPCPort += nodeNumber
	conf.P2PPort += nodeNumber
	conf.DataDir = fmt.Sprintf("%s/node-%d", conf.DataDir, nodeNumber)

	args := conf.ToFlag()

	logs := fmt.Sprintf("1>/dev/null 2>%s/%s/platone_error.log &", conf.DataDir, defaultSNFlag.logsDir)
	cmd := exec.Command("nohup", platoneBin, args, logs)
	ret, err := cmd.Output()
	if nil != err {
		panic(err)
	}
	log.Println("cmd:", platoneBin, args, "ret:", string(ret))

	return nil
}
