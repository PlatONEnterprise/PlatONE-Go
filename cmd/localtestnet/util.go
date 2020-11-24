package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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

func StartCmd(name string, errHandler *os.File, arg ...string) int {
	var (
		//out    bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := exec.Command(name, arg...)
	fmt.Println("cmd:", cmd.String())

	//cmd.Stdout = &out
	cmd.Stderr = &stderr
	if nil != errHandler {
		cmd.Stderr = errHandler
	}

	err := cmd.Start()
	if err != nil {
		panic(err.Error() + ": " + stderr.String())
	}

	return cmd.Process.Pid
}

func RunCmd(name string, arg ...string) string {
	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := exec.Command(name, arg...)
	fmt.Println("cmd:", cmd.String())

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		panic(err.Error() + ": " + stderr.String())
	}
	fmt.Println("Result: " + out.String())

	return out.String()
}

func clearDataAndKillProcess(datadir string) {
	killProcessByPids(fmt.Sprintf("%s/pids", datadir))

	if err := os.RemoveAll(datadir); nil != err {
		panic(err)
	}
}

func killProcessByPids(pidFilePath string) {
	pidsStr, err := ioutil.ReadFile(pidFilePath)
	if nil != err {
		panic(err)
	}
	pids := strings.Split(string(pidsStr), "|")

	for _, pid := range pids {
		if "" == strings.TrimSpace(pid) {
			continue
		}
		p, err := strconv.Atoi(strings.Trim(pid, "|"))
		if nil != err {
			panic(err)
		}

		err = syscall.Kill(p, syscall.SIGKILL)
		if nil != err && err != syscall.ESRCH {
			panic(err)
		}
	}
}
