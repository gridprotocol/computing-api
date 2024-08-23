package version

import (
	"flag"
	"fmt"
)

const Version = "0.4.7"

var BuildFlag string

func CurrentVersion() string {
	return Version + "+" + BuildFlag
}

func CheckVersion() bool {
	check := flag.Bool("version", false, "print version")
	flag.Parse()

	if *check {
		fmt.Println("Current version: ", Version)
		fmt.Println("Building info: ", BuildFlag)
		return true
	}
	return false
}
