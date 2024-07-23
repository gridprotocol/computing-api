package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

const Version = "0.3.4"

var BuildFlag string

var VersionCmd = &cli.Command{
	Name:    "version",
	Usage:   "print computing version",
	Aliases: []string{"V"},
	Action: func(_ *cli.Context) error {
		fmt.Println(Version + "+" + BuildFlag)
		return nil
	},
}
