package network

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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var itemStyle = lipgloss.NewStyle().PaddingLeft(2)

type state int

const (
	stateUnknown state = iota
	stateSelect
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
	err    error
	state  state

	help  help.Model
	kv    keyvalue.Model
	list  list.Model
	items []item

	networkIndex int
}

type item struct {
	network *database.Network
}

func (i item) FilterValue() string { return i.network.FullName }
func (i item) String() string      { return i.network.FullName }

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

func NewModule(module *modules.Default) *Module {
	return &Module{
		D: modules.Default{
			PrePipe:     module.PrePipe,
			Pipe:        module.Pipe,
			PostPipe:    module.PostPipe,
			ForkBackMsg: module.ForkBackMsg,
		},
		cancel: func() {},
		help:   help.New(),
		kv:     keyvalue.New(),
		list:   list.New(nil, itemDelegate{}, 0, 0),
	}
}

func (m *Module) Cancel()                        { m.cancel() }
func (m *Module) State() int                     { return int(m.state) }
func (m *Module) String() string                 { return "network module" }
func (m *Module) Skip(ctx *context.Context) bool { return false }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil
	items := make([]list.Item, len(m.D.Ctx.Config.Networks))
	for i, n := range m.D.Ctx.Config.Networks {
		items[i] = item{network: n}
	}
	m.list.Styles.FilterCursor.Foreground(style.GetMainColor())
	m.list.Title = "Please select the network you want to use"
	m.list.Styles.Title = lipgloss.NewStyle()
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(true)
	m.list.SetHeight(10)
	m.list.SetShowHelp(false)
	m.list.SetItems(items)
	return modules.Resize
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Enter) && !m.list.SettingFilter():
			m.err = nil
			m.D.Ctx.Network = m.list.SelectedItem().(item).network
			return modules.Next

		case key.Matches(msg, defaultKeyMap.Back) && !m.list.SettingFilter() && m.list.FilterState() != list.FilterApplied:
			m.D.Ctx.Network = nil
			if m.D.ForkBackMsg != 0 {
				return func() tea.Msg { return m.D.ForkBackMsg }
			}
			return modules.Back

		case key.Matches(msg, defaultKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return modules.Resize

		case key.Matches(msg, defaultKeyMap.Quit) && !m.list.SettingFilter():
			return tea.Quit
		}

	case modules.ErrMsg:
		m.err = msg
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return cmd
}

func (m *Module) SetHeaderWidth(width int) { m.kv.SetWidth(width) }
func (m *Module) Header() string {
	var s strings.Builder
	s.WriteString(style.GenLogo())
	s.WriteString("\n\n")
	s.WriteString(m.kv.View(
		keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet()),
	))
	return s.String()
}

func (m *Module) SetContentSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m *Module) Content() string {
	switch m.state {
	case stateSelect:
		return m.list.View()
	}
	return ""
}

func (m *Module) MinContentHeight() int {
	return 5 // TODO: don't hardcode that value
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	switch m.state {
	case stateSelect:
		if m.list.FilterState() == list.Filtering {
			return m.help.View(filteringKeys)
		}
		if m.list.FilterState() == list.FilterApplied {
			return m.help.View(filteredKeys())
		}
		return m.help.View(defaultKeyMap)
	}
	return ""
}

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
