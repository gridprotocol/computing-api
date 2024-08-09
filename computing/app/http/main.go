package main

import (
	"fmt"
	"os"

	"github.com/gridprotocol/computing-api/common/version"
	"github.com/gridprotocol/computing-api/computing/app/http/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	local := make([]*cli.Command, 0, 1)
	local = append(local, cmd.DaemonCmd)
	local = append(local, cmd.VersionCmd)
	local = append(local, cmd.WalletCmd)

	app := cli.App{
		Commands: local,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show application version",
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Bool("version") {
				fmt.Println(version.CurrentVersion())
			}
			return nil
		},
	}
	app.Setup()

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err) // nolint:errcheck
		os.Exit(1)
	}
}
