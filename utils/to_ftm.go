package utils

import "math/big"

// ToVlry number of VLRY to Wei
func ToVlry(vlry uint64) *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(vlry), big.NewInt(1e18))
}
