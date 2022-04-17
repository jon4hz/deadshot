package blockchain

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/jon4hz/deadshot/internal/blockchain/abi/uniswapv2router2"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

var (
	me  = "0xaf5fFcce7CA5115e3dEeD9Be8B7a66199121cE0D"
	url = "https://polygon-rpc.com/"
	mc  = "0x8a233a018a2e123c0D96435CF99c8e65648b429F"
)

var quickswap = *database.NewDex(
	"quickswap",
	"0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff",
	"0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32",
	9970,
	true,
)

/* func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("github.com/rjeczalik/notify.(*nonrecursiveTree).dispatch"),
		goleak.IgnoreTopFunction("github.com/rjeczalik/notify.(*nonrecursiveTree).internal"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
		goleak.IgnoreTopFunction("database/sql.(*DB).connectionOpener"),
	)
} */

func TestInitClients(t *testing.T) {
	_, err := NewClient(url, mc)
	if err != nil {
		t.Error(err)
	}
}

func TestGetTokenBalance(t *testing.T) {
	c, err := NewClient(url, mc)
	if err != nil {
		t.Fatal(err)
	}

	balance, err := c.getTokenBalanceOf(me, "0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270")
	if err != nil {
		t.Error(err)
	}
	if balance.Cmp(big.NewInt(0)) == 0 {
		t.Error("Token balance is zero")
	}
}

func TestGetNativeBalance(t *testing.T) {
	c, err := NewClient(url, mc)
	if err != nil {
		t.Fatal(err)
	}

	balance, err := c.getNativeBalanceOf(me)
	if err != nil {
		t.Error(err)
	}
	if balance.Cmp(big.NewInt(0)) == 0 {
		t.Error("Token balance is zero")
	}
}

func TestBestTradeWithDirectPath(t *testing.T) {
	c, err := NewClient(url, mc)
	if err != nil {
		t.Fatal(err)
	}

	wethS := "0x7ceb23fd6bc0add59e62ac25578270cff1b9f619"
	usdcS := "0x2791bca1f2de4661ed88a30c99a7a9449aa84174"
	tokenInfos, err := c.GetTokenInfo(wethS, usdcS)
	if err != nil {
		t.Fatal(err)
	}

	x, err := c.GetBestTradeExactOut(tokenInfos[wethS], tokenInfos[usdcS], big.NewInt(1e6), &quickswap, tokens, 5, wethS)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range x.Route.Pairs {
		t.Log(v.Token0().Symbol())
		t.Log(v.Token1().Symbol())
	}

	t.Log(x.ExecutionPrice.ToSignificant(10))
	t.Log(x.NextMidPrice.ToSignificant(10))
	t.Log(x.Route.Input.Symbol())
	t.Log(x.OutputAmount().ToSignificant(10))

	// dex fee must be subtracted from price impact
	feePercent := decimal.NewFromInt(10000).Sub(decimal.New(quickswap.GetFee(), 0)).Div(decimal.NewFromInt(100))
	t.Log(x.PriceImpact.Decimal().Sub(feePercent))
}

func TestBestTradeWithoutDirectPath(t *testing.T) {
	batS := "0x3cef98bb43d732e2f285ee605a8158cde967d219"
	sushiS := "0x0b3f868e0be5597d5db7feb59e1cadbb0fdda50a"

	// Set the dex fee
	b := database.NewToken(batS, "BAT", 18, false, nil)
	s := database.NewToken(sushiS, "SUSHI", 18, false, nil)

	c, err := NewClient(url, mc)
	if err != nil {
		t.Fatal(err)
	}

	x, err := c.GetBestTradeExactOut(b, s, big.NewInt(1e18), &quickswap, tokens, 5, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range x.Route.Pairs {
		t.Log("Part Route:", v.Token0().Symbol())
		t.Log("Part Route:", v.Token1().Symbol())
	}

	t.Log("Execution Price: ", x.ExecutionPrice.ToSignificant(10))
	t.Log("Inverted Execution Price: ", x.ExecutionPrice.Invert().ToSignificant(10))
	t.Log("Next Mid Price: ", x.NextMidPrice.ToSignificant(10))
	t.Log("Input: ", x.Route.Input.Symbol())
	t.Log("Output: ", x.Route.Output.Symbol())
	t.Log("Output Amount: ", x.OutputAmount().ToSignificant(10))

	// dex fee must be subtracted from price impact
	feePercent := decimal.NewFromInt(10000).Sub(decimal.New(quickswap.GetFee(), 0)).Div(decimal.NewFromInt(100))
	t.Log("Price Impact %: ", x.PriceImpact.Decimal().Sub(feePercent))

	for _, p := range x.Route.Path {
		t.Log("Part Path:", p.Symbol())
	}

	min, err := x.MaximumAmountIn(uniswap.NewPercent(big.NewInt(1), big.NewInt(100)))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Maximum sold: ", min.ToSignificant(10))
}

func TestEthCall(t *testing.T) {
	c, err := NewClient(url, mc)
	if err != nil {
		t.Fatal(err)
	}
	a, err := abi.JSON(strings.NewReader(uniswapv2router2.Uniswapv2router2ABI))
	if err != nil {
		t.Fatal(err)
	}

	ctr := common.HexToAddress("0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff")

	wethS := "0x7ceb23fd6bc0add59e62ac25578270cff1b9f619"
	usdcS := "0x2791bca1f2de4661ed88a30c99a7a9449aa84174"
	tokenInfos, err := c.GetTokenInfo(wethS, usdcS)
	if err != nil {
		t.Fatal(err)
	}

	x, err := c.GetBestTradeExactIn(tokenInfos[usdcS], tokenInfos[wethS], big.NewInt(1e6), &quickswap, tokens, 5, wethS)
	if err != nil {
		t.Fatal(err)
	}
	d, err := a.Pack("swapExactTokensForTokensSupportingFeeOnTransferTokens", big.NewInt(1e6), big.NewInt(0), x.Route.GetAddresses(), common.HexToAddress(me), big.NewInt(time.Now().UnixNano()+100000))
	if err != nil {
		t.Fatal(err)
	}
	gasPrice, err := c.Client.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.Client.CallContract(context.Background(), ethereum.CallMsg{
		Data:     d,
		From:     common.HexToAddress(me),
		To:       &ctr,
		Gas:      1000000,
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestTxReceipt(t *testing.T) {
	c, err := NewClient(url, mc)
	if err != nil {
		t.Fatal(err)
	}
	txh := common.HexToHash("0xd3fa1ecc8c6a0d5f3605344b3d23218d5b0a4c59be8dc6c8d8b12809d8e2dcca")
	r, err := c.Client.TransactionReceipt(context.Background(), txh)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.Status)
}
