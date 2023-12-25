package auth

import (
	"bytes"
	"crypto/ecdsa"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	SK = "fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19"
)

// data used to make signature
var Data = "hello"

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

// verify signature
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
func HASH(data []byte) []byte {
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
