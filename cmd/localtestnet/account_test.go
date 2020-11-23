package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_genAccount(t *testing.T) {
	got := genAccount("./")
	//t.Log("got account address:", got)
	assert.NotEqual(t, t, "", got)
}
