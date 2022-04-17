package uniswap

import (
	"errors"
	"strings"

	"github.com/jon4hz/deadshot/pkg/ethutils"

	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrSameAddrss     = errors.New("same address")
	ErrDiffToken      = errors.New("diff token")
	ErrInvalidAddress = errors.New("invalid address")
)

type Token struct {
	baseCurrency
}

// NewToken is the constructor for a token.
func NewToken(address common.Address, name, symbol string, decimals uint8) (*Token, error) {
	if !ethutils.IsValidAddress(address) {
		return nil, ErrInvalidAddress
	}

	return &Token{
		baseCurrency: baseCurrency{
			name:     name,
			symbol:   symbol,
			decimals: decimals,
			address:  address,
			isNative: false,
		},
	}, nil
}

func (c *Token) equals(address string) bool {
	return c.Address() == address
}

// Address is the contract address of the currency.
func (c *Token) Address() string {
	return c.address.String()
}

// CommonAddress returns the common address of the token and the other token.
func (c *Token) CommonAddress() common.Address {
	return c.address
}

// Decimals returns the number of decimals for the currency.
func (c *Token) Decimals() uint8 {
	return c.decimals
}

// Name returns the name of the token.
func (c *Token) Name() string {
	return c.name
}

// Symbol returns the symbol of the token.
func (c *Token) Symbol() string {
	return c.symbol
}

// SortsBefore returns true if the address of this token sorts before the address of the other token
// other is token to compare
// throws ErrSameAddrss if the tokens have the same address.
func (t *Token) SortsBefore(other *Token) (bool, error) {
	if t.Address() == other.Address() {
		return false, ErrSameAddrss
	}

	return strings.ToLower(t.Address()) < strings.ToLower(other.Address()), nil
}
