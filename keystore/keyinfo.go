package keystore

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"
)

// KeyInfo is used for storing keys in KeyStore
type KeyInfo struct {
	Type      KeyType
	SecretKey []byte
}

// create a keyinfo with a sk string
func NewKey() (*KeyInfo, error) {
	// new sk from random data
	k, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// to bytes
	b := (*btcec.PrivateKey)(k).Serialize()

	// ki
	ki := KeyInfo{
		SecretKey: b,
		Type:      Secp256k1,
	}

	return &ki, nil
}

// import a keyinfo with a sk string
func Import(sk string) (*KeyInfo, error) {
	// decode
	b, err := hex.DecodeString(sk)
	if err != nil {
		return nil, err
	}

	// ki
	ki := KeyInfo{
		SecretKey: b,
		Type:      Secp256k1,
	}

	return &ki, nil
}

// get address of the key
func (ki KeyInfo) Address() string {
	// btcec key
	_, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), ki.SecretKey)

	// ecdsa pubkey to addr
	addr := crypto.PubkeyToAddress(ecdsa.PublicKey(*pubKey))

	return addr.String()
}

// get the sk as a string
func (ki KeyInfo) SK() string {
	sk := hex.EncodeToString(ki.SecretKey)

	return sk
}
