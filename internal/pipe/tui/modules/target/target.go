package target

import (
	ctx "context"
	"errors"
	"fmt"
	"strings"
	"time"

	chain "github.com/jon4hz/deadshot/internal/blockchain"
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	addtarget "github.com/jon4hz/deadshot/internal/pipe/tui/modules/target/addTarget"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/keyvalue"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"
	"github.com/jon4hz/deadshot/pkg/ethutils"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg struct{}

const significantDecimals = 6

type Type int

const (
	TypeBuy Type = iota
	TypeSell
)

func (t Type) String() string {
	switch t {
	case TypeBuy:
		return "buy"
	case TypeSell:
		return "sell"
	default:
		return "unknown"
	}
}

func (t Type) GetType() *database.TargetType {
	switch t {
	case TypeBuy:
		return database.DefaultTargetTypes.GetBuy()
	case TypeSell:
		return database.DefaultTargetTypes.GetSell()
	default:
		return nil
	}
}

const priceUpdateInterval = 200

type state int

const (
	stateUnknown state = iota
	stateReady
	stateAdd
	stateView
)

type menuChoice int

const (
	choiceAdd menuChoice = iota
	choiceView
	choiceSkip
)

type targetListItem struct {
	Target *database.Target
}

func (i targetListItem) Title() string {
	return ethutils.ShowSignificant(i.Target.GetPrice(), i.Target.GetPriceDecimals(), 5)
}

func (i targetListItem) Description() string {
	return ethutils.ShowSignificant(i.Target.GetAmount(), i.Target.GetAmountDecimals(), 5)
}

func (i targetListItem) FilterValue() string {
	return i.Title()
}

var (
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

type Module struct {
	ctx        ctx.Context
	cancel     ctx.CancelFunc
	D          modules.Default
	state      state
	err        error
	help       help.Model
	kv         keyvalue.Model
	pipeMsg    string
	pipeCancel ctx.CancelFunc

	menuChoices          map[menuChoice]string
	targetType           Type
	targetList           list.Model
	menuIndex            int
	allowPercentagePrice bool
	inclStopLoss         bool
	addTarget            *addtarget.Model
}

func NewModule(module *modules.Default, targetType Type) *Module {
	del := list.NewDefaultDelegate()
	del.Styles.SelectedDesc.Foreground(style.GetMainColor()).BorderForeground(style.GetSecondColor())
	del.Styles.SelectedTitle.Foreground(style.GetMainColor()).BorderForeground(style.GetSecondColor())

	m := &Module{
		D: modules.Default{
			PrePipe:     module.PrePipe,
			Pipe:        module.Pipe,
			PostPipe:    module.PostPipe,
			ForkBackMsg: module.ForkBackMsg,
		},
		cancel:     func() {},
		help:       help.New(),
		kv:         keyvalue.New(),
		targetList: list.New(nil, del, 0, 0),
		targetType: targetType,
		pipeCancel: func() {},
	}
	switch targetType {
	case TypeBuy:
		m.allowPercentagePrice = false
		m.inclStopLoss = false
	case TypeSell:
		m.allowPercentagePrice = true
		m.inclStopLoss = true
	}
	return m
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "target module" }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil
	var items []list.Item

	if m.targetType == TypeBuy {
		items = make([]list.Item, len(c.BuyTargets))
		for i, t := range c.BuyTargets {
			items[i] = targetListItem{t}
		}
		m.targetList.Title = "Buy Targets"
	} else if m.targetType == TypeSell {
		items = make([]list.Item, len(c.SellTargets))
		for i, t := range c.SellTargets {
			items[i] = targetListItem{t}
		}
		m.targetList.Title = "Sell Targets"
	}
	m.targetList.SetItems(items)
	m.targetList.SetShowHelp(false)
	m.targetList.Styles.Title = style.GetListTitleStyle()
	m.targetList.Styles.FilterCursor.Foreground(style.GetMainColor())

	m.genMenuChoices()

	return tickCmd()
}

func (m *Module) genMenuChoices() {
	if m.targetType == TypeBuy {
		m.menuChoices = map[menuChoice]string{
			choiceAdd:  fmt.Sprintf("Add %s target", m.targetType.String()),
			choiceView: "View targets (coming soon)",
		}
	} else {
		m.menuChoices = map[menuChoice]string{
			choiceAdd:  fmt.Sprintf("Add %s target", m.targetType.String()),
			choiceView: "View targets (coming soon)",
			choiceSkip: "Skip",
		}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*priceUpdateInterval, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateReady:
			switch {
			case key.Matches(msg, defaultKeys.Enter):
				return m.handleMenuChoice()

			case key.Matches(msg, defaultKeys.Down):
				m.menuForward()

			case key.Matches(msg, defaultKeys.Up):
				m.menuBackward()

			case key.Matches(msg, defaultKeys.Back):
				m.pipeCancel()

				m.D.Ctx.BuyTargets = make(database.Targets, 0) // maybe remove this if multiple buy targets should be allowed

				if m.D.ForkBackMsg != 0 {
					return func() tea.Msg { return m.D.ForkBackMsg }
				}
				return modules.Back

			case key.Matches(msg, defaultKeys.Help):
				m.help.ShowAll = !m.help.ShowAll
				return modules.Resize

			case key.Matches(msg, defaultKeys.Quit):
				m.pipeCancel()
				return tea.Quit
			}

		case stateAdd:
			return m.addTarget.Update(msg)
		}

	case modules.PipeCancelFuncMsg:
		m.pipeCancel = ctx.CancelFunc(msg)

	case modules.ErrMsg:
		m.err = msg

	case tickMsg:
		return tickCmd()

	case addtarget.DoneMsg:
		m.state = stateReady
		return nil

	case addtarget.TargetMsg:
		m.state = stateReady
		rawTarget := msg.Target
		if m.targetType == TypeBuy {
			target := rawTarget.Target(m.D.Ctx.Token0, m.D.Ctx.Token0, database.DefaultAmountModes.GetAmountIn(), database.DefaultTargetTypes.GetBuy(), 0) // TODO: properly implement tradeID
			m.D.Ctx.BuyTargets = append(m.D.Ctx.BuyTargets, target)
			return modules.Next // TODO: remove to support multiple targets on all occasions
		} else {
			target := rawTarget.Target(m.D.Ctx.Token0, m.D.Ctx.Token1, database.DefaultAmountModes.GetAmountIn(), database.DefaultTargetTypes.GetSell(), 0) // TODO: properly implement tradeID
			m.D.Ctx.SellTargets = append(m.D.Ctx.SellTargets, target)
		}
	}
	return nil
}

