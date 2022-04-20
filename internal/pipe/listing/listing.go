package listing

import (
	ctx "context"
	"errors"

	chain "github.com/jon4hz/deadshot/internal/blockchain"
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
)

var (
	_ modules.Piper         = &Pipe{}
	_ modules.Canceler      = &Pipe{}
	_ modules.PipeMessenger = &Pipe{}
)

type Pipe struct {
	c      ctx.Context
	cancel ctx.CancelFunc
}

func (p Pipe) String() string               { return "listing" }
func (Pipe) Message() string                { return "Checking if all tokens are listed" }
func (Pipe) Skip(ctx *context.Context) bool { return false }

func (p *Pipe) CancelFunc() ctx.CancelFunc {
	p.c, p.cancel = ctx.WithCancel(ctx.Background())
	return p.cancel
}

func (p Pipe) Run(ctx *context.Context) error {
	errChan := make(chan error)
	go func() {
		_, err := ctx.Client.CheckListed(ctx.Token0, ctx.Token1, ctx.Dex, ctx.Network.GetWETH(), ctx.Network.Connectors())
		errChan <- err
	}()
	select {
	case <-p.c.Done():
		return nil
	case err := <-errChan:
		if err != nil {
			if errors.Is(err, chain.ErrNoTradeFound) || errors.Is(err, chain.ErrNoPairsFound) {
				ctx.TradeType = database.DefaultTradeTypes.GetSnipe()
				return nil
			}
			return err
		}
	}
	return nil
}
