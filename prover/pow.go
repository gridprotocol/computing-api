package prover

/*
int generatePOW(char *rand, int len, int diffcult, long long *index);
#cgo LDFLAGS: -L../.. -lpow -lrt
*/
import "C"

import (
	"unsafe"

	"github.com/gridprotocol/computing-api/prover/types"

	"golang.org/x/xerrors"
)

// generate a proof with POW
func GeneratePOW(nodeID types.NodeID, rand []byte, diffcult int) (int64, error) {
	var res int64
	var prefixBuf = append(rand, nodeID.ToBytes()...)

	c_prefix := (*C.char)(unsafe.Pointer(&prefixBuf[0]))
	c_index := (*C.longlong)(unsafe.Pointer(&res))

	if C.generatePOW(c_prefix, C.int(len(prefixBuf)), C.int(diffcult), c_index) != 0 {
		return 0, xerrors.New("Unexpected Error")
	}
	return res, nil
}
