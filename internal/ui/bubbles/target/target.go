package target

import (
	"fmt"
	"strings"
	"time"

	"github.com/jon4hz/deadshot/internal/config"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/panel"
	"github.com/jon4hz/deadshot/internal/ui/style"
	"github.com/jon4hz/deadshot/pkg/ethutils"

	uiAddTarget "github.com/jon4hz/deadshot/internal/ui/bubbles/addtarget"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	SkipTargetMsg      struct{}
	TargetsSelectedMsg struct{ Targets []*database.Target }
	tickMsg            struct{}
)

type TargetType int

const (
	TargetTypeBuy TargetType = iota
	TargetTypeSell
)

func (t TargetType) String() string {
	switch t {
	case TargetTypeBuy:
		return "buy"
	case TargetTypeSell:
		return "sell"
	default:
		return "unknown"
	}
}

func (t TargetType) GetType() *database.TargetType {
	switch t {
	case TargetTypeBuy:
		return database.DefaultTargetTypes.GetBuy()
	case TargetTypeSell:
		return database.DefaultTargetTypes.GetSell()
	default:
		return nil
	}
}

const priceUpdateInterval = 200

type state int

const (
	stateReady state = iota
	stateAdd
	stateView
)

type menuChoice int

const (
	choiceAdd menuChoice = iota
	choiceView
	choiceSkip
)

var menuChoices map[menuChoice]string

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

type Model struct {
	Done      bool
	allowSkip bool

	cfg *config.RuntimeConfig

	targetType TargetType
	targets    []*database.Target
	tradeID    uint

	token0, token1 *database.Token

	addTarget  *uiAddTarget.Model
	targetList list.Model
	state      state
	menuIndex  int

	allowPercentagePrice bool
	inclStopLoss         bool

	infoPanel *panel.Model
	mainPanel *panel.Model
}

func NewModel(
	cfg *config.RuntimeConfig,
	targetType TargetType,
	token0, token1 *database.Token,
	targets []*database.Target,
	tradeID uint,
	allowSkip, allowPercentagePrice, inclStopLoss bool,
) *Model {
	infoPanel := panel.NewModel(
		false,
		false,
		lipgloss.Border{}, lipgloss.NoColor{}, lipgloss.NoColor{},
	)

	mainPanel := panel.NewModel(
		false,
		false,
		lipgloss.Border{}, lipgloss.NoColor{}, lipgloss.NoColor{},
	)

	return &Model{
		state:      stateReady,
		targetType: targetType,
		targets:    targets,
		tradeID:    tradeID,

		cfg: cfg,

		token0: token0,
		token1: token1,

		allowSkip:            allowSkip,
		allowPercentagePrice: allowPercentagePrice,
		inclStopLoss:         inclStopLoss,

		infoPanel: infoPanel,
		mainPanel: mainPanel,
	}
}

func (m *Model) Init() tea.Cmd {
	// ensure targets is not nil
	if m.targets == nil {
		m.targets = make([]*database.Target, 0)
	}

	m.genMenuChoices()

	m.state = stateReady
	return tea.Batch(
		textinput.Blink,
		tickCmd(),
	)
}

func (m *Model) genMenuChoices() {
	if m.allowSkip {
		menuChoices = map[menuChoice]string{
			choiceAdd:  fmt.Sprintf("Add %s target", m.targetType.String()),
			choiceView: "View targets (coming soon)",
			choiceSkip: "Skip",
		}
	} else {
		menuChoices = map[menuChoice]string{
			choiceAdd:  fmt.Sprintf("Add %s target", m.targetType.String()),
			choiceView: "View targets (coming soon)",
		}
	}
}

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Done  key.Binding
	Enter key.Binding
}

