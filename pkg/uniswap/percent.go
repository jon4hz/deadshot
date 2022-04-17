package uniswap

import (
	"math/big"

	"github.com/jon4hz/deadshot/pkg/uniswap/number"

	"github.com/shopspring/decimal"
)

// Percent100 percent 100.
var Percent100 = NewFraction(big.NewInt(100), big.NewInt(1))

// Percent warps Fraction.
type Percent struct {
	*fraction
}

// NewPercent creates Percent.
func NewPercent(num, deno *big.Int) *Percent {
	return &Percent{
		fraction: NewFraction(num, deno),
	}
}

// ToSignificant format output.
func (p *Percent) ToSignificant(significantDigits uint, opt ...number.Option) string {
	return p.multiply(Percent100).ToSignificant(significantDigits, opt...)
}

// ToFixed format output.
func (p *Percent) ToFixed(decimalPlaces uint, opt ...number.Option) string {
	return p.multiply(Percent100).ToFixed(decimalPlaces, opt...)
}

func (p *Percent) Raw() *fraction {
	return p.fraction
}

func (p *Percent) Decimal() decimal.Decimal {
	return decimal.NewFromBigInt(p.numerator, 0).Div(decimal.NewFromBigInt(p.denominator, 0)).Mul(decimal.NewFromInt(100))
}
