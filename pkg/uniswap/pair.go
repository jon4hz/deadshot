package uniswap

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	Decimals18  = 18
	Univ2Symbol = "UNI-V2"
	Univ2Name   = "Uniswap V2"
)

var (
	MinimumLiquidity = big.NewInt(1000)
	defaultDexFee    = big.NewInt(9970)
)

var (
	// ErrInvalidLiquidity invalid liquidity.
	ErrInvalidLiquidity = errors.New("invalid liquidity")
	// ErrInvalidKLast invalid kLast.
	ErrInvalidKLast = errors.New("invalid kLast")
	// ErrTokenAmountNil token amount is nil.
	ErrTokenAmountNil = errors.New("token amount is nil")
)

// TokenAmounts warps TokenAmount array.
type TokenAmounts [2]*TokenAmount

// Tokens warps Token array.
type Tokens [2]*Token

// NewTokenAmounts creates a TokenAmount.
func NewTokenAmounts(tokenAmountA, tokenAmountB *TokenAmount) (TokenAmounts, error) {
	if tokenAmountA == nil || tokenAmountB == nil {
		return TokenAmounts{}, ErrTokenAmountNil
	}
	ok, err := tokenAmountA.Token.SortsBefore(tokenAmountB.Token)
	if err != nil {
		return TokenAmounts{}, err
	}
	if ok {
		return TokenAmounts{tokenAmountA, tokenAmountB}, nil
	}
	return TokenAmounts{tokenAmountB, tokenAmountA}, nil
}

// Pair warps uniswap pair.
type Pair struct {
	LiquidityToken *Token
	// sorted tokens
	TokenAmounts
}

// NewPair creates Pair
// address must be looked up on-chain.
func NewPair(address common.Address, tokenAmountA, tokenAmountB *TokenAmount) (*Pair, error) {
	if tokenAmountA == nil || tokenAmountB == nil {
		return nil, ErrTokenAmountNil
	}
	tokenAmounts, err := NewTokenAmounts(tokenAmountA, tokenAmountB)
	if err != nil {
		return nil, err
	}

	pair := &Pair{
		TokenAmounts: tokenAmounts,
	}
	pair.LiquidityToken, err = NewToken(address, Univ2Symbol, Univ2Name, Decimals18)
	return pair, err
}

// InvolvesToken Returns true if the token is either token0 or token1.
func (p *Pair) InvolvesToken(token *Token) bool {
	return token.equals(p.TokenAmounts[0].Token.Address()) || token.equals(p.TokenAmounts[1].Token.Address())
}

// Token0Price Returns the current mid price of the pair in terms of token0, i.e. the ratio of reserve1 to reserve0.
func (p *Pair) Token0Price() *Price {
	return NewPrice(p.Token0(), p.Token1(), p.TokenAmounts[0].Raw(), p.TokenAmounts[1].Raw())
}

// Token1Price Returns the current mid price of the pair in terms of token1, i.e. the ratio of reserve0 to reserve1.
func (p *Pair) Token1Price() *Price {
	return NewPrice(p.Token1(), p.Token0(), p.TokenAmounts[1].Raw(), p.TokenAmounts[0].Raw())
}

// PriceOf Returns the price of the given token in terms of the other token in the pair.
// @param token token to return price of.
func (p *Pair) PriceOf(token *Token) (*Price, error) {
	if !p.InvolvesToken(token) {
		return nil, ErrDiffToken
	}

	if token.equals(p.Token0().Address()) {
		return p.Token0Price(), nil
	}
	return p.Token1Price(), nil
}

// Token0 returns the first token in the pair.
func (p *Pair) Token0() *Token {
	return p.TokenAmounts[0].Token
}

// Token1 returns the last token in the pair.
func (p *Pair) Token1() *Token {
	return p.TokenAmounts[1].Token
}

// Reserve0 returns the first TokenAmount in the pair.
func (p *Pair) Reserve0() *TokenAmount {
	return p.TokenAmounts[0]
}

// Reserve1 returns the last TokenAmount in the pair.
func (p *Pair) Reserve1() *TokenAmount {
	return p.TokenAmounts[1]
}

