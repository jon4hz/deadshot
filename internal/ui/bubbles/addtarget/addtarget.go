package addtarget

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/button"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/panel"
	"github.com/jon4hz/deadshot/internal/ui/common"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jon4hz/ethconvert/pkg/ethconvert"
	"github.com/shopspring/decimal"
)

const percentageBitSize = 64

type (
	errInvalidPrice    struct{}
	errInvalidAmount   struct{}
	errInvalidSlippage struct{}
	errInvalidGasPrice struct{}
	TargetMsg          struct{ Target *database.RawTarget }
)

type state int

const (
	stateReady state = iota
	stateInput
)

type menuChoice int

const (
	choicePrice menuChoice = iota
	choiceAmount
	choiceStopLoss
	choiceSlippage
	choiceGasPrice
)

type Model struct {
	Done bool

	state       state
	menuIndex   int
	menuChoices []menuChoice
	errMsg      string
	tokenSymbol string

	allowPercentagePrice bool

	priceInput    textinput.Model
	amountInput   textinput.Model
	slippageInput textinput.Model
	gasPriceInput textinput.Model
	stopLoss      button.Model
	inclStopLoss  bool

	mainPanel *panel.Model
}

func NewModel(tokenSymbol string, allowPercentagePrice, inclStopLoss bool) *Model {
	var menuChoices []menuChoice
	if inclStopLoss {
		menuChoices = []menuChoice{
			choicePrice,
			choiceAmount,
			choiceStopLoss,
			choiceSlippage,
			choiceGasPrice,
		}
	} else {
		menuChoices = []menuChoice{
			choicePrice,
			choiceAmount,
			choiceStopLoss,
			choiceSlippage,
			choiceGasPrice,
		}
	}

	pi := textinput.NewModel()
	pi.Prompt = ""
	pi.CursorStyle = lipgloss.NewStyle().Foreground(style.GetMainColor())
	pi.Placeholder = "0.00"
	pi.Focus()

	ai := textinput.NewModel()
	ai.Prompt = ""
	ai.CursorStyle = lipgloss.NewStyle().Foreground(style.GetMainColor())
	ai.Placeholder = "100"

	si := textinput.NewModel()
	si.Prompt = ""
	si.CursorStyle = lipgloss.NewStyle().Foreground(style.GetMainColor())
	si.Placeholder = "1%"

	gp := textinput.NewModel()
	gp.Prompt = ""
	gp.CursorStyle = lipgloss.NewStyle().Foreground(style.GetMainColor())
	gp.Placeholder = "1 GWEI"

	mainPanel := panel.NewModel(
		false,
		false,
		lipgloss.Border{}, lipgloss.NoColor{}, lipgloss.NoColor{},
	)

	slTrigger := []string{" ", "x"}
	slButton := button.New(slTrigger, false)

	return &Model{
		state:       stateReady,
		menuIndex:   int(stateReady),
		menuChoices: menuChoices,

		priceInput:    pi,
		amountInput:   ai,
		slippageInput: si,
		gasPriceInput: gp,
		stopLoss:      slButton,
		inclStopLoss:  inclStopLoss,

		allowPercentagePrice: allowPercentagePrice,

		mainPanel:   mainPanel,
		tokenSymbol: tokenSymbol,
	}
}

