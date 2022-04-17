package uniswap

import (
	"errors"
	"math/big"
)

var (
	ErrNilToken  = errors.New("token is nil")
	ErrNilAmount = errors.New("amount is nil")
)

type TokenAmount struct {
	*CurrencyAmount
	Token *Token
}

// amount _must_ be raw, i.e. in the native representation.
func NewTokenAmount(token *Token, amount *big.Int) (*TokenAmount, error) {
	if token == nil {
		return nil, ErrNilToken
	}
	if amount == nil {
		return nil, ErrNilAmount
	}

	currencyAmount, err := NewCurrencyAmount(token, amount)
	if err != nil {
		return nil, err
	}

	return &TokenAmount{
		Token:          token,
		CurrencyAmount: currencyAmount,
	}, nil
}

func (t *TokenAmount) add(other *TokenAmount) (*TokenAmount, error) {
	if !t.Token.equals(other.Token.Address()) {
		return nil, ErrDiffToken
	}

	return NewTokenAmount(t.Token, big.NewInt(0).Add(t.Raw(), other.Raw()))
}

func (t *TokenAmount) subtract(other *TokenAmount) (*TokenAmount, error) {
	if !t.Token.equals(other.Token.Address()) {
		return nil, ErrDiffToken
	}

	return NewTokenAmount(t.Token, big.NewInt(0).Sub(t.Raw(), other.Raw()))
}

func (t *TokenAmount) equals(other *TokenAmount) bool {
	return t.Token.equals(other.Token.Address()) && t.fraction.equalTo(other.fraction)
}
