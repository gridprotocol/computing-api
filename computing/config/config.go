package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

var conf *GatewayConfig

type GatewayConfig struct {
	Grpc   Grpc
	Http   Http
	Local  Local
	Remote Remote
}

type Local struct {
	DBPath string
}

type Remote struct {
	Wallet string
}

type Grpc struct {
	Listen string
}

type Http struct {
	Listen string
}

func InitConfig() error {
	currentDir, _ := os.Getwd()
	configFile := filepath.Join(currentDir, "config.toml")

	if metaData, err := toml.DecodeFile(configFile, &conf); err != nil {
		return fmt.Errorf("failed load config file, path: %s, error: %w", configFile, err)
	} else {
		if !requiredFieldsAreGiven(metaData) {
			log.Fatal("Required fields not given")
		}
	}
	return nil
}

func GetConfig() *GatewayConfig {
	return conf
}

func requiredFieldsAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"Grpc"},
		{"Local"},
		{"Remote"},

		{"Grpc", "Listen"},
		{"Http", "Listen"},

		{"Local", "DBPath"},

		{"Remote", "Wallet"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("Required fields ", v)
		}
	}

	return true
}
