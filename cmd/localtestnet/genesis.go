package main

import (
	"encoding/json"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core"
	"github.com/PlatONEnetwork/PlatONE-Go/p2p/discover"
	"math/big"
	"os"
)

var (
	genesisTplPath = "./conf/genesis.json.istanbul.template"
	genesisPath    = "./conf/genesis.json"
)

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

func buildGenesis(datadir string, g *core.Genesis) {
	node, err := discover.ParseNode(getEnode())
	if nil != err {
		panic(err)
	}

	g.Config.Istanbul.ValidatorNodes = append(g.Config.Istanbul.ValidatorNodes, *node)

	addr := genAccount(datadir)
	g.Alloc[common.HexToAddress(addr)] = core.GenesisAccount{Balance: big.NewInt(100000000000000)}
}

func genGenesisFile(g *core.Genesis) {
	file, err := os.Create(genesisPath)
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
