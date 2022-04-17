package price

import (
	ctx "context"

	chain "github.com/jon4hz/deadshot/internal/blockchain"
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ratelimit"
)

const priceFeedMaxHops = 3

var (
	_ modules.Piper    = &Pipe{}
	_ modules.Canceler = &Pipe{}
)

type Pipe struct {
	c      ctx.Context
	cancel ctx.CancelFunc
}

func (p Pipe) String() string               { return "price" }
func (Pipe) Skip(ctx *context.Context) bool { return false }

func (p *Pipe) CancelFunc() ctx.CancelFunc {
	p.c, p.cancel = ctx.WithCancel(ctx.Background())
	return p.cancel
}

func (p Pipe) Run(ctx *context.Context) error {
	ctx.Price = chain.NewPriceWithContext(p.c, p.cancel)
	ctx.Price.StartFeed(
		ctx.Client,
		ctx.Token0, ctx.Token1,
		ctx.Dex, ctx.Network.GetTokens(),
		ratelimit.GetPriceFeedInterval(ctx.Endpoint.GetURL()), priceFeedMaxHops, ctx.Network.GetWETH())
	return nil
}
