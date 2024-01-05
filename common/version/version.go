package version

import (
	"flag"
	"fmt"
)

const version = "0.2.1"

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
