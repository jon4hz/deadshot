package calls

import (
	"errors"
	"math/big"

	"github.com/jon4hz/web3-multicall-go/multicall"
)

var (
	ErrGettingTokenSymbols  = errors.New("error getting token symbols")
	ErrGettingTokenDecimals = errors.New("error getting token decimals")
	ErrGettingTokenBalance  = errors.New("error getting token balance")
)

// GetErc20SymbolCall is a multicall.ViewCall for getting the symbol of an erc20 token.
func GetErc20SymbolCall(contract string) multicall.ViewCall {
	return multicall.NewViewCall(
		erc20Symbol.getID(contract),
		contract,
		"symbol()(string)",
		[]any{},
	)
}

// GetErc20Symbol returns the symbol of a contract.
func GetErc20Symbol(contract string, res *multicall.Result) (string, error) {
	symbol, ok := res.Calls[erc20Symbol.getID(contract)].Decoded[0].(string)
	if !ok {
		return "", ErrGettingTokenSymbols
	}
	return symbol, nil
}

// GetErc20DecimalsCall is a multicall.ViewCall for getting the decimals of an erc20 token.
func GetErc20DecimalsCall(contract string) multicall.ViewCall {
	return multicall.NewViewCall(
		erc20Decimals.getID(contract),
		contract,
		"decimals()(uint8)",
		[]any{},
	)
}

// GetErc20Decimals returns the decimals of a contract.
func GetErc20Decimals(contract string, res *multicall.Result) (uint8, error) {
	decimals, ok := res.Calls[erc20Decimals.getID(contract)].Decoded[0].(uint8)
	if !ok {
		return 0, ErrGettingTokenDecimals
	}
	return decimals, nil
}

// GetErc20BalanceOfCall is a multicall.ViewCall for getting the balance of an erc20 token.
func GetErc20BalanceOfCall(contract, address string) multicall.ViewCall {
	return multicall.NewViewCall(
		erc20BalanceOf.getID(contract),
		contract,
		"balanceOf(address)(uint256)",
		[]any{address},
	)
}

// GetErc20BalanceOf returns the balance of a contract.
func GetErc20BalanceOf(contract string, res *multicall.Result) (*big.Int, error) {
	balance, ok := res.Calls[erc20BalanceOf.getID(contract)].Decoded[0].(*big.Int)
	if !ok {
		return nil, ErrGettingTokenBalance
	}
	return balance, nil
}
