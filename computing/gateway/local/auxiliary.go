package local

import (
	"encoding/hex"
)

func address2bytes(addr string) ([]byte, error) {
	if addr[:2] == "0x" {
		addr = addr[2:]
	}
	b, err := hex.DecodeString(addr)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func checkAPIkey(api_key string, pk string) bool {
	return true
}

func checkExpire(expire string) bool {
	return false
}
