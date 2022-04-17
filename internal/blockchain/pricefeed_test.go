package blockchain

import (
	"testing"
	"time"

	"github.com/jon4hz/deadshot/internal/database"
)

var (
	usdc = database.NewToken("0x2791bca1f2de4661ed88a30c99a7a9449aa84174", "USDC", 6, false, nil)
	weth = database.NewToken("0x7ceb23fd6bc0add59e62ac25578270cff1b9f619", "WETH", 18, false, nil)
)

func TestPriceFeed(t *testing.T) {
	c, err := NewClient("https://polygon-rpc.com", "0x8a233a018a2e123c0D96435CF99c8e65648b429F")
	if err != nil {
		t.Fatal(err)
	}

	p := NewPrice()
	interval := time.Millisecond * 200
	p.StartFeed(c, usdc, weth, &quickswap, nil, interval, 3, "0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270")
	p.SetHeartbeat(true)
	go func() {
		time.Sleep(time.Second * 10)
		p.Stop()
	}()

loop:
	for {
		select {
		case <-p.ctx.Done():
			break loop
		case <-p.Heartbeat:
			if err := p.GetError(); err != nil {
				t.Fatal(err)
			}
			if p.GetBuyTrade() != nil {
				t.Log(p.GetBuyTrade().ExecutionPrice.ToSignificant(60))
				t.Log(p.GetSellTrade().ExecutionPrice.ToSignificant(60))
			}
		}
	}
}
