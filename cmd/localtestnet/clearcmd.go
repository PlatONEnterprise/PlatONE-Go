package main

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	clearCommand = cli.Command{
		Action:      clearCmd,
		Name:        "clear",
		Usage:       "kill platone process and clear data[flags]",
		Category:    "TESTNET COMMANDS",
		Description: `clear data`,
		Flags: []cli.Flag{
			DataDirFlag,
		},
	}
)

func clearCmd(ctx *cli.Context) error {
	datadir := ctx.String(DataDirFlag.Name)
	clearDataAndKillProcess(datadir)
	return nil
}
