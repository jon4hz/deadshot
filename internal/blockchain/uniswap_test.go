package blockchain

import (
	"math/big"
	"testing"

	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/pkg/ethutils"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v2"
)

var dexFee = big.NewInt(9970)

var tokensB = []byte(`
- contract: 0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270
  symbol: WMATIC
  decimals: 18
- contract: 0x7ceb23fd6bc0add59e62ac25578270cff1b9f619
  symbol: WETH
  decimals: 18
- contract: 0x2791bca1f2de4661ed88a30c99a7a9449aa84174
  symbol: USDC
  decimals: 6
- contract: 0xc2132d05d31c914a87c6611c10748aeb04b58e8f
  symbol: USDT
  decimals: 6
- contract: 0xdab529f40e671a1d4bf91361c21bf9f0c9712ab7
  symbol: BUSD
  decimals: 18
- contract: 0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6
  symbol: WBTC
  decimals: 8
- contract: 0x8f3cf7ad23cd3cadbd9735aff958023239c6a063
  symbol: DAI
  decimals: 18`)

var tokens []*database.Token

func init() {
	err := yaml.Unmarshal(tokensB, &tokens)
	if err != nil {
		panic(err)
	}
}

func TestGeneratePairs(t *testing.T) {
	var tokens []*database.Token
	err := yaml.Unmarshal(tokensB, &tokens)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewClient(url, mc)
	if err != nil {
		t.Fatal(err)
	}

	pairs, err := c.GetValidPairTokens(tokens, "0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32")
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) == 0 {
		t.Error("Expected pairs to be not empty")
	}
}

func TestToken(t *testing.T) {
	token, err := uniswap.NewToken(common.HexToAddress("0x6e7a5FAFcec6BB1e78bAE2A1F0B612012BF14827"), uniswap.Univ2Name, uniswap.Univ2Symbol, 18)
	if err != nil {
		t.Error(err)
	}

	if token.Symbol() != uniswap.Univ2Symbol {
		t.Errorf("Expected symbol to be %s,got %s", uniswap.Univ2Symbol, token.Symbol())
	}
	if token.Address() != common.HexToAddress("0x6e7a5FAFcec6BB1e78bAE2A1F0B612012BF14827").String() {
		t.Errorf("Expected address to be %s,got %s", common.HexToAddress("0x6e7a5FAFcec6BB1e78bAE2A1F0B612012BF14827"), token.Address())
	}
	if token.Name() != uniswap.Univ2Name {
		t.Errorf("Expected name to be %s,got %s", uniswap.Univ2Name, token.Name())
	}
	if token.Decimals() != 18 {
		t.Errorf("Expected decimals to be %d,got %d", 18, token.Decimals())
	}
}

