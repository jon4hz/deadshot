package uniswap

import (
	"errors"
	"math/big"

	"github.com/jon4hz/deadshot/pkg/uniswap/number"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/shopspring/decimal"
)

var ErrRouteNil = errors.New("Route is nil")

type Price struct {
	*fraction
	baseCurrency  currency  // input i.e. denominator
	quoteCurrency currency  // output i.e. numerator
	scalar        *fraction // used to adjust the raw fraction w/r/t the decimals of the {base,quote}Token
}

func NewPriceFromRoute(route *Route) (*Price, error) {
	if route == nil {
		return nil, ErrRouteNil
	}
	length := len(route.Pairs)
	// NOTE: check route Pairs len?
	prices := make([]*Price, length)
	for i := range route.Pairs {
		if route.Path[i].equals(route.Pairs[i].Token0().Address()) {
			prices[i] = NewPrice(route.Pairs[i].Reserve0().currency, route.Pairs[i].Reserve1().currency,
				route.Pairs[i].Reserve0().Raw(), route.Pairs[i].Reserve1().Raw())
		} else {
			prices[i] = NewPrice(route.Pairs[i].Reserve1().currency, route.Pairs[i].Reserve0().currency,
				route.Pairs[i].Reserve1().Raw(), route.Pairs[i].Reserve0().Raw())
		}
	}

	price := prices[0]
	var err error
	for i := 1; i < length; i++ {
		price, err = price.Multiply(prices[i])
		if err != nil {
			return nil, err
		}
	}
	return price, nil
}

// NewPrice constructs a Price from two currencies and two raw values
// denominator and numerator _must_ be raw, i.e. in the native representation.
func NewPrice(baseCurrency, quoteCurrency currency, denominator, numerator *big.Int) *Price {
	return &Price{
		fraction:      NewFraction(numerator, denominator),
		baseCurrency:  baseCurrency,
		quoteCurrency: quoteCurrency,
		scalar: NewFraction(math.BigPow(10, int64(baseCurrency.Decimals())),
			math.BigPow(10, int64(quoteCurrency.Decimals()))),
	}
}

func (p *Price) Raw() *fraction {
	return NewFraction(p.fraction.numerator, p.fraction.denominator)
}

func (p *Price) Adjusted() *fraction {
	return p.fraction.multiply(p.scalar)
}

func (p *Price) Invert() *Price {
	return NewPrice(p.quoteCurrency, p.baseCurrency, p.numerator, p.denominator)
}

func (p *Price) Multiply(other *Price) (*Price, error) {
	if !p.quoteCurrency.equals(other.baseCurrency.Address()) {
		return nil, ErrInvalidCurrency
	}

	fraction := p.fraction.multiply(other.fraction)
	return NewPrice(p.baseCurrency, other.quoteCurrency, fraction.denominator, fraction.numerator), nil
}

// performs floor division on overflow.
func (p *Price) Quote(currencyAmount *CurrencyAmount) (*CurrencyAmount, error) {
	if !p.baseCurrency.equals(currencyAmount.currency.Address()) {
		return nil, ErrInvalidCurrency
	}

	return NewNative(
		currencyAmount.currency.Name(),
		currencyAmount.currency.Symbol(),
		currencyAmount.currency.Decimals(),
		p.fraction.multiply(NewFraction(currencyAmount.Raw(), nil)).Quotient(),
	)
}

func (p *Price) ToSignificant(significantDigits uint, opt ...number.Option) string {
	return p.Adjusted().ToSignificant(significantDigits, opt...)
}

func (p *Price) ToFixed(decimalPlaces uint, opt ...number.Option) string {
	return p.Adjusted().ToFixed(decimalPlaces, opt...)
}

func (p *Price) Decimal() decimal.Decimal {
	return p.Adjusted().Decimal()
}
