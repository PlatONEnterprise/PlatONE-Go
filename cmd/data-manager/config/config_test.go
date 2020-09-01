package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_initFromFile(t *testing.T) {
	initFromFile("./testconfig.toml")
	assert.Equal(t, "./data-manager.log", Config.LogConf.FilePath)
}

func Test_syncConf_RandomURL(t *testing.T) {
	initFromFile("./testconfig.toml")
	assert.Equal(t, 2, len(Config.SyncConf.urls))
}
