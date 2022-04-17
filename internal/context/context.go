package context

import (
	ctx "context"
	"time"

	"github.com/jon4hz/deadshot/internal/config"
	"github.com/jon4hz/deadshot/internal/database"

	chain "github.com/jon4hz/deadshot/internal/blockchain"
)

type Context struct {
	ctx.Context
	Cfg        *config.Cfg
	Config     *config.Config
	DisableTUI bool

	KeystoreExists bool
	SetNewSecret   bool
	EditSecret     bool
	NewSecret      *Secret

	Unlocked         bool
	LastUnlockedTime time.Time
	UnlockerMsg      string

	Network *database.Network

	LatencyResultChan chan LatencyResult
	LatencyResultDone chan struct{}
	BestLatency       int
	Endpoint          *database.Endpoint
	Client            *chain.Client

	TokenContract string
	Token0        *database.Token
	Token1        *database.Token

	Dex       *database.Dex
	TradeType *database.TradeType

	Price *chain.Price

	RawBuyTargets  []*database.RawTarget
	BuyTargets     database.Targets
	RawSellTargets []*database.RawTarget
	SellTargets    database.Targets

	Trade *database.Trade
}

type Secret struct {
	Secret     string
	Index      int
	IsMnmeonic bool
}

type LatencyResult struct {
	URL     string
	Latency int
}

// New context.
func New(config *config.Config, cfg *config.Cfg) *Context {
	return Wrap(ctx.Background(), config, cfg)
}

// Wrap wraps an existing context.
func Wrap(ctx ctx.Context, config *config.Config, cfg *config.Cfg) *Context {
	return &Context{
		Context:           ctx,
		Config:            config,
		Cfg:               cfg,
		LatencyResultChan: make(chan LatencyResult),
		LatencyResultDone: make(chan struct{}),
	}
}
