package main

import (
	"encoding/json"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core"
	"github.com/PlatONEnetwork/PlatONE-Go/p2p/discover"
	"math/big"
	"os"
	"path/filepath"
)

var (
	genesisTplPath = "./conf/genesis.json.istanbul.template"
	genesisFile    = "genesis.json"
)

func initGenesis(conf *startNodeConfig) {
	RunCmd("./platone", "init", genesisFile, "--datadir", conf.DataDir)
}

func setGenesisFilePath() {
	genesisFile = filepath.Join(dataDirBase, genesisFile)
}

func genGenesisFile(account, bootnode string) {
	g := loadGenesisTpl()
	buildGenesis(account, bootnode, g)

	file, err := os.Create(genesisFile)
	if err != nil {
		panic(fmt.Errorf("Failed to read genesis file: %v", err))
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(g); nil != err {
		panic(err)
	}
}

func loadGenesisTpl() *core.Genesis {
	file, err := os.Open(genesisTplPath)
	if err != nil {
		panic(fmt.Errorf("Failed to read genesis file: %v", err))
	}
	defer file.Close()

	genesis := new(core.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		panic(fmt.Errorf("invalid genesis file: %v", err))
	}

	return genesis
}

func buildGenesis(account, bootnode string, g *core.Genesis) {
	node, err := discover.ParseNode(bootnode)
	if nil != err {
		panic(err)
	}

	g.Config.Istanbul.FirstValidatorNode = *node

	g.Alloc[common.HexToAddress(account)] = core.GenesisAccount{Balance: big.NewInt(100000000000000)}
}
