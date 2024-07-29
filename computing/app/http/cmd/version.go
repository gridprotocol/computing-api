package cmd

import (
	"fmt"

	"github.com/gridprotocol/computing-api/common/version"
	"github.com/urfave/cli/v2"
)

var VersionCmd = &cli.Command{
	Name:    "version",
	Usage:   "print computing version",
	Aliases: []string{"V"},
	Action: func(_ *cli.Context) error {
		fmt.Println(version.CurrentVersion())
		return nil
	},
}
