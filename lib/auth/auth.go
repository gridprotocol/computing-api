package auth

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// signature to address(hex string)
func SigToAddress(hash []byte, sig []byte) []byte {
	sig[len(sig)-1] %= 27
	pub, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return nil
	}
	return crypto.PubkeyToAddress(*pub).Bytes()
}

// Decode decodes a hex string with 0x prefix.
func HexDecode(input string) ([]byte, error) {
	return hexutil.Decode(input)
}

func HexEncode(input []byte) string {
	return hexutil.Encode(input)
}

// sign on data hash
func Sign(hash []byte, sk string) ([]byte, error) {
	// string to pk
	privateKey, err := crypto.HexToECDSA(sk)
	if err != nil {
		log.Fatal(err)
	}

	// sign
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verify signature with Pubkey
func Verify(signature []byte, hash []byte, publicKeyBytes []byte) (bool, error) {
	// recover signature into pubkey
	sigPublicKey, err := crypto.Ecrecover(hash, signature)
	if err != nil {
		return false, err
	}

	// check pubkey
	matches := bytes.Equal(sigPublicKey, publicKeyBytes)

	return matches, nil
}

// data to hash
func Hash(data []byte) []byte {
	h := crypto.Keccak256Hash(data)
	return h.Bytes()
}

// get pubkey bytes from privatekey string
func GetPubKey(sk string) ([]byte, error) {
	// get privatekeyECDSA
	privateKeyECDSA, err := crypto.HexToECDSA(sk)
	if err != nil {
		return nil, err
	}

	// get pubkey from privatekey
	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	// pubkey
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	return publicKeyBytes, nil
}

// enclose the msg with eth message for sign
func EncloseEth(msg string) string {
	return fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
}
