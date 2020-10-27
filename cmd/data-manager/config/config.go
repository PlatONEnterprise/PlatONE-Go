package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"math/rand"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//std | file
const (
	LogOutputStd  = "std"
	LogOutputFile = "file"
)

type config struct {
	HttpConf        *httpConf        `toml:"http"`
	DBConf          *dbConf          `toml:"db"`
	LogConf         *logConf         `toml:"log"`
	SyncConf        *syncConf        `toml:"sync"`
	ChainConf       *chainConf       `toml:"chain"`
	SyncTxCountConf *syncTxCountConf `toml:"sync-tx-count"`
}

type chainConf struct {
	ID             uint64 `toml:"id"`
	NodeRestServer string `toml:"node_rest_server"`
	NodeRpcAddress string `toml:"node_rpc_address"`
}

type logConf struct {
	Level      string `toml:"level"`
	Output     string `toml:"output"`
	FilePath   string `toml:"filepath"`
	FileDirAbs string
	FileName   string
}

type httpConf struct {
	IP    string `toml:"ip"`
	Port  int    `toml:"port"`
	Debug bool   `toml:"debug"`
}

func (this *httpConf) Addr() string {
	return fmt.Sprintf("%s:%d", this.IP, this.Port)
}

type dbConf struct {
	IP       string `toml:"ip"`
	Port     string `toml:"port"`
	UserName string `toml:"username"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
}

func (this *dbConf) Uri() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s",
		this.UserName,
		this.Password,
		this.IP,
		this.Port,
	)
}

type syncConf struct {
	Interval int      `toml:"interval"`
	Urls     []string `toml:"urls"`
}

func (this *syncConf) SyncInterval() time.Duration {
	return time.Second * time.Duration(this.Interval)
}

func (this *syncConf) URLs() []string {
	return this.Urls
}

func (this *syncConf) RandomURL() string {
	randIndex := rand.Intn(len(this.Urls))

	return this.Urls[randIndex]
}

type syncTxCountConf struct {
	When     string `toml:"when"`
	TryTimes int    `toml:"try_times"`
}

func (this *syncTxCountConf) GetWhen() time.Time {
	reg, err := regexp.Compile(`([0-9]+):([0-9]+):([0-9]+)`)
	if nil != err {
		logrus.Panic(err)
	}

	hms := reg.FindAllStringSubmatch(this.When, -1)[0]

	h, err := strconv.Atoi(hms[1])
	if nil != err {
		logrus.Panic(err)
	}

	m, err := strconv.Atoi(hms[2])
	if nil != err {
		logrus.Panic(err)
	}

	s, err := strconv.Atoi(hms[3])
	if nil != err {
		logrus.Panic(err)
	}

	now := time.Now()
	when := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, now.Location())

	return when
}

var Config config

const configFile = "./config.toml"

func init() {
	initFromFile(configFile)
}

func initFromFile(file string) {
	if _, err := toml.DecodeFile(file, &Config); err != nil {
		panic(err)
	}

	validateConfig()

	initLog()
}

func validateConfig() {
	if _, err := logrus.ParseLevel(Config.LogConf.Level); err != nil {
		panic(err)
	}

	if Config.LogConf.Output != LogOutputStd &&
		Config.LogConf.Output != LogOutputFile {
		panic("invalid log output")
	}

	if Config.LogConf.Output == LogOutputFile {
		if "" == strings.TrimSpace(Config.LogConf.FilePath) {
			panic("invalid log.filepath")
		}

		Config.LogConf.FileDirAbs, Config.LogConf.FileName = filepath.Split(Config.LogConf.FilePath)
	}

	if Config.DBConf.UserName == "" ||
		Config.DBConf.IP == "" ||
		Config.DBConf.Port == "" ||
		Config.DBConf.DBName == "" {
		panic("invalid db config")
	}
}