func (m *Model) Init() tea.Cmd {
	m.state = stateInput
	return textinput.Blink
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
		case stateInput:
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
		}

	case errInvalidPrice:
		m.priceInput.Reset()
		m.handleFocus()
		head := lipgloss.NewStyle().Foreground(style.GetErrColor()).Render("Invalid price. ")
		body := common.Subtle("Please enter another price.")
		m.errMsg = common.Wrap(head + body)
		return nil

	case errInvalidAmount:
		m.amountInput.Reset()
		m.handleFocus()
		head := lipgloss.NewStyle().Foreground(style.GetErrColor()).Render("Invalid amount. ")
		body := common.Subtle("Please enter another amount.")
		m.errMsg = common.Wrap(head + body)
		return nil

	case errInvalidSlippage:
		m.slippageInput.Reset()
		m.handleFocus()
		head := lipgloss.NewStyle().Foreground(style.GetErrColor()).Render("Invalid slippage. ")
		body := common.Subtle("Please enter another slippage.")
		m.errMsg = common.Wrap(head + body)
		return nil

	case errInvalidGasPrice:
		m.gasPriceInput.Reset()
		m.handleFocus()
		head := lipgloss.NewStyle().Foreground(style.GetErrColor()).Render("Invalid gas price. ")
		body := common.Subtle("Please enter another gas price.")
		m.errMsg = common.Wrap(head + body)
		return nil
	}

	var cmd tea.Cmd
	switch m.menuIndex {
	case int(choicePrice):
		m.priceInput, cmd = m.priceInput.Update(msg)
		return cmd
	case int(choiceAmount):
		m.amountInput, cmd = m.amountInput.Update(msg)
		return cmd
	case int(choiceStopLoss):
		m.stopLoss, cmd = m.stopLoss.Update(msg)
		return cmd
	case int(choiceSlippage):
		m.slippageInput, cmd = m.slippageInput.Update(msg)
		return cmd
	case int(choiceGasPrice):
		m.gasPriceInput, cmd = m.gasPriceInput.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) handleFocus() {
	switch m.menuIndex {
	case int(choiceAmount):
		m.amountInput.Focus()
		m.priceInput.Blur()
		m.slippageInput.Blur()
		m.gasPriceInput.Blur()
	case int(choicePrice):
		m.amountInput.Blur()
		m.priceInput.Focus()
		m.slippageInput.Blur()
		m.gasPriceInput.Blur()
	case int(choiceStopLoss):
		m.amountInput.Blur()
		m.priceInput.Blur()
		m.slippageInput.Blur()
		m.gasPriceInput.Blur()
	case int(choiceSlippage):
		m.amountInput.Blur()
		m.priceInput.Blur()
		m.slippageInput.Focus()
		m.gasPriceInput.Blur()
	case int(choiceGasPrice):
		m.amountInput.Blur()
		m.priceInput.Blur()
		m.slippageInput.Blur()
		m.gasPriceInput.Focus()
	}
}

func (m *Model) menuChoiceForward() {
	m.menuIndex++
	if !m.inclStopLoss && m.menuIndex == int(choiceStopLoss) {
		m.menuIndex++
	}
	if m.menuIndex >= len(m.menuChoices) {
		m.menuIndex = 0
	}
	m.updateInputFocus()
}

func (m *Model) menuChoiceBackward() {
	m.menuIndex--
	if !m.inclStopLoss && m.menuIndex == int(choiceStopLoss) {
		m.menuIndex--
	}
	if m.menuIndex < 0 {
		m.menuIndex = len(m.menuChoices) - 1
	}
	m.updateInputFocus()
}

func (m *Model) updateInputFocus() {
	m.handleFocus()
}

func (m *Model) View() string {
	m.mainPanel.SetContent(m.mainView())
	return m.mainPanel.View()
}

func (m *Model) mainView() string {
	switch m.state {
	case stateReady:
		return "Whoops"
	case stateInput:
		return m.inputView()
	}
	return "Whoops 2.0"
}

