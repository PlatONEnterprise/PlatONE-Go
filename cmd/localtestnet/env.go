package main

import (
	"fmt"
	"os"
)

func initStartNodeEnv(conf *startNodeConfig) {
	if err := os.MkdirAll(conf.DataDir, os.ModePerm); nil != err {
		panic(err)
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/%s/%s", conf.DataDir, conf.logsDir, conf.wasmlog), os.ModePerm); nil != err {
		panic(err)
	}

	if 0 == conf.index {
		genAccount(conf.DataDir)
	}

	genNodeKeyFile(fmt.Sprintf("%s/%s", conf.DataDir, conf.nodeKeyFile))

	//create platone error log
	var err error
	if conf.errLogFileHandler, err = os.Create(fmt.Sprintf("%s/%s/platone_error.log", conf.DataDir, conf.logsDir)); nil != err {
		panic(err)
	}
}
