package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func curDir() string {
	dir, err := os.Getwd()
	if nil != err {
		panic(err)
	}
	return dir
}

func calcAbsPath(path string) string {
	return filepath.Join(curDir(), path)
}

func RunCmd(name string, errHandler *os.File, arg ...string, ) string {
	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := exec.Command(name, arg...)
	fmt.Println("cmd:", cmd.String())

	cmd.Stdout = &out
	cmd.Stderr = &stderr
	//if nil != errHandler {
	//	cmd.Stderr = errHandler
	//}

	err := cmd.Run()
	if err != nil {
		panic(err.Error() + ": " + stderr.String())
	}
	fmt.Println("Result: " + out.String())

	return out.String()
}
