package main

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/config"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/syncer"
	_ "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/controller"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/engine"
)

func main() {
	syncer.DefaultSyncer.Run()

	engine.Default.Run(config.Config.HttpConf.Addr())
}
