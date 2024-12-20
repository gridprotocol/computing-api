package httpserver

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// recover address from sig and msg
func recover(sig string, msg string) string {
	// 构造消息前缀
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(msg))
	// 将前缀和消息连接起来
	prefixedMessage := fmt.Sprintf("%s%s", prefix, msg)

	// 将字符串消息转换为字节
	messageBytes := []byte(prefixedMessage)

	// 哈希消息
	hash := crypto.Keccak256Hash(messageBytes)

	// 解码签名
	signature, err := hexutil.Decode(sig)
	if err != nil {
		log.Fatal(err)
	}

	// pubkey to address
	addressbyte := SigToAddress(hash.Bytes(), signature)

	addr := hex.EncodeToString(addressbyte)

	return addr
}

func SigToAddress(hash []byte, sig []byte) []byte {
	sig[len(sig)-1] %= 27
	pub, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return nil
	}
	return crypto.PubkeyToAddress(*pub).Bytes()
}
