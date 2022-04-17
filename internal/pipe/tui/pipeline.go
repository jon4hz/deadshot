package tui

import (
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/balance"
	endpointPipe "github.com/jon4hz/deadshot/internal/pipe/endpoint"
	keystorePipe "github.com/jon4hz/deadshot/internal/pipe/keystore"
	"github.com/jon4hz/deadshot/internal/pipe/listing"
	"github.com/jon4hz/deadshot/internal/pipe/price"
	secretPipe "github.com/jon4hz/deadshot/internal/pipe/secret"
	tokenPipe "github.com/jon4hz/deadshot/internal/pipe/token"
	"github.com/jon4hz/deadshot/internal/pipe/trade"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/endpoint"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/exchange"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/keyderivation"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/keystore"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/market"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/menu"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/network"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/order"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/quit"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/secret"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/settings"
	settingsEndpoint "github.com/jon4hz/deadshot/internal/pipe/tui/modules/settings/endpoint"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/settings/walletsettings"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/target"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/token"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules/tradetype"

	tea "github.com/charmbracelet/bubbletea"
)

type Pipe struct{}

func (Pipe) String() string { return "starting tui" }

func (Pipe) Run(ctx *context.Context) error {
	ui := tea.NewProgram(newTui(ctx), tea.WithAltScreen())
	if err := ui.Start(); err != nil {
		return err
	}
	return nil
}

type pipeline struct {
	modules []modules.Module
}

var newTuiPipeline = func() []modules.Module {
	return []modules.Module{
		newQuit(),
		keystore.NewModule(
			&modules.Default{
				PrePipe: []modules.Piper{
					keystorePipe.PreCheck{},
				},
				Pipe: nil,
				PostPipe: []modules.Piper{
					keystorePipe.Pipe{},
				},
			},
		),
		secret.NewModule(&modules.Default{}),
		keyderivation.NewModule(
			&modules.Default{
				PrePipe: nil,
				Pipe:    nil,
				PostPipe: []modules.Piper{
					secretPipe.Pipe{},
				},
			},
		),
		menu.NewModule(&modules.Default{}),
	}
}

func newQuit() modules.Module {
	return quit.NewModule(&modules.Default{})
}

func newTradePipeline() []modules.Module {
	ms := []modules.Module{
		network.NewModule(&modules.Default{}),
		endpoint.NewModule(&modules.Default{
			Pipe: []modules.Piper{
				&endpointPipe.Pipe{},
			},
		}),
		token.NewModule(&modules.Default{
			PostPipe: []modules.Piper{
				tokenPipe.Pipe{},
				balance.Pipe{},
			},
		}, true),
		token.NewModule(&modules.Default{
			PostPipe: []modules.Piper{
				tokenPipe.Pipe{},
				balance.Pipe{},
			},
		}, false),
		exchange.NewModule(&modules.Default{}),
		tradetype.NewModule(&modules.Default{
			Pipe: []modules.Piper{
				&listing.Pipe{},
			},
		}),
	}

	m := ms[0].(*network.Module) // Make sure we have the right type
	m.D.ForkBackMsg = modules.ForkBackMsg(len(ms))
	ms[0] = m

	return ms
}

var newSwapPipeline = func() []modules.Module {
	ms := []modules.Module{
		market.New(&modules.Default{
			PrePipe: []modules.Piper{
				trade.Spawn{},
			},
		}),
		newQuit(),
	}

	m := ms[0].(*market.Module) // Make sure we have the right type
	m.D.ForkBackMsg = modules.ForkBackMsg(len(ms))
	ms[0] = m

	return ms
}

var newOrderPipeline = func() []modules.Module {
	ms := []modules.Module{
		target.NewModule(&modules.Default{
			PrePipe: []modules.Piper{
				&price.Pipe{},
			},
		}, target.TypeBuy),
		target.NewModule(&modules.Default{}, target.TypeSell),
		order.New(&modules.Default{
			PrePipe: []modules.Piper{
				trade.Spawn{},
			},
		}),
		newQuit(),
	}

	m := ms[0].(*target.Module) // Make sure we have the right type
	m.D.ForkBackMsg = modules.ForkBackMsg(len(ms))
	ms[0] = m

	return ms
}

var newSettingsPipeline = func() []modules.Module {
	ms := []modules.Module{
		settings.New(&modules.Default{}),
	}

	m := ms[0].(*settings.Module) // Make sure we have the right type
	m.D.ForkBackMsg = modules.ForkBackMsg(len(ms))
	ms[0] = m

	return ms
}

var newSettingsEndpointPipeline = func() []modules.Module {
	ms := []modules.Module{
		network.NewModule(&modules.Default{}),
		settingsEndpoint.New(&modules.Default{}),
	}

	mn := ms[0].(*network.Module) // Make sure we have the right type
	mn.D.ForkBackMsg = modules.ForkBackMsg(len(ms))
	ms[0] = mn

	return ms
}

var newSettingsWalletPipeline = func() []modules.Module {
	ms := []modules.Module{
		walletsettings.New(&modules.Default{}),
	}
	m := ms[0].(*walletsettings.Module) // Make sure we have the right type
	m.D.ForkBackMsg = modules.ForkBackMsg(len(ms))
	ms[0] = m

	return ms
}

var newSettingsWalletDerivationPipeline = func() []modules.Module {
	ms := []modules.Module{
		keyderivation.NewModule(
			&modules.Default{
				PrePipe: nil,
				Pipe:    nil,
				PostPipe: []modules.Piper{
					secretPipe.Pipe{},
				},
			},
		),
		newQuit(), // this will make the bot quit when setting a new wallet. It's a bit a hack but feel free to improve it
	}

	m := ms[0].(*keyderivation.Module) // Make sure we have the right type
	m.D.ForkBackMsg = modules.ForkBackMsg(len(ms))
	ms[0] = m

	return ms
}

var newSettingsWalletNewPipeline = func() []modules.Module {
	ms := []modules.Module{
		secret.NewModule(&modules.Default{}),
		keyderivation.NewModule(
			&modules.Default{
				PrePipe: nil,
				Pipe:    nil,
				PostPipe: []modules.Piper{
					secretPipe.Pipe{},
				},
			},
		),
		newQuit(), // this will make the bot quit when setting a new wallet. It's a bit a hack but feel free to improve it
	}

	m := ms[0].(*secret.Module) // Make sure we have the right type
	m.D.ForkBackMsg = modules.ForkBackMsg(len(ms))
	ms[0] = m

	return ms
}
