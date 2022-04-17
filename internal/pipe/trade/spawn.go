package trade

import (
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
)

type Spawn struct{}

func (Spawn) String() string { return "Spawn a new trade" }

func (Spawn) Run(ctx *context.Context) error {
	trade := database.NewTrade(ctx.Token0, ctx.Token1,
		ctx.BuyTargets, ctx.SellTargets, ctx.TradeType,
		ctx.Endpoint, ctx.Network, ctx.Dex)
	ctx.Trade = trade
	return nil
}
