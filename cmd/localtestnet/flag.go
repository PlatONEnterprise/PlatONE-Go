package main

import (
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
)

var (
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: 6500,
	}

	P2PPortFlag = cli.IntFlag{
		Name:  "p2pport",
		Usage: "P2P network listening port",
		Value: 6600,
	}

	WSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "WS-RPC server listening port",
		Value: 6700,
	}
	DataDirFlag = cli.StringFlag{
		Name:  "datadir",
		Usage: "Testnet Data directory for node data",
		Value: "./localtestnetdata",
	}

	GCModeFlag = cli.StringFlag{
		Name:  "gcmode",
		Usage: `Blockchain garbage collection mode ("full", "archive")`,
		Value: "full",
	}

	NodeNumberFlag = cli.UintFlag{
		Name:  "number",
		Usage: "Number of node in testnet configuration (default = 4)",
		Value: 4,
	}

	AutoClearOldDataFlag = cli.BoolTFlag{
		Name:  "autoclear",
		Usage: "auto clear all old data",
	}
)

var (
	rpcPortBase int
	p2pPortBase int
	wsPortBase  int
	dataDirBase string
	gcmode      string

	nodeNumber int
	autoClear  bool
)

var curPath string

func init() {
	curPath = curDir()
}

func parseFlag(ctx *cli.Context) {
	rpcPortBase = ctx.Int(RPCPortFlag.Name)
	p2pPortBase = ctx.Int(P2PPortFlag.Name)
	wsPortBase = ctx.Int(WSPortFlag.Name)
	dataDirBase = ctx.String(DataDirFlag.Name)
	if !filepath.IsAbs(dataDirBase) {
		dataDirBase = filepath.Join(curPath, dataDirBase)
	}
	gcmode = ctx.String(GCModeFlag.Name)

	nodeNumber = ctx.Int(NodeNumberFlag.Name)
	autoClear = ctx.BoolT(AutoClearOldDataFlag.Name)

	setGenesisFilePath()
}
