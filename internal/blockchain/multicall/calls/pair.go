package calls

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jon4hz/web3-multicall-go/multicall"
)

var (
	ErrGettingPairReserves    = errors.New("error getting pair reserves")
	ErrGettingPairTotalSupply = errors.New("error getting pair total supply")
	ErrGettingPairToken0      = errors.New("error getting pair token0")
	ErrGettingPairToken1      = errors.New("error getting pair token1")
	ErrGettingPairToken       = errors.New("error getting pair token")
)

// GetPairReservesCall is a multicall.Viewcall to get the pair reserves.
func GetPairReservesCall(contract string) multicall.ViewCall {
	return multicall.NewViewCall(
		pairReserves.getID(contract),
		contract,
		"getReserves()(uint256,uint256,uint32)",
		[]any{},
	)
}

// GetPairReserves returns the pair reserves.
func GetPairReserves(contract string, res *multicall.Result) (*big.Int, *big.Int, error) {
	reserve0, ok := res.Calls[pairReserves.getID(contract)].Decoded[0].(*big.Int)
	if !ok {
		return nil, nil, ErrGettingPairReserves
	}
	reserve1, ok := res.Calls[pairReserves.getID(contract)].Decoded[1].(*big.Int)
	if !ok {
		return nil, nil, ErrGettingPairReserves
	}
	return reserve0, reserve1, nil
}

// GetPairTotalSupplyCall is a multicall.Viewcall to get the pair total supply.
func GetPairTotalSupplyCall(contract string) multicall.ViewCall {
	return multicall.NewViewCall(
		pairTotalSupply.getID(contract),
		contract,
		"totalSupply()(uint256)",
		[]any{},
	)
}

// GetPairTotalSupply returns the pair total supply.
func GetPairTotalSupply(contract string, res *multicall.Result) (*big.Int, error) {
	totalSupply, ok := res.Calls[pairTotalSupply.getID(contract)].Decoded[0].(*big.Int)
	if !ok {
		return nil, ErrGettingPairTotalSupply
	}
	return totalSupply, nil
}

// GetPairToken0Call is a multicall.Viewcall to get the pair token0.
func GetPairToken0Call(contract string) multicall.ViewCall {
	return multicall.NewViewCall(
		pairToken0.getID(contract),
		contract,
		"token0()(address)",
		[]any{},
	)
}

// GetPairToken0 returns the pair token0.
func GetPairToken0(contract string, res *multicall.Result) (common.Address, error) {
	token0, ok := res.Calls[pairToken0.getID(contract)].Decoded[0].(common.Address)
	if !ok {
		return common.Address{}, ErrGettingPairToken0
	}
	return token0, nil
}

// GetPairToken1Call is a multicall.Viewcall to get the pair token1.
func GetPairToken1Call(contract string) multicall.ViewCall {
	return multicall.NewViewCall(
		pairToken1.getID(contract),
		contract,
		"token1()(address)",
		[]any{},
	)
}

// GetPairToken1 returns the pair token1.
func GetPairToken1(contract string, res *multicall.Result) (common.Address, error) {
	token1, ok := res.Calls[pairToken1.getID(contract)].Decoded[0].(common.Address)
	if !ok {
		return common.Address{}, ErrGettingPairToken1
	}
	return token1, nil
}

const pairCallsCount = 2

// GetPairToken0InfoCalls is a multicall.Viewcall to get the pair token0 info.
func GetPairToken0InfoCalls(contract, token string) []multicall.ViewCall {
	vcs := make([]multicall.ViewCall, pairCallsCount)
	vcs[0] = multicall.NewViewCall(
		erc20Symbol.getPairTokenInfoID(contract, 0),
		token,
		"symbol()(string)",
		[]any{},
	)
	vcs[1] = multicall.NewViewCall(
		erc20Decimals.getPairTokenInfoID(contract, 0),
		token,
		"decimals()(uint8)",
		[]any{},
	)
	return vcs
}

// GetPairToken0Info returns the pair token0 info.
func GetPairToken0Info(contract string, res *multicall.Result) (string, uint8, error) {
	symbol, ok := res.Calls[erc20Symbol.getPairTokenInfoID(contract, 0)].Decoded[0].(string)
	if !ok {
		return "", 0, ErrGettingPairToken0
	}
	decimals, ok := res.Calls[erc20Decimals.getPairTokenInfoID(contract, 0)].Decoded[0].(uint8)
	if !ok {
		return "", 0, ErrGettingPairToken0
	}
	return symbol, decimals, nil
}

// GetPairToken1InfoCalls is a multicall.Viewcall to get the pair token1 info.
func GetPairToken1InfoCalls(contract, token string) []multicall.ViewCall {
	vcs := make([]multicall.ViewCall, pairCallsCount)
	vcs[0] = multicall.NewViewCall(
		erc20Symbol.getPairTokenInfoID(contract, 1),
		token,
		"symbol()(string)",
		[]any{},
	)
	vcs[1] = multicall.NewViewCall(
		erc20Decimals.getPairTokenInfoID(contract, 1),
		token,
		"decimals()(uint8)",
		[]any{},
	)
	return vcs
}

// GetPairToken1Info returns the pair token1 info.
func GetPairToken1Info(contract string, res *multicall.Result) (string, uint8, error) {
	symbol, ok := res.Calls[erc20Symbol.getPairTokenInfoID(contract, 1)].Decoded[0].(string)
	if !ok {
		return "", 0, ErrGettingPairToken1
	}
	decimals, ok := res.Calls[erc20Decimals.getPairTokenInfoID(contract, 1)].Decoded[0].(uint8)
	if !ok {
		return "", 0, ErrGettingPairToken1
	}
	return symbol, decimals, nil
}

// GetPairTokenCall is a multicall.Viewcall to get the pair token.
func GetPairTokenCall(token0, token1, factory string) multicall.ViewCall {
	return multicall.NewViewCall(
		pairToken.getPairTokenID(token0, token1),
		factory,
		"getPair(address,address)(address)",
		[]any{token0, token1},
	)
}

// GetPairToken returns the pair token.
func GetPairToken(token0, token1, factory string, res *multicall.Result) (common.Address, error) {
	pair, ok := res.Calls[pairToken.getPairTokenID(token0, token1)].Decoded[0].(common.Address)
	if !ok {
		return common.Address{}, ErrGettingPairToken
	}
	return pair, nil
}