func (m *Model) inputView() string {
	var s strings.Builder
	s.WriteString("\n")

	p := style.GetCustomPrompt()
	if m.menuChoices[m.menuIndex] == choicePrice {
		p = style.GetFocusedPrompt()
	}
	s.WriteString(p + lipgloss.NewStyle().Render("Price: ") + m.priceInput.View() + "\n")

	p = style.GetCustomPrompt()
	if m.menuChoices[m.menuIndex] == choiceAmount {
		p = style.GetFocusedPrompt()
	}
	var tokenString string
	if m.tokenSymbol != "" {
		tokenString = fmt.Sprintf(" (%s)", m.tokenSymbol)
	}
	s.WriteString(p + lipgloss.NewStyle().Render(fmt.Sprintf("Amount%s: ", tokenString)) + m.amountInput.View() + "\n")

	if m.inclStopLoss {
		p = style.GetCustomPrompt()
		if m.menuChoices[m.menuIndex] == choiceStopLoss {
			p = style.GetFocusedPrompt()
		}
		s.WriteString(p + lipgloss.NewStyle().Render("Stop Loss: ") + m.stopLoss.View() + "\n")
	}

	p = style.GetCustomPrompt()
	if m.menuChoices[m.menuIndex] == choiceSlippage {
		p = style.GetFocusedPrompt()
	}
	s.WriteString(p + lipgloss.NewStyle().Render("Slippage: ") + m.slippageInput.View() + "\n")

	p = style.GetCustomPrompt()
	if m.menuChoices[m.menuIndex] == choiceGasPrice {
		p = style.GetFocusedPrompt()
	}
	s.WriteString(p + lipgloss.NewStyle().Render("Gas Price: ") + m.gasPriceInput.View() + "\n")

	if m.errMsg != "" {
		s.WriteString("\n" + m.errMsg + "\n")
	}

	m.mainPanel.SetHeight(strings.Count(s.String(), "\n") + 1)
	return s.String()
}

func (m *Model) GetHeight() int {
	return m.mainPanel.GetHeight()
}

func (m *Model) handleMenuChoice() tea.Cmd {
	return func() tea.Msg {
		price := strings.TrimSpace(m.priceInput.Value())
		if price == "" {
			return errInvalidPrice{}
		}
		amount := strings.TrimSpace(m.amountInput.Value())
		if amount == "" {
			return errInvalidAmount{}
		}

		exactPrice := true
		if strings.HasSuffix(price, "%") {
			if !m.allowPercentagePrice {
				return errInvalidPrice{}
			}

			exactPrice = false
			price = strings.TrimSuffix(price, "%")
			p, err := strconv.ParseFloat(price, percentageBitSize)
			if err != nil {
				return errInvalidPrice{}
			}
			if p < 0 {
				return errInvalidPrice{}
			}
		}

		exactAmount := true
		if strings.HasSuffix(amount, "%") {
			amount = strings.TrimSuffix(amount, "%")
			a, err := strconv.ParseFloat(amount, percentageBitSize)
			if err != nil {
				return errInvalidAmount{}
			}
			if a <= 0 || a > 100 {
				return errInvalidAmount{}
			}
			exactAmount = false
		}

		// if the price and amount are not percentages, we need to check if it is a valid number
		if exactPrice {
			_, err := decimal.NewFromString(price)
			if err != nil {
				return errInvalidPrice{}
			}
		}
		if exactAmount {
			a, err := decimal.NewFromString(amount)
			if err != nil {
				return errInvalidAmount{}
			}
			if a.Equal(decimal.Zero) {
				return errInvalidAmount{}
			}
		}

		slippage := strings.TrimSpace(m.slippageInput.Value())
		if slippage != "" {
			slippage = strings.TrimSuffix(slippage, "%")
		}
		var slip float64
		if slippage != "" {
			s, err := strconv.ParseFloat(slippage, percentageBitSize)
			if err != nil {
				return errInvalidSlippage{}
			}
			if s < 0 || s > 100 {
				return errInvalidSlippage{}
			}
			slip = s * 100 // support up to 2 decimal places
		} else {
			slip = database.DefaultSlippage
		}

		gasPrice := strings.TrimSpace(m.gasPriceInput.Value())
		if gasPrice != "" {
			gasPrice = strings.TrimSpace(strings.TrimSuffix(gasPrice, "GWEI"))
		}
		var gasP *big.Int
		if gasPrice != "" {
			gas, err := decimal.NewFromString(gasPrice)
			if err != nil {
				return errInvalidGasPrice{}
			}
			if gas.LessThan(decimal.Zero) {
				return errInvalidGasPrice{}
			}
			gwei, err := ethconvert.ToWei(gas, ethconvert.Gwei)
			if err != nil {
				return errInvalidGasPrice{}
			}
			gasP = gwei.BigInt()
		} else {
			gasP = nil
		}

		return TargetMsg{
			Target: database.NewRawTarget(price, amount, exactPrice, exactAmount, slip, gasP, m.stopLoss.Triggered()),
		}
	}
}
