package tradetype

import (
	ctx "context"
	"fmt"
	"io"
	"strings"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/keyvalue"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var itemStyle = lipgloss.NewStyle().PaddingLeft(2)

type state int

const (
	stateUnknown state = iota
	stateCheckListing
	stateSelectType
)

type item struct {
	tradeType *database.TradeType
	text      string
	forkMsg   modules.ForkMsg
}

func genListItems() []item {
	return []item{
		{
			text:      "Order Trade",
			tradeType: database.DefaultTradeTypes.GetOrder(),
			forkMsg:   modules.ForkMsgOrder,
		},
		{
			text:      "Swap",
			tradeType: database.DefaultTradeTypes.GetMarket(),
			forkMsg:   modules.ForkMsgSwap,
		},
	}
}

var (
	_ modules.Forker          = (*Module)(nil)
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

func (i item) FilterValue() string { return i.text }
func (i item) String() string      { return i.text }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	fn := itemStyle.Width(m.Width()).Render
	if index == m.Index() {
		fn = func(s string) string {
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				style.MainStyle.Copy().
					Render("> "),
				style.MainStyle.Copy().
					Width(m.Width()-itemStyle.GetPaddingLeft()).
					Render(s),
			)
		}
	}

	fmt.Fprintf(w, fn(i.String()))
}

type Module struct {
	ctx        ctx.Context
	cancel     ctx.CancelFunc
	D          modules.Default
	state      state
	err        error
	help       help.Model
	kv         keyvalue.Model
	spinner    spinner.Model
	pipeMsg    string
	pipeCancel ctx.CancelFunc
	list       list.Model
	items      []item
}

func NewModule(module *modules.Default) *Module {
	return &Module{
		D: modules.Default{
			PrePipe:     module.PrePipe,
			Pipe:        module.Pipe,
			PostPipe:    module.PostPipe,
			ForkBackMsg: module.ForkBackMsg,
		},
		cancel:  func() {},
		help:    help.New(),
		kv:      keyvalue.New(),
		list:    list.New(nil, itemDelegate{}, 0, 0),
		spinner: style.GetSpinnerPoints(),
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "tradetype module" }

func (m *Module) Skip() bool {
	return false
}

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil

	listItems := genListItems()

	items := make([]list.Item, len(listItems))
	for i := range listItems {
		items[i] = listItems[i]
	}
	m.list.SetItems(items)
	m.list.SetShowHelp(false)
	m.list.SetFilteringEnabled(false)
	m.list.Title = "Please select the trade type"
	m.list.Styles.Title = lipgloss.NewStyle()
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(true)
	m.list.SetHeight(10)
	return tea.Batch(
		modules.Resize,
		spinner.Tick,
	)
}

func (m *Module) Fork() modules.ForkMsg {
	i, ok := m.list.SelectedItem().(item)
	if !ok {
		return modules.ForkMsgNone
	}
	m.D.Ctx.TradeType = i.tradeType
	return i.forkMsg
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateCheckListing:
			switch {
			case key.Matches(msg, defaultKeys.Back):
				m.pipeCancel()
				return modules.Back
			case key.Matches(msg, defaultKeys.Quit):
				m.pipeCancel()
				return tea.Quit
			case key.Matches(msg, defaultKeys.Help):
				m.help.ShowAll = !m.help.ShowAll
			}

		case stateSelectType:
			switch {
			case key.Matches(msg, defaultKeys.Enter):
				return func() tea.Msg { return m.Fork() }
			case key.Matches(msg, defaultKeys.Back):
				return modules.Back
			case key.Matches(msg, defaultKeys.Quit):
				return tea.Quit
			case key.Matches(msg, defaultKeys.Help):
				m.help.ShowAll = !m.help.ShowAll
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return cmd
		}

	case modules.PipeCancelFuncMsg:
		m.pipeCancel = ctx.CancelFunc(msg)

	case modules.PipeMsg:
		m.pipeMsg = string(msg)

	case modules.PipeDoneMsg:
		m.state = stateSelectType
		return nil

	case modules.ErrMsg:
		m.err = msg

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}
	return nil
}

func (m *Module) SetHeaderWidth(width int) { m.kv.SetWidth(width) }
func (m *Module) Header() string {
	var s strings.Builder
	s.WriteString(style.GenLogo())
	s.WriteString("\n\n")
	s.WriteString(m.kv.View(
		keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet()),
		keyvalue.NewKV("Endpoint", m.D.Ctx.Endpoint.GetURL()),
		keyvalue.NewKV("Tokens", m.D.Ctx.Token0.GetSymbol()+" / "+m.D.Ctx.Token1.GetSymbol()),
		keyvalue.NewKV("Balance",
			m.D.Ctx.Token0.GetBalanceDecimal(m.D.Ctx.Token0.GetDecimals()).String()+
				" / "+m.D.Ctx.Token1.GetBalanceDecimal(m.D.Ctx.Token1.GetDecimals()).String(),
		),
		keyvalue.NewKV("Exchange", m.D.Ctx.Dex.GetName()),
	))
	return s.String()
}

func (m *Module) SetContentSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m *Module) MinContentHeight() int {
	return 5 // TODO: don't hardcode that value
}

func (m *Module) Content() string {
	var s strings.Builder
	switch m.state {
	case stateCheckListing:
		s.WriteString("\n")
		s.WriteString(m.spinner.View())
		s.WriteString(" " + m.pipeMsg)
	case stateSelectType:
		s.WriteString(m.list.View())
	}
	return s.String()
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	switch m.state {
	case stateCheckListing:
		return ""
	case stateSelectType:
		return m.help.View(defaultKeys)
	}
	return ""
}

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
