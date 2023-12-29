package local

import (
	"bytes"
	"computing-api/lib/auth"
	"strconv"
	"time"
)

const (
	leasePrefix    = "l"
	entrancePrefix = "e"
)

func prefixKey(key, prefix string) []byte {
	return []byte(prefix + key)
}

// func address2bytes(addr string) ([]byte, error) {
// 	if addr[:2] == "0x" {
// 		addr = addr[2:]
// 	}
// 	b, err := hex.DecodeString(addr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return b, nil
// }

func checkSignature(sig string, addr string, msg string) (bool, error) {
	sigByte, err := auth.HexDecode(sig)
	if err != nil {
		return false, err
	}
	addrByte, err := auth.HexDecode(addr)
	if err != nil {
		return false, err
	}
	hash := auth.Hash([]byte(msg))
	addrFromSig := auth.SigToAddress(hash, sigByte)
	if !bytes.Equal(addrByte, addrFromSig) {
		return false, nil
	}
	return true, nil
}

// unit: second
func checkExpire(expire string, within int64) (bool, error) {
	expireunix, err := strconv.ParseInt(expire, 10, 64)
	if err != nil {
		return false, err
	}
	now := time.Now().Unix()
	if expireunix+within > now {
		return true, nil
	}
	return false, nil
}
