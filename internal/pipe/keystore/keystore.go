package keystore

import (
	"errors"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/wallet"
)

type PreCheck struct{}

func (PreCheck) String() string  { return "check keystore" }
func (PreCheck) Message() string { return "" }

func (PreCheck) Skip(ctx *context.Context) bool {
	return wallet.IsInitialized()
}

func (PreCheck) Run(ctx *context.Context) error {
	if err := wallet.InitKeystore("", ctx.Cfg.Keystore); err != nil {
		return err
	}
	exists, err := wallet.CheckKeystore()
	if err != nil {
		return err
	}
	ctx.KeystoreExists = exists
	return nil
}

type Pipe struct{}

func (Pipe) String() string                 { return "keystore" }
func (Pipe) Skip(ctx *context.Context) bool { return false }

func (Pipe) Run(ctx *context.Context) error {
	if err := wallet.InitKeystore(ctx.Cfg.Password, ctx.Cfg.Keystore); err != nil {
		return err
	}
	if ctx.KeystoreExists {
		if err := wallet.Load(ctx.Config.Wallet); err != nil {
			if errors.Is(err, wallet.ErrEmptyTradeWallet) {
				secret, isMnemonic, err := wallet.SearchMnemonic() //nolint: govet
				if err != nil {
					return err
				}
				newSecret := &context.Secret{
					Secret: secret,
				}
				if isMnemonic {
					newSecret.Index = -1
					newSecret.IsMnmeonic = true
				}
				ctx.NewSecret = newSecret
				return nil
			}
			return err
		}
	}
	return nil
}
