package entity

import (
	"sort"
)

var TableExchangeChipsToVipPoint = make(map[int64]int64)
var MinChipsFroExchangeVipPoint = make([]int64, 0, len(TableExchangeChipsToVipPoint))

func init() {
	TableExchangeChipsToVipPoint[7.5*1000] = 1
	TableExchangeChipsToVipPoint[15*1000] = 2
	TableExchangeChipsToVipPoint[37.5*1000] = 5
	TableExchangeChipsToVipPoint[90*1000] = 10
	TableExchangeChipsToVipPoint[180*1000] = 20
	TableExchangeChipsToVipPoint[450*1000] = 50
	TableExchangeChipsToVipPoint[900*1000] = 100
	for k := range TableExchangeChipsToVipPoint {
		MinChipsFroExchangeVipPoint = append(MinChipsFroExchangeVipPoint, k)
	}
	// sort by desc
	sort.SliceStable(MinChipsFroExchangeVipPoint, func(i, j int) bool {
		x := MinChipsFroExchangeVipPoint[i]
		y := MinChipsFroExchangeVipPoint[j]
		return x > y
	})
}

func ExchangeChipsToVipPoint(chips int64) int64 {
	vipPoint := int64(0)
	for _, minChips := range MinChipsFroExchangeVipPoint {
		if chips > minChips {
			vipPoint = TableExchangeChipsToVipPoint[minChips]
			break
		}
	}
	return vipPoint
}
