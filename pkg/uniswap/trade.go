package uniswap

import (
	"errors"
	"math/big"
)

type TradeType int

const (
	ExactInput TradeType = iota
	ExactOutput
)

var ErrInvalidSlippageTolerance = errors.New("invalid slippage tolerance")

// Trade Represents a trade executed against a list of pairs.
// Does not account for slippage, i.e. trades that front run this trade and move the price.
type Trade struct {
	// Route is the route of the trade, i.e. which pairs the trade goes through.
	Route *Route
	// TradeType is the type of the trade, either exact in or exact out.
	TradeType TradeType
	// inputAmount is the input amount for the trade assuming no slippage.
	inputAmount *TokenAmount
	// outputAmount is the output amount for the trade assuming no slippage.
	outputAmount *TokenAmount
	// ExecutionPrice is the price expressed in terms of output amount/input amount.
	ExecutionPrice *Price
	// NextMidPrice is the mid price after the trade executes assuming no slippage.
	NextMidPrice *Price
	// PriceImpact is the percent difference between the mid price before the trade and the trade execution price.
	PriceImpact *Percent
}

func (t *Trade) InputAmount() *TokenAmount {
	return t.inputAmount
}

func (t *Trade) OutputAmount() *TokenAmount {
	return t.outputAmount
}

// ExactIn constructs an exact in trade with the given amount in and route
// route is route of the exact in trade
// amountIn is the amount being passed in.
func ExactIn(route *Route, amountIn *TokenAmount, dexFee *big.Int) (*Trade, error) {
	return NewTrade(route, amountIn, ExactInput, dexFee)
}

// ExactOut constructs an exact out trade with the given amount out and route
// route is the route of the exact out trade
// amountOut is the amount returned by the trade.
func ExactOut(route *Route, amountOut *TokenAmount, dexFee *big.Int) (*Trade, error) {
	return NewTrade(route, amountOut, ExactOutput, dexFee)
}

// NewTrade creates a new trade.
func NewTrade(route *Route, amount *TokenAmount, tradeType TradeType, dexFee *big.Int) (*Trade, error) {
	if route == nil {
		return nil, ErrRouteNil
	}
	if amount == nil {
		return nil, ErrTokenAmountNil
	}

	amounts := make([]*TokenAmount, len(route.Path))
	nextPairs := make([]*Pair, len(route.Pairs))

	if tradeType == ExactInput {
		if !route.Input.equals(amount.Token.Address()) {
			return nil, ErrDiffToken
		}

		amounts[0] = amount
		for i := 0; i < len(route.Path)-1; i++ {
			outputAmount, nextPair, err := route.Pairs[i].GetOutputAmount(amounts[i], dexFee)
			if err != nil {
				return nil, err
			}
			amounts[i+1] = outputAmount
			nextPairs[i] = nextPair
		}
	} else {
		if !route.Output.equals(amount.Token.Address()) {
			return nil, ErrInvalidCurrency
		}
		if !route.Output.equals(amount.Token.Address()) {
			return nil, ErrDiffToken
		}

		amounts[len(amounts)-1] = amount
		for i := len(route.Path) - 1; i > 0; i-- {
			inputAmount, nextPair, err := route.Pairs[i-1].GetInputAmount(amounts[i], dexFee)
			if err != nil {
				return nil, err
			}
			amounts[i-1] = inputAmount
			nextPairs[i-1] = nextPair
		}
	}

	nextRoute, err := NewRoute(nextPairs, route.Input, nil)
	if err != nil {
		return nil, err
	}
	nextMidPrice, err := NewPriceFromRoute(nextRoute)
	if err != nil {
		return nil, err
	}
	inputAmount := amount
	if tradeType == ExactOutput {
		inputAmount = amounts[0]
	}
	outputAmount := amount
	if tradeType == ExactInput {
		outputAmount = amounts[len(amounts)-1]
	}
	price := NewPrice(inputAmount.currency, outputAmount.currency, inputAmount.Raw(), outputAmount.Raw())
	return &Trade{
		Route:          route,
		TradeType:      tradeType,
		inputAmount:    inputAmount,
		outputAmount:   outputAmount,
		ExecutionPrice: price,
		NextMidPrice:   nextMidPrice,
		PriceImpact:    computePriceImpact(route.MidPrice, inputAmount, outputAmount),
	}, nil
}

// computePriceImpact returns the percent difference between the mid price and the execution price, i.e. price impact.
// midPrice is the mid price before the trade
// inputAmount is the input amount of the trade
// outputAmount is the output amount of the trade.
func computePriceImpact(midPrice *Price, inputAmount, outputAmount *TokenAmount) *Percent {
	exactQuote := midPrice.Raw().multiply(NewFraction(inputAmount.Raw(), nil))
	slippage := exactQuote.subtract(NewFraction(outputAmount.Raw(), nil)).divide(exactQuote)
	return &Percent{
		fraction: slippage,
	}
}

// MinimumAmountOut gets the minimum amount that must be received from this trade for the given slippage tolerance
// slippageTolerance tolerance of unfavorable slippage from the execution price of this trade.
func (t *Trade) MinimumAmountOut(slippageTolerance *Percent) (*TokenAmount, error) {
	if slippageTolerance.lessThan(ZeroFraction) {
		return nil, ErrInvalidSlippageTolerance
	}

	if t.TradeType == ExactOutput {
		return t.outputAmount, nil
	}

	slippageAdjustedAmountOut := NewFraction(big.NewInt(1), nil).
		add(slippageTolerance.fraction).
		invert().
		multiply(NewFraction(t.outputAmount.Raw(), nil)).Quotient()
	return NewTokenAmount(t.outputAmount.Token, slippageAdjustedAmountOut)
}

/**
 * Get the maximum amount in that can be spent via this trade for the given slippage tolerance
 * @param slippageTolerance tolerance of unfavorable slippage from the execution price of this trade.
 */
func (t *Trade) MaximumAmountIn(slippageTolerance *Percent) (*TokenAmount, error) {
	if slippageTolerance.lessThan(ZeroFraction) {
		return nil, ErrInvalidSlippageTolerance
	}

	if t.TradeType == ExactInput {
		return t.inputAmount, nil
	}

	slippageAdjustedAmountIn := NewFraction(big.NewInt(1), nil).
		add(slippageTolerance.fraction).
		multiply(NewFraction(t.inputAmount.Raw(), nil)).Quotient()
	return NewTokenAmount(t.inputAmount.Token, slippageAdjustedAmountIn)
}
