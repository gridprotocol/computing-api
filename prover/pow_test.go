package prover

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/gridprotocol/computing-api/prover/types"

	"github.com/ethereum/go-ethereum/common"
)

func TestPOW(t *testing.T) {
	address := common.HexToAddress("0xf0f06FB91e42FB0fB62aB0a020bF1c2F7E93FA44")
	nodeID := types.NodeID{
		Address: address.Hex(),
		ID:      1,
	}
	var challenge []byte = make([]byte, 32)
	var buf = make([]byte, 8)
	for index := range challenge {
		challenge[index] = byte(rand.Int())
	}

	res, err := GeneratePOW(nodeID, challenge, 8)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)

	binary.LittleEndian.PutUint64(buf, uint64(res))

	hash := sha256.New()
	hash.Write(challenge)
	hash.Write(buf)
	t.Log(hex.EncodeToString(hash.Sum(nil)))
}

func TestRate(t *testing.T) {
	address := common.HexToAddress("0xf0f06FB91e42FB0fB62aB0a020bF1c2F7E93FA44")
	nodeID := types.NodeID{
		Address: address.Hex(),
		ID:      1,
	}
	var success = 0
	var challenge []byte = make([]byte, 32)
	for i := 0; i < 500; i++ {
		for index := range challenge {
			challenge[index] = byte(rand.Int())
		}
		res, err := GeneratePOW(nodeID, challenge, 8)
		if err != nil {
			t.Fatal(err)
		}
		if res != 0 {
			success++
		}
	}

	t.Logf("success rate %d", success)
}
