package dao

import (
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/utils"
	"github.com/shopspring/decimal"
)

func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return utils.DecimalZeroMaxPrec
	}
	return d
}
