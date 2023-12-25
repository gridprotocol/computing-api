package main

import (
	"computing-api/lib/auth"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func main() {
	// get data hash
	hash := auth.HASH([]byte(auth.Data))

	// sign
	signature, err := auth.Sign(hash, auth.SK)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("signature:", hexutil.Encode(signature))

	// verify
	pk, err := auth.GetPubKey(auth.SK)
	if err != nil {
		log.Fatal(err)
	}
	b, err := auth.Verify(signature, hash, pk)
	if err != nil {
		log.Fatal(err)
	}
	// check result
	if !b {
		fmt.Println("verify failed")
	} else {
		fmt.Println("verify passed")
	}
}
