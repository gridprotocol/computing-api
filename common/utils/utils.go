package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Yaml struct {
	Name string `json:"name"`
	ID   uint64 `json:"id"`
	Path string `json:"path"`
}

// get a yaml's path by it's id in the list file
func GetPathByID(id string) (string, error) {
	uID, err := StringToUint64(id)
	if err != nil {
		return "", err
	}

	// read list file data
	listData, err := LoadYaml("./list.json")
	if err != nil {
		return "", err
	}

	// unmarshal data into yaml structs
	var yamls []Yaml
	if err := json.Unmarshal(listData, &yamls); err != nil {
		panic(err)
	}

	// get path with id
	for _, yaml := range yamls {
		// check id
		if yaml.ID == uID {
			return yaml.Path, nil
		}
	}

	// id not found
	return "", fmt.Errorf("yaml id not found")
}

// save yaml data into pathName
func SaveYaml(encData []byte, pathName string) error {
	err := os.WriteFile(pathName, encData, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON: %v", err)
	}

	return nil
}

// load yaml  from pathName
func LoadYaml(pathName string) ([]byte, error) {
	b, err := os.ReadFile(pathName)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON: %v", err)
	}

	return b, nil
}

// string to uint64
func StringToUint64(s string) (uint64, error) {
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return u, nil
}

func Uint64ToString(u uint64) string {
	res := strconv.FormatUint(u, 10) //uint64转字符串
	return res
}
