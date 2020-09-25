package main

import (
	"data-manager/config"
	"data-manager/syncer"
	_ "data-manager/web/controller"
	"data-manager/web/engine"
)

func main() {
	syncer.DefaultSyncer.Run()

	engine.Default.Run(config.Config.HttpConf.Addr())
}
