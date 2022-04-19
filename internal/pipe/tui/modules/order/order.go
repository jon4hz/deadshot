package order

import (
	ctx "context"
	"fmt"
	"math/big"
	"strings"
	"time"

	chain "github.com/jon4hz/deadshot/internal/blockchain"
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/panel"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/common"
	"github.com/jon4hz/deadshot/internal/ui/style"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const priceUpdateInterval = 300

type (
	tickMsg struct{}
	logMsg  string
)

type state int

const (
	stateUnknown state = iota
	stateReady
)

var (
	_ modules.Module           = (*Module)(nil)
	_ modules.ModulePiper      = (*Module)(nil)
	_ simpleview.ContentViewer = (*Module)(nil)
)

type Module struct {
	ctx    ctx.Context
	cancel ctx.CancelFunc
	D      modules.Default
	state  state

	width int

	logs              string
	logChan           chan string
	tradeDispatchCtx  ctx.Context
	tradeDispatchDone ctx.CancelFunc

	infoPanel       *panel.Model
	helpPanel       *panel.Model
	buyTargetPanel  *panel.Model
	sellTargetPanel *panel.Model
	buyTradePanel   *panel.Model
	sellTradePanel  *panel.Model
	logPanel        *panel.Model
}

func New(module *modules.Default) *Module {
	infoPanel := panel.NewModel(
		false, true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(), style.GetInactiveColor(),
	)
	helpPanel := panel.NewModel(
		false, true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(), style.GetInactiveColor(),
	)
	buyTargetPanel := panel.NewModel(
		false, true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(), style.GetInactiveColor(),
	)
	sellTargetPanel := panel.NewModel(
		false, true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(), style.GetInactiveColor(),
	)
	buyTradePanel := panel.NewModel(
		false, true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(), style.GetInactiveColor(),
	)
	sellTradePanel := panel.NewModel(
		false, true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(), style.GetInactiveColor(),
	)
	logPanel := panel.NewModel(
		false, true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(), style.GetInactiveColor(),
	)

	tdctx, cancel := ctx.WithCancel(ctx.Background())
	return &Module{
		D: modules.Default{
			PrePipe:  module.PrePipe,
			Pipe:     module.Pipe,
			PostPipe: module.PostPipe,
		},
		cancel: func() {},

		infoPanel:         infoPanel,
		helpPanel:         helpPanel,
		buyTargetPanel:    buyTargetPanel,
		sellTargetPanel:   sellTargetPanel,
		buyTradePanel:     buyTradePanel,
		sellTradePanel:    sellTradePanel,
		logPanel:          logPanel,
		logChan:           make(chan string),
		tradeDispatchCtx:  tdctx,
		tradeDispatchDone: cancel,
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "order module" }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c

	// TODO: move that to a pipe
	go m.D.Ctx.Client.TradeDispatcher(m.tradeDispatchCtx, m.tradeDispatchDone,
		m.D.Ctx.Config.Wallet, m.D.Ctx.Trade, m.D.Ctx.Price, m.logChan,
	)

	return tea.Batch(
		tickCmd(),
		listenForLogs(m.logChan),
		modules.Resize,
	)
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.D.Ctx.Price.Stop()
			m.tradeDispatchDone()
			return tea.Quit
		}
	case tickMsg:
		return tickCmd()
	case logMsg:
		m.logs += string(msg)
		return listenForLogs(m.logChan)
	}
	return nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*priceUpdateInterval, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func listenForLogs(logs <-chan string) tea.Cmd {
	return func() tea.Msg {
		return logMsg(<-logs)
	}
}

func (m *Module) SetContentSize(width, height int) {
	if width%2 == 1 {
		width--
	}
	m.infoPanel.SetSize(width/2, height/3)
	m.helpPanel.SetSize(width/2, height/3)
	m.buyTradePanel.SetSize(width/4, height/3)
	m.sellTradePanel.SetSize(width/4, height/3)
	m.logPanel.SetSize(width, height/3)
	m.buyTargetPanel.SetSize(width/4, height/3)
	m.sellTargetPanel.SetSize(width/4, height/3)
	m.width = width
}

func (m *Module) Content() string {
	switch m.state {
	case stateReady:
		m.infoPanel.SetContent(m.infoView())
		m.helpPanel.SetContent(m.helpView())
		m.buyTargetPanel.SetContent(m.buyTargetView())
		m.sellTargetPanel.SetContent(m.sellTargetView())
		m.buyTradePanel.SetContent(m.buyTradeView())
		m.sellTradePanel.SetContent(m.sellTradeView())
		m.logPanel.SetContent(m.logView())
		//.logPanel.GotoBottom() // causes a runtime panic (?)

		middleRow := make([]string, 0)
		middleRow = append(middleRow, m.buyTargetPanel.View())
		var padding string
		if m.width%4 != 0 {
			padding = " "
		}
		middleRow = append(middleRow, padding, m.sellTargetPanel.View(), m.buyTradePanel.View(), padding, m.sellTradePanel.View())

		return lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				m.infoPanel.View(),
				m.helpPanel.View(),
			),
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				middleRow...,
			),
			m.logPanel.View(),
		)
	}
	return ""
}