// ReserveOf returns the TokenAmount that equals to the token.
func (p *Pair) ReserveOf(token *Token) (*TokenAmount, error) {
	if !p.InvolvesToken(token) {
		return nil, ErrDiffToken
	}

	if token.equals(p.Token0().Address()) {
		return p.Reserve0(), nil
	}
	return p.Reserve1(), nil
}

// GetOutputAmount returns OutputAmount and a Pair for the InputAmout.
func (p *Pair) GetOutputAmount(inputAmount *TokenAmount, dexFee *big.Int) (*TokenAmount, *Pair, error) {
	if !p.InvolvesToken(inputAmount.Token) {
		return nil, nil, ErrDiffToken
	}

	if p.Reserve0().Raw().Cmp(big.NewInt(0)) == 0 ||
		p.Reserve1().Raw().Cmp(big.NewInt(0)) == 0 {
		return nil, nil, ErrInsufficientReserves
	}

	inputReserve, err := p.ReserveOf(inputAmount.Token)
	if err != nil {
		return nil, nil, err
	}
	token := p.Token0()
	if inputAmount.Token.equals(p.Token0().Address()) {
		token = p.Token1()
	}
	outputReserve, err := p.ReserveOf(token)
	if err != nil {
		return nil, nil, err
	}

	inputAmountWithFee := big.NewInt(0).Mul(inputAmount.Raw(), dexFee)
	numerator := big.NewInt(0).Mul(inputAmountWithFee, outputReserve.Raw())
	denominator := big.NewInt(0).Add(big.NewInt(0).Mul(inputReserve.Raw(), big.NewInt(10000)), inputAmountWithFee)
	outputAmount, err := NewTokenAmount(token, big.NewInt(0).Div(numerator, denominator))
	if err != nil {
		return nil, nil, err
	}
	if outputAmount.Raw().Cmp(big.NewInt(0)) == 0 {
		return nil, nil, ErrInsufficientInputAmount
	}

	tokenAmountA, err := inputAmount.add(inputReserve)
	if err != nil {
		return nil, nil, err
	}
	tokenAmountB, err := outputReserve.subtract(outputAmount)
	if err != nil {
		return nil, nil, err
	}
	addr := common.HexToAddress(p.LiquidityToken.Address())
	pair, err := NewPair(addr, tokenAmountA, tokenAmountB)
	if err != nil {
		return nil, nil, err
	}
	return outputAmount, pair, nil
}

// GetInputAmount returns InputAmout and a Pair for the OutputAmount.
func (p *Pair) GetInputAmount(outputAmount *TokenAmount, dexFee *big.Int) (*TokenAmount, *Pair, error) {
	if !p.InvolvesToken(outputAmount.Token) {
		return nil, nil, ErrDiffToken
	}

	outputReserve, err := p.ReserveOf(outputAmount.Token)
	if err != nil {
		return nil, nil, err
	}
	if p.Reserve0().Raw().Cmp(big.NewInt(0)) == 0 ||
		p.Reserve1().Raw().Cmp(big.NewInt(0)) == 0 ||
		outputAmount.Raw().Cmp(outputReserve.Raw()) >= 0 {
		return nil, nil, ErrInsufficientReserves
	}

	token := p.Token0()
	if outputAmount.Token.equals(p.Token0().Address()) {
		token = p.Token1()
	}
	inputReserve, err := p.ReserveOf(token)
	if err != nil {
		return nil, nil, err
	}

	numerator := big.NewInt(0).Mul(inputReserve.Raw(), outputAmount.Raw())
	numerator.Mul(numerator, big.NewInt(10000))
	denominator := big.NewInt(0).Sub(outputReserve.Raw(), outputAmount.Raw())
	denominator.Mul(denominator, dexFee)
	amount := big.NewInt(0).Div(numerator, denominator)
	amount.Add(amount, big.NewInt(1))
	inputAmount, err := NewTokenAmount(token, amount)
	if err != nil {
		return nil, nil, err
	}

	tokenAmountA, err := inputAmount.add(inputReserve)
	if err != nil {
		return nil, nil, err
	}
	tokenAmountB, err := outputReserve.subtract(outputAmount)
	if err != nil {
		return nil, nil, err
	}
	addr := common.HexToAddress(p.LiquidityToken.Address())
	pair, err := NewPair(addr, tokenAmountA, tokenAmountB)
	if err != nil {
		return nil, nil, err
	}
	return inputAmount, pair, nil
}

