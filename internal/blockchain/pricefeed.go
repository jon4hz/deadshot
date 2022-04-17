package blockchain

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/pkg/ethutils"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/shopspring/decimal"
)

const (
	defaultPriceFetchInterval = time.Millisecond * 200
)

var (
	ErrNoBuyTrade  = errors.New("no buy trade found")
	ErrNoSellTrade = errors.New("no sell trade found")
)

type Price struct {
	err              error
	Heartbeat        chan struct{}
	ctx              context.Context
	cancel           context.CancelFunc
	heartbeatRunning bool
	feedRunning      bool
	buyTrade         *uniswap.Trade
	sellTrade        *uniswap.Trade
	buyAmount        *big.Int
	sellAmount       *big.Int
	mu               sync.Mutex
}

type PriceResult struct {
	buyTrade  *uniswap.Trade
	sellTrade *uniswap.Trade
	err       error
}

// New is the constructor for the PriceFeed.
func NewPrice() *Price {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	return NewPriceWithContext(ctx, cancel)
}

func NewPriceWithContext(ctx context.Context, cancel context.CancelFunc) *Price {
	return &Price{
		ctx:       ctx,
		cancel:    cancel,
		Heartbeat: make(chan struct{}, 1),
	}
}

func (p *Price) BuyTrade() (*uniswap.Trade, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buyTrade, p.err
}

func (p *Price) SellTrade() (*uniswap.Trade, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sellTrade, p.err
}

func (p *Price) SetHeartbeat(running bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.heartbeatRunning = running
}

func (p *Price) GetHeartbeatRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.heartbeatRunning
}

func (p *Price) setRunning(running bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.feedRunning = running
}

func (p *Price) GetRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.feedRunning
}

func (p *Price) SetPriceResult(r PriceResult) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var heartbeat bool
	if r.buyTrade != nil { // this prevents the price from being set to nil
		// send a heartbeat if the price has changed and the heartbeat is running
		if p.buyTrade != nil && !p.buyTrade.ExecutionPrice.Decimal().Equal(r.buyTrade.ExecutionPrice.Decimal()) {
			heartbeat = true
		} else if p.buyTrade == nil && r.buyTrade != nil { // first time the price is set
			heartbeat = true
		}
		p.buyTrade = r.buyTrade
	}
	if r.sellTrade != nil {
		if p.sellTrade != nil && !p.sellTrade.ExecutionPrice.Decimal().Equal(r.sellTrade.ExecutionPrice.Decimal()) {
			heartbeat = true
		} else if p.sellTrade == nil && r.sellTrade != nil { // first time the price is set
			heartbeat = true
		}
		p.sellTrade = r.sellTrade
	}
	p.err = r.err
	if heartbeat && r.err != nil {
		p.sendHeartbeat()
	}
}

func (p *Price) sendHeartbeat() {
	if p.heartbeatRunning {
		p.Heartbeat <- struct{}{}
	}
}

func (p *Price) GetTrades() (*uniswap.Trade, *uniswap.Trade) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buyTrade, p.sellTrade
}

func (p *Price) GetBuyTrade() *uniswap.Trade {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buyTrade
}

func (p *Price) GetSellTrade() *uniswap.Trade {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sellTrade
}

func (p *Price) GetError() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

func (p *Price) GetBuyAmount() *big.Int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buyAmount
}

func (p *Price) SetBuyAmount(amount *big.Int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buyAmount = amount
}

func (p *Price) GetSellAmount() *big.Int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sellAmount
}

func (p *Price) SetSellAmount(amount *big.Int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sellAmount = amount
}

// StartFeed starts a new price feed for the given token.
func (p *Price) StartFeed(c *Client, token0, token1 *database.Token, dex *database.Dex, tokens []*database.Token, interval time.Duration, maxHops int, weth string) {
	p.setRunning(true)
	if interval == 0 {
		interval = defaultPriceFetchInterval
	}

	go func() {
		defer func() {
			p.setRunning(false)
		}()
		ticker := time.NewTicker(interval)

		res := p.fetchPrice(c, token0, token1, dex, maxHops, weth, tokens...)
		p.SetPriceResult(res)

		for {
			select {
			case <-ticker.C:
				res := p.fetchPrice(c, token0, token1, dex, maxHops, weth, tokens...)
				p.SetPriceResult(res)
			case <-p.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *Price) Stop() {
	defer func() {
		recover() // prevent panic if the price feed is already stopped
	}()

	p.SetHeartbeat(false)
	p.setRunning(false)
	p.cancel()
}

func (p *Price) fetchPrice(c *Client, token0, token1 *database.Token, dex *database.Dex, maxHops int, weth string, tokens ...*database.Token) PriceResult {
	buyAmount := p.GetBuyAmount()
	if buyAmount == nil {
		buyAmount = ethutils.ToWei(1, token0.GetDecimals())
	}
	sellAmount := p.GetSellAmount()
	if sellAmount == nil {
		sellAmount = ethutils.ToWei(1, token1.GetDecimals())
	}
	buy, sell, err := c.GetBestOrderTrades(token0, token1, buyAmount, sellAmount, dex, tokens, maxHops, weth)
	return PriceResult{
		buyTrade:  buy,
		sellTrade: sell,
		err:       err,
	}
}

// GetActualPriceImpact returns the price impact of the given trade.
// The dex fee gets subtracted from the impact for every trade in the route.
func GetActualPriceImpact(trade *uniswap.Trade, dexFee *big.Int) float64 {
	feePercent := decimal.NewFromInt(10000).Sub(decimal.NewFromBigInt(dexFee, 0)).Div(decimal.NewFromInt(100))
	feeAmount := decimal.NewFromInt(int64(len(trade.Route.Path) - 1))
	fees := feePercent.Mul(feeAmount)
	impact, _ := trade.PriceImpact.Decimal().Sub(fees).Float64()
	return impact
}
