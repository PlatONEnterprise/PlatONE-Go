package main

import (
	"path/filepath"
	"regexp"
)

var (
	account = ""
)

func genAccount(datadir string) string {
	if !filepath.IsAbs(datadir) {
		datadir = calcAbsPath(datadir)
	}
	password := calcAbsPath("./conf/account.password")

	out := RunCmd(calcAbsPath("platone"), "account", "new", "--password", password, "--datadir", datadir)

	r := regexp.MustCompile(`Address: \{(.*)\}`)
	acc := r.FindStringSubmatch(string(out))[1]
	if account == "" {
		account = acc
	}

	return account
}