func TestRandomPrice(t *testing.T) {
	c, err := NewClient(url, mc)
	if err != nil {
		t.Fatal(err)
	}

	pairs, err := c.generatePairs(quickswap.GetFactory(), tokens...)
	if err != nil {
		t.Fatal(err)
	}

	var pair *uniswap.Pair
	var token0 *uniswap.Token
	var token1 *uniswap.Token
	for _, v := range pairs {
		pair, token0, token1, err = v.ToUniswap()
		if err != nil {
			t.Fatal(err)
		}
		break // only test the first pair
	}

	price, err := pair.PriceOf(token0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(price.ToSignificant(10))
	t.Log(token0.Symbol())
	t.Log(token0.Decimals())
	t.Log(token1.Symbol())
}

func TestRouteMidPrice(t *testing.T) {
	pairToken := common.HexToAddress("0x6e7a5FAFcec6BB1e78bAE2A1F0B612012BF14827")

	wmatic, err := uniswap.NewToken(common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270"), "wmatic", "wmatic", 18)
	if err != nil {
		t.Error(err)
	}
	usdc, err := uniswap.NewToken(common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"), "usdc", "usdc", 6)
	if err != nil {
		t.Error(err)
	}
	amountA, ok := new(big.Int).SetString("13130776391388495358817242", 10)
	if !ok {
		t.Error("Failed to parse amount")
	}
	tokenAmountA, err := uniswap.NewTokenAmount(wmatic, amountA)
	if err != nil {
		t.Error(err)
	}
	amountB, ok := new(big.Int).SetString("15111050172890", 10)
	if !ok {
		t.Error("Failed to parse amount")
	}
	tokenAmountB, err := uniswap.NewTokenAmount(usdc, amountB)
	if err != nil {
		t.Error(err)
	}
	pair, err := uniswap.NewPair(pairToken, tokenAmountA, tokenAmountB)
	if err != nil {
		t.Error(err)
	}
	route, err := uniswap.NewRoute([]*uniswap.Pair{pair}, wmatic, usdc)
	if err != nil {
		t.Error(err)
	}
	if route.MidPrice.ToSignificant(10) != "1.1508116293" {
		t.Errorf("Expected price to be %s,got %s", "1.1508116293", route.MidPrice.ToSignificant(10))
	}
}

func TestExecutionPrice(t *testing.T) {
	trade := genTrade(t)
	t.Log(trade.ExecutionPrice.ToSignificant(10))
	t.Log(ethutils.ToWei(trade.ExecutionPrice.Decimal(), 6))
	t.Log(trade.ExecutionPrice.Decimal())
}

func genTrade(t *testing.T) *uniswap.Trade {
	pairToken := common.HexToAddress("0x6e7a5FAFcec6BB1e78bAE2A1F0B612012BF14827")
	wmatic, err := uniswap.NewToken(common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270"), "wmatic", "wmatic", 18)
	if err != nil {
		t.Error(err)
	}
	usdc, err := uniswap.NewToken(common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"), "usdc", "usdc", 6)
	if err != nil {
		t.Error(err)
	}
	amountA, ok := new(big.Int).SetString("13130776391388495358817242", 10)
	if !ok {
		t.Error("Failed to parse amount")
	}
	tokenAmountA, err := uniswap.NewTokenAmount(wmatic, amountA)
	if err != nil {
		t.Error(err)
	}
	amountB, ok := new(big.Int).SetString("15111050172890", 10)
	if !ok {
		t.Error("Failed to parse amount")
	}
	tokenAmountB, err := uniswap.NewTokenAmount(usdc, amountB)
	if err != nil {
		t.Error(err)
	}
	pair, err := uniswap.NewPair(pairToken, tokenAmountA, tokenAmountB)
	if err != nil {
		t.Error(err)
	}
	route, err := uniswap.NewRoute([]*uniswap.Pair{pair}, wmatic, usdc)
	if err != nil {
		t.Error(err)
	}
	amountIn, err := uniswap.NewTokenAmount(wmatic, ethutils.ToWei(1, 18))
	if err != nil {
		t.Error(err)
	}
	trade, err := uniswap.NewTrade(route, amountIn, uniswap.ExactInput, dexFee)
	if err != nil {
		t.Error(err)
	}
	return trade
}

func TestPriceChange(t *testing.T) {
	x := decimal.NewFromBigInt(big.NewInt(200), 0)
	y := decimal.NewFromBigInt(big.NewInt(100), 0)
	t.Log(y.Sub(x).Div(x).Mul(decimal.New(100, 0)))
}

func TestPriceDifference(t *testing.T) {
	x := decimal.NewFromBigInt(big.NewInt(1000), 0)
	y := decimal.NewFromBigInt(big.NewInt(100), 0)
	a := x.Sub(y).Abs()
	b := x.Add(y).Div(decimal.NewFromInt(2))
	t.Log(a.Div(b).Mul(decimal.NewFromInt(100)))
}

func TestSlippage(t *testing.T) {
	trade := genTrade(t)
	percent := uniswap.NewPercent(big.NewInt(1), big.NewInt(100))
	t.Log(percent.Decimal())
	a, err := trade.MinimumAmountOut(percent)
	if err != nil {
		t.Error(err)
	}
	t.Log(a.ToSignificant(10))
}
