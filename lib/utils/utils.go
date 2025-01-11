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

// for responsedata
type Order struct {
	ID       uint64 `json:"id"`
	User     string `json:"user"`
	Provider string `json:"provider"`
	NodeID   uint64 `json:"node_id"`
	AppName  string `json:"app_name"`
	// TotalValue      uint256  `json:"total_value"` // Go 语言中没有 uint256 类型，可以使用 big.Int
	Remain         string `json:"remain"`       // 使用 string 类型存储大整数
	Remuneration   string `json:"remuneration"` // 使用 string 类型存储大整数
	ActivateTime   uint64 `json:"activate_time"`
	LastSettleTime uint64 `json:"last_settle_time"`
	Probation      uint64 `json:"probation"`
	Duration       uint64 `json:"duration"`
	Status         uint8  `json:"status"`
}

// SendGetOrderRequest 发送 HTTP GET 请求并解析响应内容为 Order 结构体
func SendGetOrderRequest(id uint64) (*Order, error) {
	url := fmt.Sprintf("http://localhost:8002/v1/order/%d/info", id)

	fmt.Println("url:", url)

	// 发送 GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error sending GET request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	fmt.Println("reading response")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	fmt.Println("response body:", body)

	// 解析响应内容为 Order 结构体
	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	fmt.Println("order info from response body: ", order)

	return &order, nil
}
