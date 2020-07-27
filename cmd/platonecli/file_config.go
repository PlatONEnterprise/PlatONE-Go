package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
)

// Config store the values from config.json file
type Config struct {
	Account string `json:"account"` // the address used to send the transaction
	//Gas      string `json:"gas,omitempty"` 		//future feature
	//GasPrice string `json:"gasPrice,omitempty"`
	Url      string `json:"url"`      // the ip address of the remote node
	Keystore string `json:"keystore"` // the path of the keystore file
}

var config = &Config{}

const (
	defaultConfigFilePath = "./config/config.json"
)

// configInit read values from config file
func configInit() {
	runPath := utl.GetRunningTimePath()
	configFile := runPath + defaultConfigFilePath

	// create the config folder if it is not exist
	utl.FileDirectoryInit(runPath + "./config")

	_, err := os.Stat(configFile)
	if !os.IsNotExist(err) {
		config = ParseConfigJson(configFile)
	}
}

// isConfigKeys limits the keys of the config.json
func isConfigKeys(key string) bool {
	var isMatch bool
	var m = []string{"account", "url", "keystore"}

	for _, v := range m {
		if key == v {
			return true
		}
	}

	return isMatch
}

// WriteConfigFile writes data into config.json
func WriteConfig(filePath string, config *Config) {
	// Open or create file
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		utils.Fatalf(utl.ErrOpenFileFormat, "config", err.Error())
	}
	defer file.Close()

	fileBytes, _ := json.Marshal(config)

	// write file
	_ = file.Truncate(0)
	_, err = file.Write(fileBytes)
	if err != nil {
		utils.Fatalf(utl.ErrWriteFileFormat, err.Error())
	}
}

// WriteConfigFile writes data into config.json
func WriteConfigFile(filePath, key, value string) {
	var m = make(map[string]string)

	if !isConfigKeys(key) {
		utils.Fatalf("The %s can not be written into %s", key, filePath)
	}

	// Open or create file
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		utils.Fatalf(utl.ErrOpenFileFormat, "config", err.Error())
	}
	defer file.Close()

	// Read file
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		utils.Fatalf(utl.ErrReadFileFormat, filePath, err.Error())
	}

	// file is not null
	if string(fileBytes) != "" {
		err = json.Unmarshal(fileBytes, &m)
		if err != nil {
			utils.Fatalf(utl.ErrUnmarshalBytesFormat, "config", err.Error())
		}
	}

	// update config value
	m[key] = value
	fileBytes, _ = json.Marshal(m)

	// write file
	_ = file.Truncate(0)
	_, err = file.Write(fileBytes)
	if err != nil {
		utils.Fatalf(utl.ErrWriteFileFormat, err.Error())
	}
}

// ParseConfigJson parses the data in config.json to Config object
func ParseConfigJson(configPath string) *Config {

	var config = &Config{}

	configBytes, err := utl.ParseFileToBytes(configPath)
	if err != nil {
		utils.Fatalf(utl.ErrParseFileFormat, configPath, err.Error())
	}

	if len(configBytes) == 0 {
		return config
	}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		utils.Fatalf(utl.ErrUnmarshalBytesFormat, configPath, err.Error())
	}

	// file may be modified by the user incorrectly, reset the value if it is incorrect
	if !utl.IsUrl(config.Url) {
		config.Url = ""
	}
	if !utl.IsMatch(config.Account, "address") {
		config.Account = ""
	}

	return config
}
