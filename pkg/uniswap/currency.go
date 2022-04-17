package uniswap

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

// ErrInvalidCurrency diff currency error.
var ErrInvalidCurrency = errors.New("diff currency")

type currency interface {
	equals(string) bool
	// Address returns the address of the currency.
	Address() string
	// Decimals returns the number of the currency..
	Decimals() uint8
	// Name returns the name of the currency.
	Name() string
	// Symbol returns the symbol of the currency.
	Symbol() string
}

type baseCurrency struct {
	address  common.Address
	decimals uint8
	symbol   string
	name     string
	isNative bool
}

type NativeCurrency struct {
	baseCurrency
}

// NewNativeCurrency is the constructor for the native currency.
func NewNativeCurrency(name, symbol string, decimals uint8) *NativeCurrency {
	return &NativeCurrency{
		baseCurrency: baseCurrency{
			name:     name,
			symbol:   symbol,
			decimals: decimals,
			address:  ZeroAddr,
			isNative: true,
		},
	}
}

func (c *NativeCurrency) equals(symbol string) bool {
	return c.symbol == symbol
}

// Address is the contract address of the currency.
func (c *NativeCurrency) Address() string {
	return c.address.String()
}

// CommonAddress is the common address of the currency.
func (c *NativeCurrency) CommonAddress() common.Address {
	return c.address
}

// Decimals returns the number of decimals for the currency.
func (c *NativeCurrency) Decimals() uint8 {
	return c.decimals
}

// Name returns the name of the currency.
func (c *NativeCurrency) Name() string {
	return c.name
}

// Symbol returns the symbol of the currency.
func (c *NativeCurrency) Symbol() string {
	return c.symbol
}
