package uniswap

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// nolint funlen
func TestTrade(t *testing.T) {
	token0, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000001"), "t0", "t0", 18)
	token1, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000002"), "t1", "t1", 18)
	token2, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000003"), "t2", "t2", 18)
	token3, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000004"), "t3", "t3", 18)

	tokenAmount_0_100, _ := NewTokenAmount(token0, big.NewInt(100))
	tokenAmount_0_1000, _ := NewTokenAmount(token0, big.NewInt(1000))
	tokenAmount_1_1000, _ := NewTokenAmount(token1, big.NewInt(1000))
	tokenAmount_1_1200, _ := NewTokenAmount(token1, big.NewInt(1200))
	tokenAmount_2_1000, _ := NewTokenAmount(token2, big.NewInt(1000))
	tokenAmount_2_1100, _ := NewTokenAmount(token2, big.NewInt(1100))
	tokenAmount_3_900, _ := NewTokenAmount(token3, big.NewInt(900))
	tokenAmount_3_1300, _ := NewTokenAmount(token3, big.NewInt(1300))

	pair_0_1, _ := NewPair(common.HexToAddress("0xF60dDd2C94A754f0C2bC4770fE0fFF6e526fDeF8"), tokenAmount_0_1000, tokenAmount_1_1000)
	pair_0_2, _ := NewPair(common.HexToAddress("0xD9cf6Be4BBb62f301Aa5d9a9B1929aCe9013A073"), tokenAmount_0_1000, tokenAmount_2_1100)
	pair_0_3, _ := NewPair(common.HexToAddress("0xB8d53323b877B9dc5746E4E94a1374Bd84CdC99A"), tokenAmount_0_1000, tokenAmount_3_900)
	pair_1_2, _ := NewPair(common.HexToAddress("0x2f30b99E339A0a511c133eD343C649AB9FD2AF67"), tokenAmount_1_1200, tokenAmount_2_1000)
	pair_1_3, _ := NewPair(common.HexToAddress("0x67f0046163849515942c562FebbFc8db55AffDCa"), tokenAmount_1_1200, tokenAmount_3_1300)

	// use WETH as ETHR
	tokenETHER, err := NewToken(common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), "WETH", "WETH", 18)
	if err != nil {
		t.Error(err)
	}
	tokenAmountETHER, _ := NewTokenAmount(tokenETHER, big.NewInt(100))
	tokenAmount_0_weth, _ := NewTokenAmount(tokenETHER, big.NewInt(1000))
	pair_weth_0, _ := NewPair(common.HexToAddress("0xF3283A14E3a1134Ef2af61244c2CE12045c1e85A"), tokenAmount_0_weth, tokenAmount_0_1000)

	tokenAmount_0_0, _ := NewTokenAmount(token0, big.NewInt(0))
	tokenAmount_1_0, _ := NewTokenAmount(token1, big.NewInt(0))
	empty_pair_0_1, _ := NewPair(common.HexToAddress("0xF60dDd2C94A754f0C2bC4770fE0fFF6e526fDeF8"), tokenAmount_0_0, tokenAmount_1_0)
	_ = empty_pair_0_1

	{
		route, _ := NewRoute([]*Pair{pair_weth_0}, tokenETHER, nil)
		trade, _ := NewTrade(route, tokenAmountETHER, ExactInput, defaultDexFee)

		// can be constructed with ETHER as input
		{
			expect := tokenETHER
			output := trade.inputAmount.Token
			if !expect.equals(output.Address()) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		{
			expect := token0
			output := trade.outputAmount.currency
			if !expect.equals(output.Address()) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		// can be constructed with ETHER as input for exact output
		route, _ = NewRoute([]*Pair{pair_weth_0}, tokenETHER, token0)
		trade, _ = NewTrade(route, tokenAmount_0_100, ExactOutput, defaultDexFee)
		{
			expect := tokenETHER
			output := trade.inputAmount
			if !expect.equals(output.Address()) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		{
			expect := token0
			output := trade.outputAmount
			if !expect.equals(output.Address()) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		route, _ = NewRoute([]*Pair{pair_weth_0}, token0, tokenETHER)
		// can be constructed with ETHER as output
		trade, _ = NewTrade(route, tokenAmountETHER, ExactOutput, defaultDexFee)
		{
			expect := token0
			output := trade.inputAmount
			if !expect.equals(output.Address()) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		{
			expect := tokenETHER
			output := trade.outputAmount
			if !expect.equals(output.Address()) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		// can be constructed with ETHER as output for exact input
		trade, _ = NewTrade(route, tokenAmount_0_100, ExactInput, defaultDexFee)
		{
			expect := token0
			output := trade.inputAmount
			if !expect.equals(output.Address()) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		{
			expect := tokenETHER
			output := trade.outputAmount
			if !expect.equals(output.Address()) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
	}

	// bestTradeExactIn
	{
		pairs := []*Pair{}
		_, output := BestTradeExactIn(pairs, tokenAmount_0_100, token2,
			NewDefaultBestTradeOptions(), nil, tokenAmount_0_100, nil)
		// throws with empty pairs
		{
			expect := ErrInvalidPairs
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		pairs = []*Pair{pair_0_2}
		_, output = BestTradeExactIn(pairs, tokenAmount_0_100, token2, &BestTradeOptions{},
			nil, tokenAmount_0_100, nil)
		// throws with max hops of 0
		{
			expect := ErrInvalidOption
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		pairs = []*Pair{pair_0_1, pair_0_2, pair_1_2}
		result, _ := BestTradeExactIn(pairs, tokenAmount_0_100, token2,
			NewDefaultBestTradeOptions(), nil, tokenAmount_0_100, nil)
		// provides best route
		{
			{
				tests := []struct {
					expect int
					output int
				}{
					{2, len(result)},
					{1, len(result[0].Route.Pairs)},
					{2, len(result[1].Route.Pairs)},
				}
				for i, test := range tests {
					if test.expect != test.output {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{token0, token2}, result[0].Route.Path},
					{[]*Token{token0, token1, token2}, result[1].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}

			{
				tokenAmount_2_99, _ := NewTokenAmount(token2, big.NewInt(99))
				tokenAmount_2_69, _ := NewTokenAmount(token2, big.NewInt(69))
				tests := []struct {
					expect *TokenAmount
					output *TokenAmount
				}{
					{result[0].inputAmount, tokenAmount_0_100},
					{result[0].outputAmount, tokenAmount_2_99},
					{result[1].inputAmount, tokenAmount_0_100},
					{result[1].outputAmount, tokenAmount_2_69},
				}
				for i, test := range tests {
					if !test.expect.equals(test.output) {
						t.Errorf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
		}

		// doesnt throw for zero liquidity pairs
		// throws with max hops of 0
		{
			pairs := []*Pair{empty_pair_0_1}
			results, err := BestTradeExactIn(pairs, tokenAmount_0_100, token1,
				NewDefaultBestTradeOptions(), nil, tokenAmount_0_100, nil)
			if err != nil {
				t.Fatalf("err should be nil, got[%+v]", err)
			}
			expect := 0
			output := len(results)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		tokenAmount, _ := NewTokenAmount(token0, big.NewInt(10))
		result, _ = BestTradeExactIn(pairs, tokenAmount, token2,
			&BestTradeOptions{MaxNumResults: 3, MaxHops: 1}, nil, tokenAmount, nil)
		// respects maxHops
		{
			{
				tests := []struct {
					expect int
					output int
				}{
					{1, len(result)},
					{1, len(result[0].Route.Pairs)},
				}
				for i, test := range tests {
					if test.expect != test.output {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{token0, token2}, result[0].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}
		}

		tokenAmount, _ = NewTokenAmount(token0, big.NewInt(1))
		result, _ = BestTradeExactIn(pairs, tokenAmount, token2,
			nil, nil, nil, nil)
		// insufficient input for one pair
		{
			{
				tests := []struct {
					expect int
					output int
				}{
					{1, len(result)},
					{1, len(result[0].Route.Pairs)},
				}
				for i, test := range tests {
					if test.expect != test.output {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{token0, token2}, result[0].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}
			{
				expect, _ := NewTokenAmount(token2, big.NewInt(1))
				output := result[0].outputAmount
				if !expect.equals(output) {
					t.Errorf("expect[%+v], but got[%+v]", expect, output)
				}
			}
		}

		tokenAmount, _ = NewTokenAmount(token0, big.NewInt(10))
		result, _ = BestTradeExactIn(pairs, tokenAmount, token2,
			&BestTradeOptions{MaxNumResults: 1, MaxHops: 3}, nil, nil, nil)
		// respects n
		{
			expect := 1
			output := len(result)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		pairs = []*Pair{pair_0_1, pair_0_3, pair_1_3}
		result, _ = BestTradeExactIn(pairs, tokenAmount, token2,
			&BestTradeOptions{MaxNumResults: 1, MaxHops: 3}, nil, nil, nil)
		// no path
		{
			expect := 0
			output := len(result)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		pairs = []*Pair{pair_weth_0, pair_0_1, pair_0_3, pair_1_3}
		result, _ = BestTradeExactIn(pairs, tokenAmountETHER, token3,
			nil, nil, nil, nil)
		// works for ETHER currency input
		{
			{
				expect := 2
				output := len(result)
				if expect != output {
					t.Fatalf("expect[%+v], but got[%+v]", expect, output)
				}
			}
			{
				tests := []struct {
					expect currency
					output currency
				}{
					{tokenETHER, result[0].inputAmount.currency},
					{token3, result[0].outputAmount.currency},
					{tokenETHER, result[1].inputAmount.currency},
					{token3, result[1].outputAmount.currency},
				}
				for i, test := range tests {
					if !test.expect.equals(test.output.Address()) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{tokenETHER, token0, token1, token3}, result[0].Route.Path},
					{[]*Token{tokenETHER, token0, token3}, result[1].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}
		}

		tokenAmount, _ = NewTokenAmount(token3, big.NewInt(100))
		result, _ = BestTradeExactIn(pairs, tokenAmount, tokenETHER,
			nil, nil, nil, nil)
		// works for ETHER currency output
		{
			{
				expect := 2
				output := len(result)
				if expect != output {
					t.Fatalf("expect[%+v], but got[%+v]", expect, output)
				}
			}
			{
				tests := []struct {
					expect currency
					output currency
				}{
					{token3, result[0].inputAmount.currency},
					{tokenETHER, result[0].outputAmount.currency},
					{token3, result[1].inputAmount.currency},
					{tokenETHER, result[1].outputAmount.currency},
				}
				for i, test := range tests {
					if !test.expect.equals(test.output.Address()) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{token3, token0, tokenETHER}, result[0].Route.Path},
					{[]*Token{token3, token1, token0, tokenETHER}, result[1].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}
		}
	}

	// maximumAmountIn
	{
		// tradeType = EXACT_INPUT
		route, _ := NewRoute([]*Pair{pair_0_1, pair_1_2}, token0, nil)
		exactIn, _ := ExactIn(route, tokenAmount_0_100, defaultDexFee)

		// throws if less than 0
		{
			percent := NewPercent(big.NewInt(-1), big.NewInt(100))
			_, output := exactIn.MaximumAmountIn(percent)
			expect := ErrInvalidSlippageTolerance
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		// returns exact if 0
		{
			percent := NewPercent(big.NewInt(0), big.NewInt(100))
			output, _ := exactIn.MaximumAmountIn(percent)
			expect := exactIn.inputAmount
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		// returns exact if nonzero
		{
			percent := NewPercent(big.NewInt(0), big.NewInt(100))
			output, _ := exactIn.MaximumAmountIn(percent)
			expect, _ := NewTokenAmount(token0, big.NewInt(100))
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}

			percent = NewPercent(big.NewInt(5), big.NewInt(100))
			output, _ = exactIn.MaximumAmountIn(percent)
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}

			percent = NewPercent(big.NewInt(200), big.NewInt(100))
			output, _ = exactIn.MaximumAmountIn(percent)
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		// tradeType = EXACT_OUTPUT
		tokenAmount, _ := NewTokenAmount(token2, big.NewInt(100))
		exactOut, _ := ExactOut(route, tokenAmount, defaultDexFee)

		// throws if less than 0
		{
			percent := NewPercent(big.NewInt(-1), big.NewInt(100))
			_, output := exactOut.MaximumAmountIn(percent)
			expect := ErrInvalidSlippageTolerance
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		// returns exact if 0
		{
			percent := NewPercent(big.NewInt(0), big.NewInt(100))
			output, _ := exactOut.MaximumAmountIn(percent)
			expect := exactOut.inputAmount
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		// returns slippage amount if nonzero
		{
			percent := NewPercent(big.NewInt(0), big.NewInt(100))
			output, _ := exactOut.MaximumAmountIn(percent)
			expect, _ := NewTokenAmount(token0, big.NewInt(156))
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}

			percent = NewPercent(big.NewInt(5), big.NewInt(100))
			output, _ = exactOut.MaximumAmountIn(percent)
			expect, _ = NewTokenAmount(token0, big.NewInt(163))
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}

			percent = NewPercent(big.NewInt(200), big.NewInt(100))
			output, _ = exactOut.MaximumAmountIn(percent)
			expect, _ = NewTokenAmount(token0, big.NewInt(468))
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
	}

	// #minimumAmountOut
	{
		// tradeType = EXACT_INPUT
		route, _ := NewRoute([]*Pair{pair_0_1, pair_1_2}, token0, nil)
		exactIn, _ := ExactIn(route, tokenAmount_0_100, defaultDexFee)

		// throws if less than 0
		{
			percent := NewPercent(big.NewInt(-1), big.NewInt(100))
			_, output := exactIn.MinimumAmountOut(percent)
			expect := ErrInvalidSlippageTolerance
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		// returns exact if 0
		{
			percent := NewPercent(big.NewInt(0), big.NewInt(100))
			output, _ := exactIn.MinimumAmountOut(percent)
			expect := exactIn.outputAmount
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		// returns exact if nonzero
		{
			percent := NewPercent(big.NewInt(0), big.NewInt(100))
			output, _ := exactIn.MinimumAmountOut(percent)
			expect, _ := NewTokenAmount(token2, big.NewInt(69))
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}

			percent = NewPercent(big.NewInt(5), big.NewInt(100))
			output, _ = exactIn.MinimumAmountOut(percent)
			expect, _ = NewTokenAmount(token2, big.NewInt(65))
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}

			percent = NewPercent(big.NewInt(200), big.NewInt(100))
			output, _ = exactIn.MinimumAmountOut(percent)
			expect, _ = NewTokenAmount(token2, big.NewInt(23))
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		// tradeType = EXACT_OUTPUT
		tokenAmount, _ := NewTokenAmount(token2, big.NewInt(100))
		exactOut, _ := ExactOut(route, tokenAmount, defaultDexFee)

		// throws if less than 0
		{
			percent := NewPercent(big.NewInt(-1), big.NewInt(100))
			_, output := exactOut.MinimumAmountOut(percent)
			expect := ErrInvalidSlippageTolerance
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		// returns exact if 0
		{
			percent := NewPercent(big.NewInt(0), big.NewInt(100))
			output, _ := exactOut.MinimumAmountOut(percent)
			expect := exactOut.outputAmount
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
		// returns slippage amount if nonzero
		{
			percent := NewPercent(big.NewInt(0), big.NewInt(100))
			output, _ := exactOut.MinimumAmountOut(percent)
			expect, _ := NewTokenAmount(token2, big.NewInt(100))
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}

			percent = NewPercent(big.NewInt(5), big.NewInt(100))
			output, _ = exactOut.MinimumAmountOut(percent)
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}

			percent = NewPercent(big.NewInt(200), big.NewInt(100))
			output, _ = exactOut.MinimumAmountOut(percent)
			if !expect.equals(output) {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
	}

	// #bestTradeExactOut
	{
		pairs := []*Pair{}
		tokenAmount_1_100, _ := NewTokenAmount(token1, big.NewInt(100))
		tokenAmount_2_100, _ := NewTokenAmount(token2, big.NewInt(100))
		_, output := BestTradeExactOut(pairs, token2, tokenAmount_2_100,
			nil, nil, nil, nil)
		// throws with empty pairs
		{
			expect := ErrInvalidPairs
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		pairs = []*Pair{pair_0_2}
		_, output = BestTradeExactOut(pairs, token0, tokenAmount_2_100,
			&BestTradeOptions{MaxNumResults: 3}, nil, nil, nil)
		// throws with max hops of 0
		{
			expect := ErrInvalidOption
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		pairs = []*Pair{pair_0_1, pair_0_2, pair_1_2}
		result, _ := BestTradeExactOut(pairs, token0, tokenAmount_2_100,
			nil, nil, nil, nil)
		// provides best route
		{
			{
				tests := []struct {
					expect int
					output int
				}{
					{2, len(result)},
					{1, len(result[0].Route.Pairs)},
					{2, len(result[1].Route.Pairs)},
				}
				for i, test := range tests {
					if test.expect != test.output {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{token0, token2}, result[0].Route.Path},
					{[]*Token{token0, token1, token2}, result[1].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}

			{
				tokenAmount_0_101, _ := NewTokenAmount(token0, big.NewInt(101))
				tokenAmount_0_156, _ := NewTokenAmount(token0, big.NewInt(156))
				tests := []struct {
					expect *TokenAmount
					output *TokenAmount
				}{
					{result[0].inputAmount, tokenAmount_0_101},
					{result[0].outputAmount, tokenAmount_2_100},
					{result[1].inputAmount, tokenAmount_0_156},
					{result[1].outputAmount, tokenAmount_2_100},
				}
				for i, test := range tests {
					if !test.expect.equals(test.output) {
						t.Errorf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
		}

		// doesnt throw for zero liquidity pairs
		{
			pairs := []*Pair{empty_pair_0_1}
			results, err := BestTradeExactOut(pairs, token1, tokenAmount_1_100,
				nil, nil, nil, nil)
			if err != nil {
				t.Fatalf("err should be nil, got[%+v]", err)
			}
			expect := 0
			output := len(results)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		tokenAmount, _ := NewTokenAmount(token2, big.NewInt(10))
		result, _ = BestTradeExactOut(pairs, token0, tokenAmount,
			&BestTradeOptions{MaxNumResults: 3, MaxHops: 1}, nil, nil, nil)
		// respects maxHops
		{
			{
				tests := []struct {
					expect int
					output int
				}{
					{1, len(result)},
					{1, len(result[0].Route.Pairs)},
				}
				for i, test := range tests {
					if test.expect != test.output {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{token0, token2}, result[0].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}
		}

		tokenAmount, _ = NewTokenAmount(token2, big.NewInt(1200))
		result, _ = BestTradeExactOut(pairs, token0, tokenAmount,
			nil, nil, nil, nil)
		// insufficient liquidity
		{
			expect := 0
			output := len(result)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		tokenAmount, _ = NewTokenAmount(token2, big.NewInt(1050))
		result, _ = BestTradeExactOut(pairs, token0, tokenAmount,
			nil, nil, nil, nil)
		// insufficient liquidity in one pair but not the other
		{
			expect := 1
			output := len(result)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		tokenAmount, _ = NewTokenAmount(token2, big.NewInt(10))
		result, _ = BestTradeExactOut(pairs, token0, tokenAmount,
			&BestTradeOptions{MaxNumResults: 1, MaxHops: 3}, nil, nil, nil)
		// respects n
		{
			expect := 1
			output := len(result)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		pairs = []*Pair{pair_0_1, pair_0_3, pair_1_3}
		result, _ = BestTradeExactOut(pairs, token0, tokenAmount,
			nil, nil, nil, nil)
		// no path
		{
			expect := 0
			output := len(result)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		pairs = []*Pair{pair_weth_0, pair_0_1, pair_0_3, pair_1_3}
		tokenAmount, _ = NewTokenAmount(token3, big.NewInt(100))
		result, _ = BestTradeExactOut(pairs, tokenETHER, tokenAmount,
			nil, nil, nil, nil)
		// works for ETHER currency input
		{
			{
				expect := 2
				output := len(result)
				if expect != output {
					t.Fatalf("expect[%+v], but got[%+v]", expect, output)
				}
			}
			{
				tests := []struct {
					expect currency
					output currency
				}{
					{tokenETHER, result[0].inputAmount.currency},
					{token3, result[0].outputAmount.currency},
					{tokenETHER, result[1].inputAmount.currency},
					{token3, result[1].outputAmount.currency},
				}
				for i, test := range tests {
					if !test.expect.equals(test.output.Address()) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{tokenETHER, token0, token1, token3}, result[0].Route.Path},
					{[]*Token{tokenETHER, token0, token3}, result[1].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}
		}

		tokenAmount, _ = NewTokenAmount(tokenETHER, big.NewInt(100))
		result, _ = BestTradeExactOut(pairs, token3, tokenAmount,
			nil, nil, nil, nil)
		// works for ETHER currency output
		{
			{
				expect := 2
				output := len(result)
				if expect != output {
					t.Fatalf("expect[%+v], but got[%+v]", expect, output)
				}
			}
			{
				tests := []struct {
					expect currency
					output currency
				}{
					{token3, result[0].inputAmount.currency},
					{tokenETHER, result[0].outputAmount.currency},
					{token3, result[1].inputAmount.currency},
					{tokenETHER, result[1].outputAmount.currency},
				}
				for i, test := range tests {
					if !test.expect.equals(test.output.Address()) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, test.expect, test.output)
					}
				}
			}
			{
				tests := []struct {
					expect []*Token
					output []*Token
				}{
					{[]*Token{token3, token0, tokenETHER}, result[0].Route.Path},
					{[]*Token{token3, token1, token0, tokenETHER}, result[1].Route.Path},
				}
				for i, test := range tests {
					if len(test.expect) != len(test.output) {
						t.Fatalf("test #%d: expect[%+v], but got[%+v]", i, len(test.expect), (test.output))
					}
					for j := range test.expect {
						if !test.expect[j].equals(test.output[j].Address()) {
							t.Errorf("test #%d#%d: expect[%+v], but got[%+v]", i, j, test.expect[j], test.output[j])
						}
					}
				}
			}
		}
	}
}

func TestRouteInvert(t *testing.T) {
	token0, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000001"), "t0", "t0", 18)
	token1, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000002"), "t1", "t1", 18)
	token2, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000003"), "t2", "t2", 18)
	token3, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000004"), "t3", "t3", 18)

	tokenAmount_0_100, _ := NewTokenAmount(token0, big.NewInt(100))
	tokenAmount_0_1000, _ := NewTokenAmount(token0, big.NewInt(1000))
	tokenAmount_1_1000, _ := NewTokenAmount(token1, big.NewInt(1000))
	tokenAmount_1_1200, _ := NewTokenAmount(token1, big.NewInt(1200))
	tokenAmount_2_1000, _ := NewTokenAmount(token2, big.NewInt(1000))
	// tokenAmount_2_1100, _ := NewTokenAmount(token2, big.NewInt(1100))
	tokenAmount_3_900, _ := NewTokenAmount(token3, big.NewInt(900))
	tokenAmount_3_1300, _ := NewTokenAmount(token3, big.NewInt(1300))

	pair_0_1, _ := NewPair(common.HexToAddress("0xF60dDd2C94A754f0C2bC4770fE0fFF6e526fDeF8"), tokenAmount_0_1000, tokenAmount_1_1000)
	// pair_0_2, _ := NewPair(common.HexToAddress("0xD9cf6Be4BBb62f301Aa5d9a9B1929aCe9013A073"), tokenAmount_0_1000, tokenAmount_2_1100)
	pair_0_3, _ := NewPair(common.HexToAddress("0xB8d53323b877B9dc5746E4E94a1374Bd84CdC99A"), tokenAmount_0_1000, tokenAmount_3_900)
	pair_1_2, _ := NewPair(common.HexToAddress("0x2f30b99E339A0a511c133eD343C649AB9FD2AF67"), tokenAmount_1_1200, tokenAmount_2_1000)
	pair_1_3, _ := NewPair(common.HexToAddress("0x67f0046163849515942c562FebbFc8db55AffDCa"), tokenAmount_1_1200, tokenAmount_3_1300)

	pairs := []*Pair{pair_0_1, pair_1_2, pair_0_3, pair_1_3}
	result, _ := BestTradeExactIn(pairs, tokenAmount_0_100, token2,
		NewDefaultBestTradeOptions(), nil, nil, nil)
	trade := result[0]

	t.Log(trade.Route.GetAddresses())
	for _, v := range trade.Route.Pairs {
		t.Log(v)
	}
	t.Log(trade.Route.MidPrice.ToSignificant(10))
	t.Log(trade.Route.GetAddresses())
	for _, v := range trade.Route.Pairs {
		t.Log(v)
	}
	t.Log(trade.Route.MidPrice.ToSignificant(10))
}
