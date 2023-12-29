package model

import (
	"encoding/json"
)

type ComputingInput struct {
	Request []byte
}

type ComputingOutput struct {
	Response []byte
}

type AuthInfo struct {
	Address string
	Sig     string
	Msg     string
}

type Lease struct {
	PublicKey string
	Contract  string
	Expire    string
	Resources Resources
}

func (l Lease) Encode() ([]byte, error) {
	return json.Marshal(l)
}

func (l *Lease) Decode(dat []byte) error {
	return json.Unmarshal(dat, l)
}

type Resources struct {
	Cpu     string
	Gpu     string
	Mem     string
	Storage string
}
