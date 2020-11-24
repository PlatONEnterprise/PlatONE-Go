package main

import (
	"fmt"
	"os"
)

type startNodeConfig struct {
	P2PPort int
	RPCPort int
	WSPort  int
	DataDir string
	GCMode  string

	//have default value
	ip            string //common net address for p2p,rpc,ws
	pprof         string
	wasmlog       string
	wasmlogsize   int
	nodeKeyFile   string
	ipcpath       string
	wsorigins     string
	rpccorsdomain string
	identity      string
	bootnodes     string
	other         string
	logsDir       string

	//not startnode config
	index             int
	errLogFileHandler *os.File
}

var (
	tplSNConfig = startNodeConfig{
		ip:          "0.0.0.0",
		pprof:       "--pprof --pprofaddr 0.0.0.0 ",
		wasmlog:     "wasm_log",
		wasmlogsize: 67108864,
		nodeKeyFile: "nodekey",
		ipcpath:     "platone.ipc",
		logsDir:     "logs",

		wsorigins:     "*",
		rpccorsdomain: "*",
		identity:      "platone",
		bootnodes:     "",
		other:         "--nodiscover --debug",
	}
)

func newNodeConfig(nodeNumber int) *startNodeConfig {
	newConf := tplSNConfig
	newConf.index = nodeNumber

	newConf.DataDir = fmt.Sprintf("%s/node-%d", dataDirBase, nodeNumber)
	newConf.P2PPort = p2pPortBase + nodeNumber
	newConf.RPCPort = rpcPortBase + nodeNumber
	newConf.WSPort = wsPortBase + nodeNumber
	newConf.GCMode = gcmode

	return &newConf
}

func (this *startNodeConfig) ToFlag() string {
	flags := fmt.Sprintf(` %s %s %s %s %s`,
		this.P2PFlag(), this.RPCFlag(), this.WSFlag(), this.GCModeFlag(),
		fmt.Sprintf(` --identity %s --wsorigins "%s" --rpccorsdomain "%s"  %s %s --nodekey %s/%s --datadir %s --ipcpath %s/%s --wasmlog %s/%s/%s --wasmlogsize %d --moduleLogParams {"platone_log":["/"],"__dir__":["%s/%s"],"__size__":["67108864"]}`,
			this.identity, this.wsorigins, this.rpccorsdomain,
			this.pprof, this.other,
			this.DataDir, this.nodeKeyFile,
			this.DataDir,
			this.DataDir, this.ipcpath,
			this.DataDir, this.logsDir, this.wasmlog,
			this.wasmlogsize,
			this.DataDir, this.logsDir,
		))

	if this.bootnodes != "" {
		flags = fmt.Sprintf("%s --bootnodes %s", flags, this.bootnodes)
	}

	return flags
}

func (this *startNodeConfig) P2PFlag() string {
	return fmt.Sprintf(" --port %d ", this.P2PPort)
}

func (this *startNodeConfig) RPCFlag() string {
	return fmt.Sprintf(" --rpc --rpcaddr %s --rpcport %d --rpcapi db,eth,net,web3,admin,personal,txpool,istanbul  ", this.ip, this.RPCPort)
}

func (this *startNodeConfig) WSFlag() string {
	return fmt.Sprintf(" --ws --wsaddr %s --wsport %d ", this.ip, this.WSPort)
}

func (this *startNodeConfig) GCModeFlag() string {
	return fmt.Sprintf(" --gcmode  %s ", this.GCMode)
}
