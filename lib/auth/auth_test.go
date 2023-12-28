package auth

import (
	"bytes"
	"strconv"
	"testing"
	"time"
)

const (
	sk   = "e4aeceb313e4ea9f4ea5e756cf930b55ce5b14dc102955c75460b9f7e37db259"
	addr = "0x0d2897e7e3ad18df4a0571a7bacb3ffe417d3b06"
)

// data used to make signature
var data = "hello"

func TestGenerateSign(t *testing.T) {
	mydata := strconv.FormatInt(time.Now().Unix(), 10)
	hash := Hash([]byte(mydata))
	sig, err := Sign(hash, sk)
	if err != nil {
		t.Fatal(err)
	}
	sigStr := HexEncode(sig)
	t.Logf("?type=2&input=%s&addr=%s&sig=%s\n", mydata, addr, sigStr)
}

func TestSignAuth(t *testing.T) {
	// get data hash
	hash := Hash([]byte(data))

	// sign
	signature, err := Sign(hash, sk)
	if err != nil {
		t.Fatal(err)
	}
	expectedSignHex := "0x6c2366862e890b4262f8b68e0c9c44f69b82bd016a6d090692dac6cb1554277a60045fb689fd3108a8d9577fb71301b2d88b23dda6b29fc2f1a8bab3dfaf79f901"
	expectedSign, err := HexDecode(expectedSignHex)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expectedSign, signature) {
		t.Errorf("expected: %v, got: %v", expectedSign, signature)
	} else {
		t.Log("signature test ok")
	}

	gotAddr := SigToAddress(hash, signature)
	expectedAddr, err := HexDecode(addr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expectedAddr, gotAddr) {
		t.Errorf("expected: %v, got: %v", expectedAddr, gotAddr)
	} else {
		t.Log("address test ok")
	}

	// // verify
	// pk, err := GetPubKey(sk)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// b, err := Verify(signature, hash, pk)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// // check result
	// if !b {
	// 	t.Error("verify failed")
	// } else {
	// 	t.Log("verify passed")
	// }
}
