package multicall

import (
	"github.com/jon4hz/deadshot/internal/blockchain/multicall/calls"
	"github.com/jon4hz/deadshot/pkg/ethutils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jon4hz/web3-multicall-go/multicall"
)

const (
	tokenInfoCallCount = 2
	pairInfoCallCount  = 4
)

// GetTokenInfo returns a map of token addresses with their info as values.
func (c *Client) GetTokenInfo(contracts []string) (map[string]Token, error) {
	vcs := make(multicall.ViewCalls, len(contracts)*tokenInfoCallCount)
	for i, contract := range contracts {
		vcs[i] = calls.GetErc20SymbolCall(contract)
		vcs[i+len(contracts)] = calls.GetErc20DecimalsCall(contract)
	}

	res, err := c.call(vcs, nil)
	if err != nil {
		return nil, err
	}

	info := make(map[string]Token)

	for _, contract := range contracts {
		tmpT := Token{}
		tmpT.Symbol, err = calls.GetErc20Symbol(contract, res)
		if err != nil {
			return nil, err
		}
		tmpT.Decimals, err = calls.GetErc20Decimals(contract, res)
		if err != nil {
			return nil, err
		}
		tmpT.Contract = contract
		info[contract] = tmpT
	}

	return info, nil
}

// GetPairInfo returns a map of token addresses with their info as values.
func (c *Client) GetPairInfo(contracts []string) (map[string]*Pair, error) {
	vcs := make(multicall.ViewCalls, len(contracts)*pairInfoCallCount)
	for i, contract := range contracts {
		vcs[i] = calls.GetPairReservesCall(contract)
		vcs[i+len(contracts)] = calls.GetPairTotalSupplyCall(contract)
		vcs[i+2*len(contracts)] = calls.GetPairToken0Call(contract)
		vcs[i+3*len(contracts)] = calls.GetPairToken1Call(contract)
	}

	res, err := c.call(vcs, nil)
	if err != nil {
		return nil, err
	}

	info := make(map[string]*Pair)

	for _, contract := range contracts {
		info[contract] = new(Pair)
		info[contract].Reserve0, info[contract].Reserve1, err = calls.GetPairReserves(contract, res)
		if err != nil {
			return nil, err
		}

		info[contract].TotalSupply, err = calls.GetPairTotalSupply(contract, res)
		if err != nil {
			return nil, err
		}
		t0, err := calls.GetPairToken0(contract, res) //nolint:govet
		if err != nil {
			return nil, err
		}
		info[contract].Token0 = Token{
			Contract: t0.String(),
		}
		t1, err := calls.GetPairToken1(contract, res)
		if err != nil {
			return nil, err
		}
		info[contract].Token1 = Token{
			Contract: t1.String(),
		}
	}

	// get the token info
	vcs = vcs[:0] // reset the view calls

	for _, contract := range contracts {
		vcs = append(vcs, calls.GetPairToken0InfoCalls(contract, info[contract].Token0.Contract)...)
		vcs = append(vcs, calls.GetPairToken1InfoCalls(contract, info[contract].Token1.Contract)...)
	}

	res, err = c.call(vcs, nil)
	if err != nil {
		return nil, err
	}

	for _, contract := range contracts {
		t0S, t0D, err := calls.GetPairToken0Info(contract, res)
		if err != nil {
			return nil, err
		}
		info[contract].Token0.Symbol = t0S
		info[contract].Token0.Decimals = t0D
		t1S, t1D, err := calls.GetPairToken1Info(contract, res)
		if err != nil {
			return nil, err
		}
		info[contract].Token1.Symbol = t1S
		info[contract].Token1.Decimals = t1D
	}
	return info, nil
}

// GetPairToken returns a slice of liquidity tokens
// zero addresses are not included.
func (c *Client) GetPairToken(tokenPairs []TokenPair) (map[TokenPair]common.Address, error) {
	vcs := make(multicall.ViewCalls, len(tokenPairs))
	for i, pair := range tokenPairs {
		vcs[i] = calls.GetPairTokenCall(pair.Token0, pair.Token1, pair.Factory)
	}

	res, err := c.call(vcs, nil)
	if err != nil {
		return nil, err
	}

	pairs := make(map[TokenPair]common.Address)
	for _, pair := range tokenPairs {
		addr, err := calls.GetPairToken(pair.Token0, pair.Token1, pair.Factory, res)
		if err != nil {
			return nil, err
		}
		if ethutils.IsZeroAddress(addr) {
			continue
		}
		tokenPair := *NewTokenPair(pair.Token0, pair.Token1, pair.Factory)
		pairs[tokenPair] = addr
	}
	return pairs, nil
}
