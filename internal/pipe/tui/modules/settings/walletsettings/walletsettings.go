package walletsettings

import (
	ctx "context"
	"fmt"
	"io"
	"strings"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/keyvalue"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"
	"github.com/jon4hz/deadshot/internal/wallet"

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
	stateReady
)

type item struct {
	text     string
	forkMsg  modules.ForkMsg
	callback func()
}

func (m *Module) genListItems(isMenmonic bool) []item {
	items := make([]item, 0)
	if isMenmonic {
		items = append(items, item{
			text:    "Select a different address",
			forkMsg: modules.ForkMsgWalletSettingsDerivation,
			callback: func() {
				secret, isMnemonic, err := wallet.SearchMnemonic() //nolint: govet
				if err != nil {
					return
				}
				newSecret := &context.Secret{
					Secret: secret,
				}
				if isMnemonic {
					newSecret.Index = -1
					newSecret.IsMnmeonic = true
				}
				m.D.Ctx.NewSecret = newSecret
			},
		})
	}
	items = append(items, []item{
		{
			text:    "Set a new mnemonic phrase",
			forkMsg: modules.ForkMsgWalletSettingsNew,
			callback: func() {
				m.D.Ctx.EditSecret = true
			},
		},
		{
			text:    "Back",
			forkMsg: modules.ForkMsgNone,
		},
	}...)
	return items
}

func (i item) FilterValue() string { return "" }
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

var (
	_ modules.Forker          = (*Module)(nil)
	_ modules.Module          = (*Module)(nil)
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

	list  list.Model
	items []item
}

func New(module *modules.Default) *Module {
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

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "settings module" }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil
	_, isMenmonic, _ := wallet.SearchMnemonic()
	listItems := m.genListItems(isMenmonic)

	items := make([]list.Item, len(listItems))
	for i := range listItems {
		items[i] = listItems[i]
	}
	m.list.SetItems(items)
	m.list.SetShowHelp(false)
	m.list.SetFilteringEnabled(false)
	m.list.Title = "Please select an option"
	m.list.Styles.Title = lipgloss.NewStyle()
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(true)
	m.list.SetHeight(10)
	return modules.Resize
}

func (m *Module) Fork() modules.ForkMsg {
	i, ok := m.list.SelectedItem().(item)
	if !ok {
		return modules.ForkMsgNone
	}
	return i.forkMsg
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateReady:
			switch {
			case key.Matches(msg, defaultKeys.Enter):
				// I'm assuming the back buttom is always the last list item
				if m.list.Index() == len(m.list.Items())-1 {
					if m.D.ForkBackMsg != 0 {
						return func() tea.Msg { return m.D.ForkBackMsg }
					}
					return modules.Back
				}
				if m.list.SelectedItem().(item).callback != nil {
					m.list.SelectedItem().(item).callback()
				}
				return func() tea.Msg { return m.Fork() }

			case key.Matches(msg, defaultKeys.Back):
				if m.D.ForkBackMsg != 0 {
					return func() tea.Msg { return m.D.ForkBackMsg }
				}
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

	case modules.ErrMsg:
		m.err = msg
	}
	return nil
}

func (m *Module) SetHeaderWidth(width int) { m.kv.SetWidth(width) }
func (m *Module) Header() string {
	var s strings.Builder
	s.WriteString(style.GenLogo())
	s.WriteString("\n\n")
	s.WriteString(m.kv.View(
		keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet())),
	)
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
	case stateReady:
		s.WriteString(m.list.View())
	}
	return s.String()
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	switch m.state {
	case stateReady:
		return m.help.View(defaultKeys)
	}
	return ""
}
