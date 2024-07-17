package remote

import (
	"math/big"

	"github.com/grid/contracts/go/market"
)

// calc the total value of an order
func orderValue(order market.MarketOrder) *big.Int {
	ncpu := order.R.NCPU
	ngpu := order.R.NGPU
	nmem := order.R.NMEM
	ndisk := order.R.NDISK

	pcpu := order.P.PCPU
	pgpu := order.P.PGPU
	pmem := order.P.PMEM
	pdisk := order.P.PDISK

	dur := order.Duration

	vcpu := new(big.Int).SetUint64(ncpu * pcpu)
	vgpu := new(big.Int).SetUint64(ngpu * pgpu)
	vmem := new(big.Int).SetUint64(nmem * pmem)
	vdisk := new(big.Int).SetUint64(ndisk * pdisk)

	v1 := new(big.Int).Add(vcpu, vgpu)
	v2 := new(big.Int).Add(vmem, vdisk)
	v3 := new(big.Int).Add(v1, v2)

	total := new(big.Int).Mul(v3, dur)

	return total
}

// // seconds to time
// func bigIntToTime(bigInt *big.Int) (time.Time, error) {
// 	// 假设big.Int是秒数，转换为time.Time
// 	seconds := bigInt.Int64()
// 	if seconds < int64(time.Second) || seconds > int64((^uint(0)>>1)) {
// 		return time.Time{}, fmt.Errorf("big.Int value is out of range for time.Time")
// 	}

// 	return time.Unix(seconds, 0), nil
// }
