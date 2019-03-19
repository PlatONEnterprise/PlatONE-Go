package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func parseConfigJson(configPath string) error {
	if configPath == "" {
		dir, _ := os.Getwd()
		configPath = dir + DefaultConfigFilePath
	}

	if !filepath.IsAbs(configPath) {
		configPath, _ = filepath.Abs(configPath)
	}

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("parse config file error,%s", err.Error()))
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		panic(fmt.Errorf("parse config to json error,%s", err.Error()))
	}
	return nil
}
