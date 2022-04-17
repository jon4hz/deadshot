package balance

import (
	"errors"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/pipe/token"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
)

var _ modules.PipeMessenger = Pipe{}

type Pipe struct{}

func (Pipe) String() string  { return "token balance" }
func (Pipe) Message() string { return "Getting Balance..." }

func (Pipe) Run(ctx *context.Context) error {
	if token.IsForbidden(ctx) {
		return modules.Error{
			Message: "Forbidden",
			Help:    "You can't use the same token twice.",
		}
	}

	var token *database.Token
	if ctx.Token0 != nil && ctx.Token1 == nil {
		token = ctx.Token0
	} else if ctx.Token0 != nil && ctx.Token1 != nil {
		token = ctx.Token1
	} else {
		return errors.New("No token found")
	}

	balance, err := ctx.Client.GetBalanceOf(ctx.Config.Wallet.GetWallet(), token.GetContract())
	if err != nil {
		return modules.Error{
			Message: "Error getting balance",
			Help:    err.Error(),
		}
	}
	token.SetBalance(balance)
	err = database.UpdateBalanceByContractAndNetworkID(token.GetContract(), ctx.Network.GetID(), balance)
	if err != nil {
		logging.Log.WithField("err", err).Error("Error saving token to database")
	}
	return nil
}
