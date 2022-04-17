package blockchain

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/jon4hz/deadshot/internal/blockchain/multicall"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/pkg/ethutils"

	"github.com/ethereum/go-ethereum/common"
)

var ErrNoContracts = errors.New("error no contracts")

// GetTokenSymbols returns a map of token addresses with their symbols as values.
func (c *Client) GetTokenInfo(contracts ...string) (map[string]*database.Token, error) {
	validContracts := make([]string, 0)
	for _, v := range contracts {
		if !ethutils.IsValidAddress(v) {
			continue
		}
		validContracts = append(validContracts, v)
	}
	if len(validContracts) == 0 {
		return nil, ErrNoContracts
	}
	x, err := c.multic.GetTokenInfo(validContracts)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*database.Token)
	for k, v := range x {
		m[k] = database.NewToken(v.Contract, v.Symbol, v.Decimals, v.Native, v.Balance)
	}
	return m, nil
}

// GetBalanceOf returns the balance of a given address in the given token or native currency if an empty string is passed.
func (c *Client) GetBalanceOf(address string, contract string) (*big.Int, error) {
	if contract == "" {
		return c.getNativeBalanceOf(address)
	}
	if ethutils.IsZeroAddress(contract) {
		return c.getNativeBalanceOf(address)
	}
	return c.getTokenBalanceOf(address, contract)
}

func (c *Client) getNativeBalanceOf(address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	return c.Client.BalanceAt(context.Background(), addr, nil)
}

func (c *Client) GetBalanceOfToken(address string, token *database.Token) (*big.Int, error) {
	if token.GetNative() {
		return c.GetBalanceOf(address, "")
	}
	return c.GetBalanceOf(address, token.GetContract())
}

// GetPairInfo returns a map of liquidity tokens with their pair info as values.
func (c *Client) GetPairInfo(contracts ...string) (map[string]*Pair, error) {
	if len(contracts) == 0 {
		return nil, ErrNoContracts
	}
	pairs, err := c.multic.GetPairInfo(contracts)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*Pair)
	for k, v := range pairs {
		m[k] = NewPair(
			k,
			v.Token0.ToDatabaseToken(),
			v.Token1.ToDatabaseToken(),
			v.Reserve0,
			v.Reserve1,
			v.TotalSupply,
		)
	}
	return m, nil
}

// GetPairTokens returns a slice of liquidity tokens.
func (c *Client) GetValidPairTokens(tokens []*database.Token, factory string) ([]string, error) {
	pairs := genPairs(tokens, factory)

	tokenPairsM := make([]multicall.TokenPair, len(pairs))
	for i, v := range pairs {
		tokenPairsM[i] = multicall.TokenPair(v)
	}

	pairTokens, err := c.multic.GetPairToken(tokenPairsM)
	if err != nil {
		return nil, err
	}

	tokenPairs := make([]string, 0)
	var i int
	for _, v := range pairTokens {
		tokenPairs = append(tokenPairs, v.String())
		i++
	}

	return tokenPairs, nil
}

func genPairs(tokens []*database.Token, factory string) []TokenPair {
	var tokenPairs []TokenPair
	for i := 0; i < len(tokens); i++ {
		for j := 0; j < len(tokens); j++ {
			if i == j {
				continue
			}
			if tokens[i] == tokens[j] {
				continue
			}
			tokenPair := NewTokenPair(tokens[i].GetContract(), tokens[j].GetContract(), factory)
			if containsPair(tokenPairs, *tokenPair) {
				continue
			}
			tokenPairs = append(tokenPairs, *tokenPair)
		}
	}
	return tokenPairs
}

func containsPair(tokenPairs []TokenPair, pair TokenPair) bool {
	for _, p := range tokenPairs {
		if strings.EqualFold(p.Token0, pair.Token0) && strings.EqualFold(p.Token1, pair.Token1) {
			return true
		}
		if strings.EqualFold(p.Token1, pair.Token0) && strings.EqualFold(p.Token0, pair.Token1) {
			return true
		}
	}
	return false
}
