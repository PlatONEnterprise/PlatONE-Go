package main

import (
	"fmt"
	"os"
)

var (
	defaultSNFlag = defaultStartNodeFlag{
		ip:          "0.0.0.0",
		pprof:       "--pprof --pprofaddr 0.0.0.0 ",
		wasmlog:     "wasm_log",
		wasmlogsize: 67108864,
		nodeKey:     "nodekey",
		ipcpath:     "platone.ipc",
		logsDir:     "logs",

		wsorigins:     "*",
		rpccorsdomain: "*",
		identity:      "platone",
		bootnodes:     "",
		other:         "--nodiscover --debug",
	}
)

type defaultStartNodeFlag struct {
	ip            string //common net address for p2p,rpc,ws
	pprof         string
	wasmlog       string
	wasmlogsize   int
	nodeKey       string
	ipcpath       string
	wsorigins     string
	rpccorsdomain string
	identity      string
	bootnodes     string
	other         string
	logsDir       string
}

func initDefaultStartNodeEnv(datadir string) {
	if err := os.MkdirAll(fmt.Sprintf("%s/%s/%s", datadir, defaultSNFlag.logsDir, defaultSNFlag.wasmlog), os.ModePerm); nil != err {
		panic(err)
	}

	genNodeKeyFile(fmt.Sprintf("%s/%s", datadir, defaultSNFlag.nodeKey))
}

func (this defaultStartNodeFlag) ToFlags(datadir string) string {
	return fmt.Sprintf(` --identity %s --wsorigins "%s" --rpccorsdomain "%s" --bootnodes %s  %s %s --nodekey %s/%s --datadir %s --ipcpath %s/%s --wasmlog %s/%s/%s --wasmlogsize %d --moduleLogParams '{"platone_log": ["/"], "__dir__": ["%s/%s"], "__size__": ["67108864"]}'`,
		this.identity, this.wsorigins, this.rpccorsdomain, this.bootnodes,
		this.pprof, this.other,
		datadir, this.nodeKey,
		datadir,
		datadir, this.ipcpath,
		datadir, this.logsDir, this.wasmlog,
		this.wasmlogsize,
		datadir, this.logsDir,
	)
}

type startNodeConfig struct {
	P2PPort int
	RPCPort int
	WSPort  int
	GCMode  string
	DataDir string
}

func newStartNodeConfig(p2PPort, RPCPort, WSPort int, GCMode, dataDir string) *startNodeConfig {
	return &startNodeConfig{P2PPort: p2PPort, RPCPort: RPCPort, WSPort: WSPort, GCMode: GCMode, DataDir: dataDir}
}

func (this *startNodeConfig) ToFlag() string {
	return fmt.Sprintf(` %s %s %s %s %s`, this.P2PFlag(), this.RPCFlag(), this.WSFlag(), this.GCModeFlag(), defaultSNFlag.ToFlags(this.DataDir))
}

func (this *startNodeConfig) P2PFlag() string {
	return fmt.Sprintf(" --port %d ", this.P2PPort)
}

func (this *startNodeConfig) RPCFlag() string {
	return fmt.Sprintf(" --rpc --rpcaddr %s --rpcport %d --rpcapi db,eth,net,web3,admin,personal,txpool,istanbul  ", defaultSNFlag.ip, this.WSPort)
}

func (this *startNodeConfig) WSFlag() string {
	return fmt.Sprintf(" --ws --wsaddr %s --wsport %d ", defaultSNFlag.ip, this.WSPort)
}

func (this *startNodeConfig) GCModeFlag() string {
	return fmt.Sprintf(" --gcmode  %s ", this.GCMode)
}
