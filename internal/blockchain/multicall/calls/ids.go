package calls

import (
	"fmt"
)

type id int

const (
	unknown id = iota //nolint:deadcode,unused,varcheck
	erc20Symbol
	erc20Decimals
	erc20BalanceOf
	pairToken
	pairReserves
	pairTotalSupply
	pairToken0
	pairToken1
)

func (i id) getID(contract string) string {
	return fmt.Sprintf("%d_%s", i, contract)
}

func (i id) getPairTokenInfoID(contract string, tokenX uint8) string {
	return fmt.Sprintf("%d_%s_%d", i, contract, tokenX)
}

func (i id) getPairTokenID(token0, token1 string) string {
	return fmt.Sprintf("%d_%s_%s", i, token0, token1)
}
