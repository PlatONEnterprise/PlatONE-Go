package main

import (
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
	"strings"
)

var (
	testnetCommand = cli.Command{
		//Action:   utils.MigrateFlags(testnetChain),
		Action:   startTestnetChain,
		Name:     "start",
		Usage:    "start platone testnet [flags]",
		Category: "TESTNET COMMANDS",
		Description: `
testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).`,
		Flags: []cli.Flag{
			NodeNumberFlag,
			DataDirFlag,
			P2PPortFlag,
			RPCPortFlag,
			WSPortFlag,
			GCModeFlag,
			AutoClearOldDataFlag,
		},
	}
)

func startTestnetChain(ctx *cli.Context) error {
	parseFlag(ctx)
	if autoClear {
		clearDataAndKillProcess(dataDirBase)
	}

	for i := 0; i < nodeNumber; i++ {
		conf := newNodeConfig(i)
		initStartNodeEnv(conf)
		if i == 0 {
			genGenesisFile(account, bootnode)
		}

		initGenesis(conf)
		if i != 0 {
			conf.bootnodes = bootnode
		}
		startNode(conf)
	}

	return nil
}

func startNode(conf *startNodeConfig) {
	args := strings.Split(conf.ToFlag(), " ")

	arg := make([]string, 0, len(args)+1)
	platoneBin := filepath.Join(curPath, "./platone")
	arg = append(arg, platoneBin)

	for _, a := range args {
		if a != "" {
			arg = append(arg, a)
		}
	}

	pid := StartCmd("nohup", conf.errLogFileHandler, arg...)
	savePID(pid)
}
