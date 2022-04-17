package addtarget

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/button"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jon4hz/ethconvert/pkg/ethconvert"
	"github.com/shopspring/decimal"
)

type (
	DoneMsg   struct{}
	TargetMsg struct{ Target *database.RawTarget }
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
	menuIndex   int
	menuChoices []menuChoice
	err         modules.Error
	tokenSymbol string
	width       int

	allowPercentagePrice bool

	priceInput    textinput.Model
	amountInput   textinput.Model
	slippageInput textinput.Model
	gasPriceInput textinput.Model
	stopLoss      button.Model
	inclStopLoss  bool

	Help help.Model
}

func New(tokenSymbol string, allowPercentagePrice, inclStopLoss bool) *Model {
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

	slTrigger := []string{" ", "x"}
	slButton := button.New(slTrigger, false)

	return &Model{
		menuChoices: menuChoices,

		priceInput:    pi,
		amountInput:   ai,
		slippageInput: si,
		gasPriceInput: gp,
		stopLoss:      slButton,
		inclStopLoss:  inclStopLoss,

		allowPercentagePrice: allowPercentagePrice,

		tokenSymbol: tokenSymbol,
		Help:        help.New(),
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeys.Back):
			return func() tea.Msg { return DoneMsg{} }

		case key.Matches(msg, defaultKeys.Up):
			return m.menuChoiceBackward()

		case key.Matches(msg, defaultKeys.Down):
			return m.menuChoiceForward()

		case key.Matches(msg, defaultKeys.Enter):
			return m.handleMenuChoice()

		case key.Matches(msg, defaultKeys.Quit):
			return tea.Quit

		case key.Matches(msg, defaultKeys.Help):
			m.Help.ShowAll = !m.Help.ShowAll
			return nil
		}
	}

	var cmd tea.Cmd
	switch m.menuIndex {
	case int(choicePrice):
		m.priceInput, cmd = m.priceInput.Update(msg)
	case int(choiceAmount):
		m.amountInput, cmd = m.amountInput.Update(msg)
	case int(choiceStopLoss):
		m.stopLoss, cmd = m.stopLoss.Update(msg)
	case int(choiceSlippage):
		m.slippageInput, cmd = m.slippageInput.Update(msg)
	case int(choiceGasPrice):
		m.gasPriceInput, cmd = m.gasPriceInput.Update(msg)
	}
	return cmd
}

func (m *Model) handleFocus() tea.Cmd {
	var cmd tea.Cmd
	switch m.menuIndex {
	case int(choiceAmount):
		if !m.amountInput.Focused() {
			cmd = m.amountInput.Focus()
		}
		m.priceInput.Blur()
		m.slippageInput.Blur()
		m.gasPriceInput.Blur()
	case int(choicePrice):
		m.amountInput.Blur()
		if !m.priceInput.Focused() {
			cmd = m.priceInput.Focus()
		}
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
		if !m.slippageInput.Focused() {
			cmd = m.slippageInput.Focus()
		}
		m.gasPriceInput.Blur()
	case int(choiceGasPrice):
		m.amountInput.Blur()
		m.priceInput.Blur()
		m.slippageInput.Blur()
		if !m.gasPriceInput.Focused() {
			cmd = m.gasPriceInput.Focus()
		}
	}
	return cmd
}

func (m *Model) menuChoiceForward() tea.Cmd {
	m.menuIndex++
	if !m.inclStopLoss && m.menuIndex == int(choiceStopLoss) {
		m.menuIndex++
	}
	if m.menuIndex >= len(m.menuChoices) {
		m.menuIndex = 0
	}
	return m.handleFocus()
}

func (m *Model) menuChoiceBackward() tea.Cmd {
	m.menuIndex--
	if !m.inclStopLoss && m.menuIndex == int(choiceStopLoss) {
		m.menuIndex--
	}
	if m.menuIndex < 0 {
		m.menuIndex = len(m.menuChoices) - 1
	}
	return m.handleFocus()
}

var (
	errInvalidPrice = modules.Error{
		Message: "Invalid price",
		Help:    "Please try another price",
	}
	errInvalidAmount = modules.Error{
		Message: "Invalid amount",
		Help:    "Please try another amount",
	}
	errInvalidSlippage = modules.Error{
		Message: "Invalid slippage",
		Help:    "Please try another slippage",
	}
	errInvalidGasPrice = modules.Error{
		Message: "Invalid gas price",
		Help:    "Please try another gas price",
	}
)

func (m *Model) handleMenuChoice() tea.Cmd {
	return func() tea.Msg {
		price := strings.TrimSpace(m.priceInput.Value())
		if price == "" {
			if !m.allowPercentagePrice {
				m.err = errInvalidPrice
				return nil
			}
		}
		amount := strings.TrimSpace(m.amountInput.Value())
		if amount == "" {
			m.err = errInvalidAmount
			return nil
		}

		exactPrice := true
		if strings.HasSuffix(price, "%") {
			if !m.allowPercentagePrice {
				m.err = errInvalidPrice
				return nil
			}

			exactPrice = false
			price = strings.TrimSuffix(price, "%")
			p, err := strconv.ParseFloat(price, 64)
			if err != nil {
				m.err = errInvalidPrice
				return nil
			}
			if p < 0 {
				m.err = errInvalidPrice
				return nil
			}
		}

		exactAmount := true
		if strings.HasSuffix(amount, "%") {
			amount = strings.TrimSuffix(amount, "%")
			a, err := strconv.ParseFloat(amount, 64)
			if err != nil {
				m.err = errInvalidAmount
				return nil
			}
			if a <= 0 || a > 100 {
				m.err = errInvalidAmount
				return nil
			}
			exactAmount = false
		}

		// if the price and amount are not percentages, we need to check if it is a valid number
		if exactPrice {
			_, err := decimal.NewFromString(price)
			if err != nil {
				m.err = errInvalidPrice
				return nil
			}
		}
		if exactAmount {
			a, err := decimal.NewFromString(amount)
			if err != nil {
				m.err = errInvalidAmount
				return nil
			}
			if a.Equal(decimal.Zero) {
				m.err = errInvalidAmount
				return nil
			}
		}

		slippage := strings.TrimSpace(m.slippageInput.Value())
		if slippage != "" {
			slippage = strings.TrimSuffix(slippage, "%")
		}
		var slip float64
		if slippage != "" {
			s, err := strconv.ParseFloat(slippage, 64)
			if err != nil {
				m.err = errInvalidSlippage
				return nil
			}
			if s < 0 || s > 100 {
				m.err = errInvalidSlippage
				return nil
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
				m.err = errInvalidGasPrice
				return nil
			}
			if gas.LessThan(decimal.Zero) {
				m.err = errInvalidGasPrice
				return nil
			}
			gwei, err := ethconvert.ToWei(gas, ethconvert.Gwei)
			if err != nil {
				m.err = errInvalidGasPrice
				return nil
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

func (m Model) View() string {
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

	if !reflect.DeepEqual(m.err, modules.Error{}) {
		s.WriteString("\n" + m.err.Render(m.width) + "\n")
	}
	return s.String()
}

func (m Model) ShowHelp() string {
	return m.Help.View(defaultKeys)
}
