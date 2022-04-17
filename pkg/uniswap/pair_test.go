package uniswap

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestPair(t *testing.T) {
	USDC, _ := NewToken(common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), "USD Coin", "USDC", 18)
	DAI, _ := NewToken(common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), "DAI Stablecoin", "DAI", 18)
	tokenAmountUSDC, _ := NewTokenAmount(USDC, big.NewInt(100))
	tokenAmountDAI, _ := NewTokenAmount(DAI, big.NewInt(100))
	tokenAmountUSDC101, _ := NewTokenAmount(USDC, big.NewInt(101))
	tokenAmountDAI101, _ := NewTokenAmount(DAI, big.NewInt(101))

	_, _ = tokenAmountDAI101, tokenAmountUSDC101

	{
		pairA, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountUSDC, tokenAmountDAI)
		pairB, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountDAI, tokenAmountUSDC)
		expect := DAI
		// always is the token that sorts before
		output := pairA.Token0()
		if !expect.equals(output.Address()) {
			t.Errorf("expect[%+v], but got[%+v]", expect, output)
		}
		output = pairB.Token0()
		if !expect.equals(output.Address()) {
			t.Errorf("expect[%+v], but got[%+v]", expect, output)
		}

		expect = USDC
		// always is the token that sorts after
		output = pairA.Token1()
		if !expect.equals(output.Address()) {
			t.Errorf("expect[%+v], but got[%+v]", expect, output)
		}
		output = pairB.Token1()
		if !expect.equals(output.Address()) {
			t.Errorf("expect[%+v], but got[%+v]", expect, output)
		}
	}

	{
		pairA, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountUSDC, tokenAmountDAI101)
		pairB, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountDAI101, tokenAmountUSDC)
		expect := tokenAmountDAI101
		// always comes from the token that sorts before
		output := pairA.Reserve0()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}
		output = pairB.Reserve0()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}

		expect = tokenAmountUSDC
		// always comes from the token that sorts after
		output = pairA.Reserve1()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}
		output = pairB.Reserve1()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}
	}

	{
		pairA, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountUSDC101, tokenAmountDAI)
		pairB, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountDAI, tokenAmountUSDC101)
		expect := NewPrice(DAI, USDC, big.NewInt(100), big.NewInt(101))
		// returns price of token0 in terms of token1
		output := pairA.Token0Price()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}
		output = pairB.Token0Price()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}

		expect = NewPrice(USDC, DAI, big.NewInt(101), big.NewInt(100))
		// returns price of token1 in terms of token0
		output = pairA.Token1Price()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}
		output = pairB.Token1Price()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}
	}

	{
		pair, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountUSDC101, tokenAmountDAI)
		// returns price of token in terms of other token
		expect := pair.Token0Price()
		output, _ := pair.PriceOf(tokenAmountDAI.Token)
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}

		expect = pair.Token1Price()
		output, _ = pair.PriceOf(tokenAmountUSDC101.Token)
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}

		{
			// throws if invalid token
			expect := ErrDiffToken
			weth, err := NewToken(common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), "WETH", "WETH", 18)
			if err != nil {
				t.Errorf("expect no error, but got %v", err)
			}
			_, output := pair.PriceOf(weth)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
	}

	{
		pairA, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountUSDC, tokenAmountDAI101)
		pairB, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountDAI101, tokenAmountUSDC)
		expect := tokenAmountUSDC
		// returns reserves of the given token
		output, _ := pairA.ReserveOf(USDC)
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}
		output, _ = pairB.ReserveOf(USDC)
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}

		expect = tokenAmountUSDC
		// always comes from the token that sorts after
		output = pairA.Reserve1()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}
		output = pairB.Reserve1()
		if !expect.fraction.equalTo(output.fraction) {
			t.Errorf("expect[%+v], but got[%+v]", expect.fraction, output.fraction)
		}

		{
			// throws if not in the pair
			expect := ErrDiffToken
			weth, err := NewToken(common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), "WETH", "WETH", 18)
			if err != nil {
				t.Errorf("expect no error, but got %v", err)
			}
			_, output := pairB.ReserveOf(weth)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}
	}

	{
		pairA, _ := NewPair(common.HexToAddress("0xAE461cA67B15dc8dc81CE7615e0320dA1A9aB8D5"), tokenAmountUSDC, tokenAmountDAI)

		{
			expect := true
			// involvesToken
			output := pairA.InvolvesToken(USDC)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
			output = pairA.InvolvesToken(DAI)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
			expect = false
			weth, err := NewToken(common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), "WETH", "WETH", 18)
			if err != nil {
				t.Errorf("expect no error, but got %v", err)
			}
			output = pairA.InvolvesToken(weth)
			if expect != output {
				t.Errorf("expect[%+v], but got[%+v]", expect, output)
			}
		}

		{
			tokenA, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000001"), "", "", 18)
			tokenB, _ := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000002"), "", "", 18)
			tokenAmountA, _ := NewTokenAmount(tokenA, big.NewInt(0))
			tokenAmountB, _ := NewTokenAmount(tokenB, big.NewInt(0))
			pair, _ := NewPair(common.HexToAddress("0x0000000000000000000000000000000000000003"), tokenAmountA, tokenAmountB)
			{
				tokenAmount, _ := NewTokenAmount(pair.LiquidityToken, big.NewInt(0))
				tokenAmountA, _ := NewTokenAmount(tokenA, big.NewInt(1000))
				tokenAmountB, _ := NewTokenAmount(tokenB, big.NewInt(1000))
				// getLiquidityMinted:0
				expect := ErrInsufficientInputAmount
				_, output := pair.GetLiquidityMinted(tokenAmount, tokenAmountA, tokenAmountB)
				if expect != output {
					t.Errorf("expect[%+v], but got[%+v]", expect, output)
				}

				tokenAmountA, _ = NewTokenAmount(tokenA, big.NewInt(1000000))
				tokenAmountB, _ = NewTokenAmount(tokenB, big.NewInt(1))
				_, output = pair.GetLiquidityMinted(tokenAmount, tokenAmountA, tokenAmountB)
				if expect != output {
					t.Errorf("expect[%+v], but got[%+v]", expect, output)
				}

				tokenAmountA, _ = NewTokenAmount(tokenA, big.NewInt(1001))
				tokenAmountB, _ = NewTokenAmount(tokenB, big.NewInt(1001))
				{
					expect := "1"
					liquidity, _ := pair.GetLiquidityMinted(tokenAmount, tokenAmountA, tokenAmountB)
					output := liquidity.Raw().String()
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}
			}

			// getLiquidityMinted:!0
			tokenAmountA, _ = NewTokenAmount(tokenA, big.NewInt(10000))
			tokenAmountB, _ = NewTokenAmount(tokenB, big.NewInt(10000))
			pair, _ = NewPair(common.HexToAddress("0x0000000000000000000000000000000000000003"), tokenAmountA, tokenAmountB)
			{
				tokenAmount, _ := NewTokenAmount(pair.LiquidityToken, big.NewInt(10000))
				tokenAmountA, _ = NewTokenAmount(tokenA, big.NewInt(2000))
				tokenAmountB, _ = NewTokenAmount(tokenB, big.NewInt(2000))
				expect := "2000"
				liquidity, _ := pair.GetLiquidityMinted(tokenAmount, tokenAmountA, tokenAmountB)
				output := liquidity.Raw().String()
				if expect != output {
					t.Errorf("expect[%+v], but got[%+v]", expect, output)
				}
			}

			// getLiquidityValue:!feeOn
			tokenAmountA, _ = NewTokenAmount(tokenA, big.NewInt(1000))
			tokenAmountB, _ = NewTokenAmount(tokenB, big.NewInt(1000))
			pair, _ = NewPair(common.HexToAddress("0x0000000000000000000000000000000000000003"), tokenAmountA, tokenAmountB)
			tokenAmount, _ := NewTokenAmount(pair.LiquidityToken, big.NewInt(1000))
			tokenAmount500, _ := NewTokenAmount(pair.LiquidityToken, big.NewInt(500))
			{
				liquidityValue, _ := pair.GetLiquidityValue(tokenA, tokenAmount, tokenAmount, false, nil)
				{
					expect := true
					output := liquidityValue.Token.equals(tokenA.Address())
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}
				{
					expect := "1000"
					output := liquidityValue.Raw().String()
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}

				liquidityValue, _ = pair.GetLiquidityValue(tokenA, tokenAmount, tokenAmount500, false, nil)
				// 500
				{
					expect := true
					output := liquidityValue.Token.equals(tokenA.Address())
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}
				{
					expect := "500"
					output := liquidityValue.Raw().String()
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}

				liquidityValue, _ = pair.GetLiquidityValue(tokenB, tokenAmount, tokenAmount, false, nil)
				// tokenB
				{
					expect := true
					output := liquidityValue.Token.equals(tokenB.Address())
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}
				{
					expect := "1000"
					output := liquidityValue.Raw().String()
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}
			}

			// getLiquidityValue:feeOn
			{
				liquidityValue, _ := pair.GetLiquidityValue(tokenA, tokenAmount500, tokenAmount500, true, big.NewInt(500*500))
				{
					expect := true
					output := liquidityValue.Token.equals(tokenA.Address())
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}
				{
					expect := "917" // ceiling(1000 - (500 * (1 / 6)))
					output := liquidityValue.Raw().String()
					if expect != output {
						t.Errorf("expect[%+v], but got[%+v]", expect, output)
					}
				}
			}
		}
	}
}