// GetLiquidityMinted returns liquidity minted TokenAmount.
func (p *Pair) GetLiquidityMinted(totalSupply, tokenAmountA, tokenAmountB *TokenAmount) (*TokenAmount, error) {
	if !p.LiquidityToken.equals(totalSupply.Token.Address()) {
		return nil, ErrDiffToken
	}

	tokenAmounts, err := NewTokenAmounts(tokenAmountA, tokenAmountB)
	if err != nil {
		return nil, err
	}
	if !(tokenAmounts[0].Token.equals(p.Token0().Address()) && tokenAmounts[1].Token.equals(p.Token1().Address())) {
		return nil, ErrDiffToken
	}

	var liquidity *big.Int
	if totalSupply.Raw().Cmp(big.NewInt(0)) == 0 {
		liquidity = big.NewInt(0).Mul(tokenAmounts[0].Raw(), tokenAmounts[1].Raw())
		liquidity.Sqrt(liquidity)
		liquidity.Sub(liquidity, MinimumLiquidity)
	} else {
		amount0 := big.NewInt(0).Mul(tokenAmounts[0].Raw(), totalSupply.Raw())
		amount0.Div(amount0, p.Reserve0().Raw())
		amount1 := big.NewInt(0).Mul(tokenAmounts[1].Raw(), totalSupply.Raw())
		amount1.Div(amount1, p.Reserve1().Raw())
		liquidity = amount0
		if liquidity.Cmp(amount1) > 0 {
			liquidity = amount1
		}
	}

	if liquidity.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrInsufficientInputAmount
	}

	return NewTokenAmount(p.LiquidityToken, liquidity)
}

// GetLiquidityValue returns liquidity value TokenAmount.
func (p *Pair) GetLiquidityValue(token *Token, totalSupply, liquidity *TokenAmount, feeOn bool, kLast *big.Int) (*TokenAmount, error) {
	if !p.InvolvesToken(token) || !p.LiquidityToken.equals(totalSupply.Token.Address()) || !p.LiquidityToken.equals(liquidity.Token.Address()) {
		return nil, ErrDiffToken
	}
	if liquidity.Raw().Cmp(totalSupply.Raw()) > 0 {
		return nil, ErrInvalidLiquidity
	}

	totalSupplyAdjusted, err := p.adjustTotalSupply(totalSupply, feeOn, kLast)
	if err != nil {
		return nil, err
	}

	tokenAmount, err := p.ReserveOf(token)
	if err != nil {
		return nil, err
	}

	amount := big.NewInt(0).Mul(liquidity.Raw(), tokenAmount.Raw())
	amount.Div(amount, totalSupplyAdjusted.Raw())
	return NewTokenAmount(token, amount)
}

func (p *Pair) adjustTotalSupply(totalSupply *TokenAmount, feeOn bool, kLast *big.Int) (*TokenAmount, error) {
	if !feeOn {
		return totalSupply, nil
	}

	if kLast == nil {
		return nil, ErrInvalidKLast
	}
	if kLast.Cmp(big.NewInt(0)) == 0 {
		return totalSupply, nil
	}

	rootK := big.NewInt(0).Mul(p.Reserve0().Raw(), p.Reserve1().Raw())
	rootK.Sqrt(rootK)
	rootKLast := big.NewInt(0).Sqrt(kLast)
	if rootK.Cmp(rootKLast) <= 0 {
		return totalSupply, nil
	}

	numerator := big.NewInt(0).Sub(rootK, rootKLast)
	numerator.Mul(numerator, totalSupply.Raw())
	denominator := big.NewInt(0).Mul(rootK, big.NewInt(5))
	denominator.Add(denominator, rootKLast)
	tokenAmount, err := NewTokenAmount(p.LiquidityToken, numerator.Div(numerator, denominator))
	if err != nil {
		return nil, err
	}
	return totalSupply.add(tokenAmount)
}