var keyMapDefault = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "move down"),
	),
	Done: key.NewBinding(
		key.WithKeys("esc", "return"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
}

var keyMapSelect = keyMap{
	Up:    keyMapDefault.Up,
	Down:  keyMapDefault.Down,
	Done:  keyMapDefault.Done,
	Enter: keyMapDefault.Enter,
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateReady:
			switch {
			case key.Matches(msg, keyMapSelect.Done):
				m.Done = true
				return nil

			case key.Matches(msg, keyMapSelect.Up):
				m.menuChoiceBackward()

			case key.Matches(msg, keyMapSelect.Down):
				m.menuChoiceForward()

			case key.Matches(msg, keyMapSelect.Enter):
				return m.handleMenuChoice()
			}
		case stateView:
		}

	case uiAddTarget.TargetMsg:
		m.state = stateReady
		rawTarget := msg.Target
		if m.targetType.GetType().GetType() == database.DefaultTargetTypes.GetBuy().GetType() {
			target := rawTarget.Target(m.token0, m.token0, database.DefaultAmountModes.GetAmountIn(), database.DefaultTargetTypes.GetBuy(), m.tradeID)
			m.targets = append(m.targets, target)
		} else {
			target := rawTarget.Target(m.token0, m.token1, database.DefaultAmountModes.GetAmountIn(), database.DefaultTargetTypes.GetSell(), m.tradeID)
			m.targets = append(m.targets, target)
		}
		if !m.inclStopLoss {
			return func() tea.Msg { return TargetsSelectedMsg{Targets: m.targets} } // TODO: remove to support multiple targets on all occasions
		}
	case tickMsg:
		if m.state == stateReady || m.state == stateAdd {
			return tickCmd()
		}
	}

	return m.updateChildren(msg)
}

func (m *Model) updateChildren(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.state {
	case stateAdd:
		cmd = m.addTarget.Update(msg)
		if m.addTarget.Done {
			m.state = stateReady
			m.addTarget = nil
			return nil
		}
	case stateView:
		//		cmd = m.v
	}
	return cmd
}

func (m *Model) menuChoiceForward() {
	m.menuIndex++
	if m.menuIndex >= len(menuChoices) {
		m.menuIndex = 0
	}
}

func (m *Model) menuChoiceBackward() {
	m.menuIndex--
	if m.menuIndex < 0 {
		m.menuIndex = len(menuChoices) - 1
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*priceUpdateInterval, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m *Model) View(infoView string) string {
	m.infoPanel.SetContent(m.infoView(infoView))
	m.mainPanel.SetContent(m.mainView())

	if infoView != "" && m.cfg.WindowHeight >= m.mainPanel.GetHeight() {
		return lipgloss.JoinVertical(
			lipgloss.Top,
			m.infoPanel.View(),
			m.mainPanel.View(),
		)
	}
	return m.mainPanel.View()
}

func (m *Model) infoView(content string) string {
	c := strings.Count(content, "\n")
	m.infoPanel.SetHeight(c + 1)
	return content
}

func (m *Model) mainView() string {
	switch m.state {
	case stateReady:
		return m.readyView()
	case stateAdd:
		return m.addView()
	case stateView:
		return "\nWIP"
	}
	return "\nWIP 2.0"
}

func (m *Model) readyView() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("\nManage your %s targets:\n\n", m.targetType.String()))
	for i := 0; i < len(menuChoices); i++ {
		e := "  "
		if i == m.menuIndex {
			e = style.GetFocusedPrompt()
			e += style.GetFocusedText(menuChoices[menuChoice(i)])
		} else {
			e += menuChoices[menuChoice(i)]
		}
		if i < len(menuChoices)-1 {
			e += "\n"
		}
		s.WriteString(e)
	}

	m.mainPanel.SetHeight(strings.Count(s.String(), "\n") + 1 + m.infoPanel.GetHeight())
	return s.String()
}

func (m *Model) addView() string {
	v := m.addTarget.View()
	m.mainPanel.SetHeight(m.addTarget.GetHeight() + m.infoPanel.GetHeight())
	return v
}

func (m *Model) handleMenuChoice() tea.Cmd {
	switch m.menuIndex {
	case int(choiceAdd):
		m.state = stateAdd
		m.addTarget = uiAddTarget.NewModel(m.token0.GetSymbol(), m.allowPercentagePrice, m.inclStopLoss)
		return m.addTarget.Init()

	case int(choiceView):
		items := make([]list.Item, len(m.targets))
		for i, target := range m.targets {
			items[i] = targetListItem{Target: target}
		}
		del := list.NewDefaultDelegate()
		del.Styles.SelectedDesc.Foreground(style.GetMainColor()).BorderForeground(style.GetSecondColor())
		del.Styles.SelectedTitle.Foreground(style.GetMainColor()).BorderForeground(style.GetSecondColor())
		l := list.NewModel(items, del, 0, len(m.targets))
		l.Title = "Select a target and press d to delete it"
		l.Styles.Title = style.GetListTitleStyle()
		l.Styles.FilterCursor.Foreground(style.GetMainColor())
		m.targetList = l
		m.state = stateView
		return nil

	case int(choiceSkip):
		if len(m.targets) == 0 {
			return func() tea.Msg { return SkipTargetMsg{} }
		}
		return func() tea.Msg { return TargetsSelectedMsg{Targets: m.targets} }
	}
	return nil
}
