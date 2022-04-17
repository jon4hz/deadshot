package multicall

import (
	"math/big"

	"github.com/jon4hz/deadshot/internal/database"
)

type Token struct {
	Balance  *big.Int
	Contract string
	Symbol   string
	Decimals uint8
	Native   bool
}

func (t Token) ToDatabaseToken() *database.Token {
	token := database.NewToken(t.Contract, t.Symbol, t.Decimals, t.Native, t.Balance)
	return token
}

type Pair struct {
	Reserve0    *big.Int
	Reserve1    *big.Int
	TotalSupply *big.Int
	Token0      Token
	Token1      Token
}

type TokenPair struct {
	Factory string
	Token0  string
	Token1  string
}

func NewTokenPair(token0, token1, factory string) *TokenPair {
	return &TokenPair{
		Factory: factory,
		Token0:  token0,
		Token1:  token1,
	}
}
