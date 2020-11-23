package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_curDir(t *testing.T) {
	dir := curDir()
	t.Log("cur dir:", dir)
	assert.NotEqual(t, "", dir)
}

func Test_calcAbsPath(t *testing.T) {
	abspath := calcAbsPath("platone")
	assert.Contains(t, abspath, "PlatONE-Go/cmd/localtestnet/platone")
}
