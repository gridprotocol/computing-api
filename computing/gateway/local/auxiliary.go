package local

import (
	"bytes"
	"computing-api/lib/auth"
	"fmt"
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

//	func address2bytes(addr string) ([]byte, error) {
//		if addr[:2] == "0x" {
//			addr = addr[2:]
//		}
//		b, err := hex.DecodeString(addr)
//		if err != nil {
//			return nil, err
//		}
//		return b, nil
//	}

// check eth signature(with prefix)
func checkSignature(sig string, addr string, msg string) (bool, error) {
	sigByte, err := auth.HexDecode(sig)
	if err != nil {
		return false, err
	}
	addrByte, err := auth.HexDecode(addr)
	if err != nil {
		return false, err
	}
	// eth wallet now
	hash := auth.Hash([]byte(ethWalletSign(msg)))
	addrFromSig := auth.SigToAddress(hash, sigByte)
	if !bytes.Equal(addrByte, addrFromSig) {
		return false, fmt.Errorf("signature check failed")
	}
	return true, nil
}

// unit: second
// check expire for signature
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

func ethWalletSign(msg string) string {
	return fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
}
