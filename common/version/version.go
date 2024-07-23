package version

import (
	"flag"
	"fmt"
)

const version = "0.3.3"

var BuildFlag string

func CurrentVersion() string {
	return version + "+" + BuildFlag
}

func CheckVersion() bool {
	check := flag.Bool("version", false, "print version")
	flag.Parse()

	if *check {
		fmt.Println("Current version: ", version)
		fmt.Println("Building info: ", BuildFlag)
		return true
	}
	return false
}
