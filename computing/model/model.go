package model

import (
	"encoding/json"
)

type ComputingInput struct {
	Prompt  string
	Options map[string]string
}

type ComputingOutput struct {
	Result string
	Extra  map[string]string
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
