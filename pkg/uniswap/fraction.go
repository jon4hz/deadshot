package uniswap

import (
	"math/big"
	"strings"

	"github.com/jon4hz/deadshot/pkg/uniswap/number"

	"github.com/shopspring/decimal"
)

// ZeroFraction zero fraction instance.
var ZeroFraction = NewFraction(big.NewInt(0), nil)

type fraction struct {
	numerator   *big.Int
	denominator *big.Int

	opts *number.Options
}

// NewFraction is the constructor for a fraction.
func NewFraction(numerator, denominator *big.Int) *fraction {
	if denominator == nil {
		denominator = big.NewInt(1)
	}

	return &fraction{
		numerator:   numerator,
		denominator: denominator,
	}
}

func (f *fraction) Decimal() decimal.Decimal {
	return decimal.NewFromBigInt(f.numerator, 0).Div(decimal.NewFromBigInt(f.denominator, 0))
}

// quotient performs floor division.
func (f *fraction) Quotient() *big.Int {
	return new(big.Int).Div(f.numerator, f.denominator)
}

// remainder remainder after floor division.
func (f *fraction) remainder() *fraction {
	return NewFraction(new(big.Int).Rem(f.numerator, f.denominator), f.denominator)
}

// invert inverts a fraction.
func (f *fraction) invert() *fraction {
	return NewFraction(f.denominator, f.numerator)
}

// Add adds two fraction and returns a new fraction.
func (f *fraction) add(other *fraction) *fraction {
	if f.denominator.Cmp(other.denominator) == 0 {
		return NewFraction(big.NewInt(0).Add(f.numerator, other.numerator), f.denominator)
	}

	return NewFraction(
		big.NewInt(0).Add(
			big.NewInt(0).Mul(f.numerator, other.denominator),
			big.NewInt(0).Mul(other.numerator, f.denominator),
		),
		big.NewInt(0).Mul(f.denominator, other.denominator),
	)
}

// subtract subtracts two fraction and returns a new fraction.
func (f *fraction) subtract(other *fraction) *fraction {
	if f.denominator.Cmp(other.denominator) == 0 {
		return NewFraction(big.NewInt(0).Sub(f.numerator, other.numerator), f.denominator)
	}

	return NewFraction(
		big.NewInt(0).Sub(
			big.NewInt(0).Mul(f.numerator, other.denominator),
			big.NewInt(0).Mul(other.numerator, f.denominator),
		),
		big.NewInt(0).Mul(f.denominator, other.denominator),
	)
}

// lessThan identifies whether the caller is less than the other.
func (f *fraction) lessThan(other *fraction) bool {
	return big.NewInt(0).Mul(f.numerator, other.denominator).
		Cmp(big.NewInt(0).Mul(other.numerator, f.denominator)) < 0
}

// equalTo identifies whether the caller is equal to the other.
func (f *fraction) equalTo(other *fraction) bool {
	return big.NewInt(0).Mul(f.numerator, other.denominator).
		Cmp(big.NewInt(0).Mul(other.numerator, f.denominator)) == 0
}

// greaterThan identifies whether the caller is greater than the other.
func (f *fraction) greaterThan(other *fraction) bool {
	return big.NewInt(0).Mul(f.numerator, other.denominator).
		Cmp(big.NewInt(0).Mul(other.numerator, f.denominator)) > 0
}

// multiply mul two fraction and returns a new fraction.
func (f *fraction) multiply(other *fraction) *fraction {
	return NewFraction(
		big.NewInt(0).Mul(f.numerator, other.numerator),
		big.NewInt(0).Mul(f.denominator, other.denominator),
	)
}

// divide mul div two fraction and returns a new fraction.
func (f *fraction) divide(other *fraction) *fraction {
	return NewFraction(
		big.NewInt(0).Mul(f.numerator, other.denominator),
		big.NewInt(0).Mul(f.denominator, other.numerator),
	)
}

// ToSignificant format output.
func (f *fraction) ToSignificant(significantDigits uint, opt ...number.Option) string {
	f.opts = number.New(number.WithGroupSeparator('\xA0'), number.WithRoundingMode(number.RoundHalfUp))
	f.opts.Apply(opt...)

	d := decimal.NewFromBigInt(f.numerator, 0).Div(decimal.NewFromBigInt(f.denominator, 0))
	if d.LessThan(decimal.New(1, 0)) {
		significantDigits += countZerosAfterDecimalPoint(d.String())
	}
	f.opts.Apply(number.WithRoundingPrecision(int(significantDigits)))
	if v, err := number.DecimalRound(d, f.opts); err == nil {
		d = v
	}
	return number.DecimalFormat(d, f.opts)
}

func countZerosAfterDecimalPoint(d string) uint {
	grp := strings.Split(d, ".")
	if len(grp) != 2 {
		return 0
	}
	for i, v := range grp[1] {
		if v != '0' {
			return uint(i)
		}
	}
	return 0
}

// ToFixed format output.
func (f *fraction) ToFixed(decimalPlaces uint, opt ...number.Option) string {
	f.opts = number.New(number.WithGroupSeparator('\xA0'), number.WithRoundingMode(number.RoundHalfUp))
	f.opts.Apply(opt...)
	f.opts.Apply(number.WithDecimalPlaces(decimalPlaces))

	d := decimal.NewFromBigInt(f.numerator, 0).Div(decimal.NewFromBigInt(f.denominator, 0))

	return number.DecimalFormat(d, f.opts)
}