func (m *Module) menuForward() {
	m.menuIndex++
	if m.menuIndex >= len(m.menuChoices) {
		m.menuIndex = 0
	}
}

func (m *Module) menuBackward() {
	m.menuIndex--
	if m.menuIndex < 0 {
		m.menuIndex = len(m.menuChoices) - 1
	}
}

func (m *Module) handleMenuChoice() tea.Cmd {
	switch menuChoice(m.menuIndex) {
	case choiceSkip:
		return modules.Next

	case choiceAdd:
		m.state = stateAdd
		var token string
		if m.D.Ctx.Token0 != nil {
			token = m.D.Ctx.Token0.GetSymbol()
		}

		add := addtarget.New(token, m.allowPercentagePrice, m.inclStopLoss)
		m.addTarget = add
		return m.addTarget.Init()
	}
	return nil
}

func (m *Module) SetHeaderWidth(width int) { m.kv.SetWidth(width) }
func (m *Module) Header() string {
	var s strings.Builder
	s.WriteString(style.GenLogo())
	s.WriteString("\n\n")
	kvs := []keyvalue.KeyValue{
		keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet()),
		keyvalue.NewKV("Endpoint", m.D.Ctx.Endpoint.GetURL()),
		keyvalue.NewKV("Tokens", m.D.Ctx.Token0.GetSymbol()+" / "+m.D.Ctx.Token1.GetSymbol()),
		keyvalue.NewKV("Balance",
			m.D.Ctx.Token0.GetBalanceDecimal(m.D.Ctx.Token0.GetDecimals()).String()+
				" / "+m.D.Ctx.Token1.GetBalanceDecimal(m.D.Ctx.Token1.GetDecimals()).String(),
		),
		keyvalue.NewKV("Exchange", m.D.Ctx.Dex.GetName()),
	}
	var price string
	p, err := m.D.Ctx.Price.BuyTrade()
	if err != nil {
		if p != nil {
			price = p.ExecutionPrice.Invert().ToSignificant(significantDecimals) + " (failed to reload price)"
		} else {
			if m.D.Ctx.TradeType.Is(database.DefaultTradeTypes.GetSnipe()) && (errors.Is(err, chain.ErrNoTradeFound) || errors.Is(err, chain.ErrNoPairsFound)) {
				price = "Waiting for price"
			} else {
				price = "Failed to fetch price"
			}
		}
	} else if p == nil {
		price = "None"
	} else {
		price = p.ExecutionPrice.Invert().ToSignificant(significantDecimals)
	}
	kvs = append(kvs, keyvalue.NewKV("Price", price))
	s.WriteString(m.kv.View(kvs...))
	return s.String()
}

func (m *Module) SetContentSize(width, height int) {
	m.targetList.SetSize(width, height)
	if m.addTarget != nil {
		m.addTarget.SetWidth(width)
	}
}

func (m *Module) MinContentHeight() int {
	return lipgloss.Height(m.Content())
	// return 5 // TODO: don't hardcode that value
}

func (m *Module) Content() string {
	var s strings.Builder
	switch m.state {
	case stateReady:
		s.WriteString(m.readyView())
	case stateAdd:
		s.WriteString(m.addTarget.View())
	}
	return s.String()
}

func (m *Module) readyView() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("\nManage your %s targets:\n\n", m.targetType.String()))
	for i := 0; i < len(m.menuChoices); i++ {
		e := "  "
		if i == m.menuIndex {
			e = style.GetFocusedPrompt()
			e += style.GetFocusedText(m.menuChoices[menuChoice(i)])
		} else {
			e += m.menuChoices[menuChoice(i)]
		}
		if i < len(m.menuChoices)-1 {
			e += "\n"
		}
		s.WriteString(e)
	}
	return s.String()
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) {
	m.help.Width = width
	if m.addTarget != nil {
		m.addTarget.Help.Width = width
	}
}

func (m *Module) Footer() string {
	switch m.state {
	case stateAdd:
		return m.addTarget.ShowHelp()
	}
	return m.help.View(defaultKeys)
}

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
