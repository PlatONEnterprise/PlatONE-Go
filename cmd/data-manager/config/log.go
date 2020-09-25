package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

func initLog() {
	if LogOutputStd == Config.LogConf.Output {
		logrus.SetOutput(os.Stdout)
	} else if LogOutputFile == Config.LogConf.Output {
		if err := os.MkdirAll(Config.LogConf.FileDirAbs, 0755); err != nil {
			panic(err)
		}

		f, err := os.Create(Config.LogConf.FilePath)
		if err != nil {
			panic(err)
		}

		logrus.SetOutput(f)
	}

	logLevel, err := logrus.ParseLevel(Config.LogConf.Level)
	if nil != err {
		panic(err)
	}

	logrus.SetLevel(logLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{})
}
