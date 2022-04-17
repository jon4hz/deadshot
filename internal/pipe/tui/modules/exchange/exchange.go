package exchange

import (
	ctx "context"
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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const minDexListHeight = 10

type dexListItem struct {
	dex *database.Dex
}

func (i dexListItem) Title() string       { return i.dex.GetName() }
func (i dexListItem) Description() string { return i.dex.GetRouter() }
func (i dexListItem) FilterValue() string { return i.dex.GetName() }

type state int

const (
	stateUnknown state = iota
	stateReady
)

var (
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

type Module struct {
	ctx    ctx.Context
	cancel ctx.CancelFunc
	D      modules.Default
	state  state
	err    error
	help   help.Model
	kv     keyvalue.Model

	dexList list.Model
}

func NewModule(module *modules.Default) *Module {
	del := list.NewDefaultDelegate()
	del.Styles.SelectedDesc.Foreground(style.GetMainColor()).BorderForeground(style.GetSecondColor())
	del.Styles.SelectedTitle.Foreground(style.GetMainColor()).BorderForeground(style.GetSecondColor())
	return &Module{
		D: modules.Default{
			PrePipe:  module.PrePipe,
			Pipe:     module.Pipe,
			PostPipe: module.PostPipe,
		},
		cancel:  func() {},
		help:    help.New(),
		kv:      keyvalue.New(),
		dexList: list.New(nil, del, 0, 0),
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "exchange module" }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil

	items := make([]list.Item, len(c.Network.GetDexes()))
	for i, dex := range c.Network.GetDexes() {
		items[i] = dexListItem{dex: dex}
	}
	m.dexList.SetItems(items)
	m.dexList.SetShowHelp(false)
	m.dexList.Title = "Please select the dex you want to use"
	m.dexList.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF")) // TODO replace color with adaptive color
	m.dexList.Styles.FilterCursor.Foreground(style.GetMainColor())

	return modules.Resize
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Back) && !m.dexList.SettingFilter() && m.dexList.FilterState() != list.FilterApplied:
			m.D.Ctx.Dex = nil
			return modules.Back

		case key.Matches(msg, defaultKeyMap.Quit) && !m.dexList.SettingFilter():
			return tea.Quit

		case key.Matches(msg, defaultKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return modules.Resize

		case key.Matches(msg, defaultKeyMap.Enter):
			dex := m.dexList.SelectedItem().(dexListItem)
			m.D.Ctx.Dex = dex.dex
			return modules.Next
		}
	}
	switch m.state {
	case stateReady:
		var cmd tea.Cmd
		m.dexList, cmd = m.dexList.Update(msg)
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
	))
	return s.String()
}

func (m *Module) SetContentSize(width, height int) {
	m.dexList.SetSize(width, height)
}

func (m *Module) MinContentHeight() int {
	return minDexListHeight
}

func (m *Module) Content() string {
	var s strings.Builder
	switch m.state {
	case stateReady:
		s.WriteString(m.dexList.View())
	}
	return s.String()
}
func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	switch m.state {
	case stateReady:
		if m.dexList.FilterState() == list.Filtering {
			return m.help.View(filteringKeys)
		}
		if m.dexList.FilterState() == list.FilterApplied {
			return m.help.View(filteredKeys())
		}
		return m.help.View(defaultKeyMap)
	}
	return ""
}

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
