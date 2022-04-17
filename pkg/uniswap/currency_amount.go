package uniswap

import (
	"errors"
	"math/big"
)

var (
	// ErrInsufficientReserves doesn't have insufficient reserves.
	ErrInsufficientReserves = errors.New("doesn't have insufficient reserves")
	// ErrInsufficientInputAmount the input amount insufficient reserves.
	ErrInsufficientInputAmount = errors.New("the input amount insufficient reserves")
	// ErrNilCurrency currency is nil.
	ErrNilCurrency = errors.New("currency is nil")
)

// CurrencyAmount warps Fraction and Currency.
type CurrencyAmount struct {
	*fraction
	currency
}

// NewCurrencyAmount creates a CurrencyAmount
// amount _must_ be raw, i.e. in the native representation.
func NewCurrencyAmount(currency currency, amount *big.Int) (*CurrencyAmount, error) {
	if currency == nil {
		return nil, ErrNilCurrency
	}
	if amount == nil {
		return nil, ErrNilAmount
	}

	fraction := NewFraction(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(currency.Decimals())), nil))
	return &CurrencyAmount{
		fraction: fraction,
		currency: currency,
	}, nil
}

// Raw returns Fraction's Numerator.
func (c *CurrencyAmount) Raw() *big.Int {
	return c.fraction.numerator
}

// NewEther Helper that calls the constructor with the ETHER currency
// amount ether amount in wei.
func NewNative(name, symbol string, decimals uint8, amount *big.Int) (*CurrencyAmount, error) {
	if amount == nil {
		return nil, ErrNilAmount
	}
	native := NewNativeCurrency(name, symbol, decimals)
	return NewCurrencyAmount(
		native,
		amount,
	)
}
