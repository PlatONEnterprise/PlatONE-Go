package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "pltestnet"
	app.Action = testnetCommand

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
