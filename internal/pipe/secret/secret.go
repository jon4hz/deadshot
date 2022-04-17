package secret

import (
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/wallet"
)

var _ modules.PipeMessenger = Pipe{}

type Pipe struct{}

func (Pipe) String() string                 { return "secret" }
func (Pipe) Message() string                { return "" }
func (Pipe) Skip(ctx *context.Context) bool { return false }

func (Pipe) Run(ctx *context.Context) error {
	if ctx.NewSecret != nil {
		if err := wallet.Set(ctx.NewSecret.Secret, uint(ctx.NewSecret.Index), ctx.Config.Wallet); err != nil {
			return err
		}
		ctx.NewSecret = nil
	}
	return wallet.Load(ctx.Config.Wallet)
}