func (m Module) infoView() string {
	s := common.KeyValueViewWithoutVerticalLine(
		"Wallet", m.D.Ctx.Config.Wallet.GetWallet(),
		"Endpoint", m.D.Ctx.Trade.GetEndpoint().GetURL(),
		"Exchange", m.D.Ctx.Trade.GetDex().GetName(),
		"Tokens", m.D.Ctx.Trade.GetToken0().GetSymbol()+" / "+m.D.Ctx.Trade.GetToken1().GetSymbol(),
		"Balance", m.D.Ctx.Trade.GetToken0().GetBalanceDecimal(m.D.Ctx.Trade.GetToken0().GetDecimals()).String()+" / "+m.D.Ctx.Trade.GetToken1().GetBalanceDecimal(m.D.Ctx.Trade.GetToken1().GetDecimals()).String(),
	) + "\n"
	// m.infoPanel.SetHeight(strings.Count(s, "\n") + 2)
	return s
}

func (m Module) buyTradeView() string {
	var s strings.Builder
	buyTrade := m.D.Ctx.Price.GetBuyTrade()
	if buyTrade == nil {
		return ""
	}
	if buyTrade.Route == nil {
		return ""
	}
	s.WriteString("Buy Price: " + buyTrade.ExecutionPrice.Invert().ToSignificant(10) + "\n")
	nb := m.D.Ctx.Trade.GetNextBuyTarget()
	var slippage float64
	if nb != nil {
		slippage = nb.GetSlippage()
	} else {
		slippage = database.DefaultSlippage
	}
	min, _ := buyTrade.MinimumAmountOut(uniswap.NewPercent(big.NewInt(int64(slippage)), big.NewInt(10000)))
	s.WriteString("Minimum received: " + min.ToSignificant(6) + " ")
	s.WriteString(m.D.Ctx.Trade.GetToken1().GetSymbol() + "\n")
	priceImpact := chain.GetActualPriceImpact(buyTrade, m.D.Ctx.Trade.GetDex().GetFeeBigInt())
	if priceImpact < 0.01 {
		s.WriteString("Price Impact: < 0.01%\n")
	} else {
		s.WriteString(fmt.Sprintf("Price Impact: %f", priceImpact) + "%\n")
	}
	s.WriteString("Buy Route:  ")
	for i, v := range buyTrade.Route.Path {
		if i < len(buyTrade.Route.Path)-1 {
			s.WriteString(v.Symbol() + " -> ")
			continue
		}
		s.WriteString(v.Symbol() + "\n")
	}
	if m.D.Ctx.Price.GetError() != nil {
		s.WriteString("Warning: failed to reload the price...")
	}
	return s.String()
}

// TODO: fix duplicated lines (buyTradeView).
func (m Module) sellTradeView() string {
	var s strings.Builder
	sellTrade := m.D.Ctx.Price.GetSellTrade()
	if sellTrade == nil {
		return ""
	}
	if sellTrade.Route == nil {
		return ""
	}
	s.WriteString("Sell Price: " + sellTrade.ExecutionPrice.ToSignificant(10) + "\n")
	nb := m.D.Ctx.Trade.GetNextBuyTarget()
	var slippage float64
	if nb != nil {
		slippage = nb.GetSlippage()
	} else {
		slippage = database.DefaultSlippage
	}
	min, _ := sellTrade.MinimumAmountOut(uniswap.NewPercent(big.NewInt(int64(slippage)), big.NewInt(10000)))
	s.WriteString("Minimum received: " + min.ToSignificant(6) + " ")
	s.WriteString(m.D.Ctx.Trade.GetToken0().GetSymbol() + "\n")
	priceImpact := chain.GetActualPriceImpact(sellTrade, m.D.Ctx.Trade.GetDex().GetFeeBigInt())
	if priceImpact < 0.01 {
		s.WriteString("Price Impact: < 0.01%\n")
	} else {
		s.WriteString(fmt.Sprintf("Price Impact: %f", priceImpact) + "%\n")
	}
	s.WriteString("Sell Route:  ")
	for i, v := range sellTrade.Route.Path {
		if i < len(sellTrade.Route.Path)-1 {
			s.WriteString(v.Symbol() + " -> ")
			continue
		}
		s.WriteString(v.Symbol() + "\n")
	}
	if m.D.Ctx.Price.GetError() != nil {
		s.WriteString("Warning: failed to reload the price...")
	}
	return s.String()
}

func (m Module) helpView() string {
	return ""
}

func (m Module) buyTargetView() string {
	var s strings.Builder
	targets := m.D.Ctx.Trade.GetBuyTargets()
	for _, v := range targets {
		var sign string
		if v.GetFailed() {
			sign = "❌"
		} else if v.GetConfirmed() {
			sign = "✅"
		} else if v.GetHit() {
			sign = "⏳"
		} else {
			sign = "📝"
		}
		s.WriteString(fmt.Sprintf("%s %s - %s %s\n", sign, v.ViewPrice(), v.ViewAmount(), m.D.Ctx.Trade.GetToken0().GetSymbol()))
	}
	if s.String() == "" {
		s.WriteString("No buy targets")
	}
	return s.String()
}

func (m Module) sellTargetView() string {
	var s strings.Builder
	targets := m.D.Ctx.Trade.GetSellTargets()
	for _, v := range targets {
		var sign string
		if v.GetFailed() {
			sign = "❌"
		} else if v.GetConfirmed() {
			sign = "✅"
		} else if v.GetHit() {
			sign = "⏳"
		} else {
			sign = "📝"
		}
		s.WriteString(fmt.Sprintf("%s %s - %s %s\n", sign, v.ViewPrice(), v.ViewAmount(), m.D.Ctx.Trade.GetToken0().GetSymbol()))
	}
	if s.String() == "" {
		s.WriteString("No sell targets")
	}
	return s.String()
}

func (m Module) logView() string {
	return m.logs
}

func (m Module) Wrap() bool { return false }

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
