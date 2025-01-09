package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Model struct {
	Name string `json:"name"`
	ID   uint64 `json:"id"`
	Path string `json:"path"`
	GPU  string `json:"gpu"`
	MEM  string `json:"mem"`
	DISK string `json:"disk"`
}

// get a yaml's path by it's id in the list file
func GetPathByID(id string) (string, error) {
	uID, err := StringToUint64(id)
	if err != nil {
		return "", err
	}

	// read list file data
	listData, err := LoadModel("./list.json")
	if err != nil {
		return "", err
	}

	// unmarshal data into yaml structs
	var models []Model
	if err := json.Unmarshal(listData, &models); err != nil {
		panic(err)
	}

	// get path with id
	for _, yaml := range models {
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

// load model info from json
func LoadModel(pathName string) ([]byte, error) {
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

// send post to platform
func SendPost(url string) {
	// 创建 HTTP POST 请求
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("Error making HTTP POST request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	// 打印响应内容
	fmt.Printf("Response: %s\n", body)
}
