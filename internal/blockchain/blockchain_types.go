package blockchain

import (
	"math/big"

	"github.com/jon4hz/deadshot/internal/database"
)

type Pair struct {
	token0      *database.Token
	token1      *database.Token
	reserve0    *big.Int
	reserve1    *big.Int
	totalSupply *big.Int
	address     string
}

// NewPair creates a new pair of tokens.
func NewPair(address string, token0, token1 *database.Token, reserve0, reserve1, totalSupply *big.Int) *Pair {
	return &Pair{
		address:     address,
		token0:      token0,
		token1:      token1,
		reserve0:    reserve0,
		reserve1:    reserve1,
		totalSupply: totalSupply,
	}
}

type TokenPair struct {
	Factory string
	Token0  string
	Token1  string
}

// NewTokenPair creates a new token pair.
func NewTokenPair(token0, token1, factory string) *TokenPair {
	return &TokenPair{
		Factory: factory,
		Token0:  token0,
		Token1:  token1,
	}
}
