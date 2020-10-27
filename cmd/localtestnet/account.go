package main

import (
	"os/exec"
	"regexp"
)

func genAccount(datadir string) string {
	out, err := exec.Command("./platone", "--datadir", datadir, "account new --password ./conf/account.password").Output()
	if nil != err {
		panic(err)
	}

	r := regexp.MustCompile(`Address: \{(.*)\}`)

	return r.FindStringSubmatch(string(out))[1]
}
