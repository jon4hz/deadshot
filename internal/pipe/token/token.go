package token

import (
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/pkg/ethutils"
)

var _ modules.PipeMessenger = Pipe{}

type Pipe struct{}

func (Pipe) String() string                 { return "token info" }
func (Pipe) Message() string                { return "Gathering Infos..." }
func (Pipe) Skip(ctx *context.Context) bool { return ctx.TokenContract == "" }

func (Pipe) Run(ctx *context.Context) error {
	if !ethutils.IsValidAddress(ctx.TokenContract) || ethutils.IsZeroAddress(ctx.TokenContract) {
		return modules.Error{
			Message: "Invalid Address",
			Help:    "Please check if the contract is valid and try again.",
		}
	}
	if IsForbidden(ctx) {
		return modules.Error{
			Message: "Forbidden",
			Help:    "You can't use the same token twice.",
		}
	}

	// TODO: check if token is already in the database

	token := database.NewToken(ctx.TokenContract, "", 0, false, nil)

	info, err := ctx.Client.GetTokenInfo(ctx.TokenContract)
	if err != nil {
		return err
	}
	t, ok := info[ctx.TokenContract]
	if !ok {
		return modules.Error{
			Message: "Token not found",
			Help:    "Please check if the contract is valid and try again.",
		}
	}
	token.SetSymbol(t.GetSymbol())
	token.SetDecimals(t.GetDecimals())
	token.SetNetworkID(ctx.Network.GetID())

	err = database.SaveTokenUniqueByContractAndNetworkID(token)
	if err != nil {
		logging.Log.WithField("error", err).Error("Failed to save token")
	}
	ctx.TokenContract = "" // reset the contract address

	if ctx.Token0 == nil {
		ctx.Token0 = token
	} else {
		ctx.Token1 = token
	}
	return nil
}

func IsForbidden(ctx *context.Context) bool {
	// skip if the first token isn't set yet
	if ctx.Token0 == nil {
		return false
	}

	ctr := ctx.TokenContract

	if ctx.Token1 != nil {
		if ctr == "" {
			ctr = ctx.Token1.GetContract()
		}
	}

	// if the first token is the same as the second one
	if ctr == ctx.Token0.GetContract() {
		return true
	}
	// if token1 is weth and token0 native
	if ctr == ctx.Network.GetWETH() && ethutils.IsZeroAddress(ctx.Token0.GetContract()) {
		return true
	}
	// if token0 is weth and token1 native
	if ethutils.IsZeroAddress(ctr) && ctx.Token0.GetContract() == ctx.Network.GetWETH() && ctr != "" {
		return true
	}
	return false
}
