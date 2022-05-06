package market

import (
	ctx "context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	chain "github.com/jon4hz/deadshot/internal/blockchain"
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/logstream"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/panel"
	"github.com/jon4hz/deadshot/internal/ui/common"
	"github.com/jon4hz/deadshot/internal/ui/style"
	"github.com/jon4hz/deadshot/pkg/ethutils"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

type (
	tradeInfoMsg     struct{ t *uniswap.Trade }
	errTradeInfo     struct{ err error }
	errSlippageInput struct{}
	slippageMsg      struct{ slippage float64 }
	swapMsg          struct{ tx *types.Transaction }
	errSwap          struct{ err error }
)

var (
	ErrInvalidToken0  = errors.New("invalid token 0")
	ErrInvalidToken1  = errors.New("invalid token 1")
	ErrGettingBalance = errors.New("getting balance failed")
)

type state int

const (
	stateUnknown state = iota
	stateReady
	stateLoadingData
)

type menuChoice int

const (
	token0Choice menuChoice = iota
	amount0Choice
	token1Choice
	amount1Choice
	swapChoice
	slippageChoice
	invertChoice
)

var menuChoices = []menuChoice{
	token0Choice,
	amount0Choice,
	token1Choice,
	amount1Choice,
	swapChoice,
	invertChoice,
	slippageChoice,
}

var (
	prompt        = style.GetCustomPrompt()
	focusedPrompt = style.GetFocusedCustomPrompt()
)

type tradeInfoResult struct {
	trade *uniswap.Trade
	err   error
}

type Module struct {
	ctx    ctx.Context
	cancel ctx.CancelFunc
	D      modules.Default
	state  state
	width  int

	err  error
	errs string

	infoPanel      *panel.Model
	primaryPanel   *panel.Model
	secondaryPanel *panel.Model
	logPanel       *panel.Model
	spinner        spinner.Model

	token0Input   textinput.Model
	token1Input   textinput.Model
	amount0Input  textinput.Model
	amount1Input  textinput.Model
	slippageInput textinput.Model
	menuIndex     int
	menuChoice    menuChoice
	valChanged    bool

	tradeInfo   *uniswap.Trade
	tradeInfoC  chan tradeInfoResult
	tradeCtx    ctx.Context
	tradeCancel ctx.CancelFunc
}

func New(module *modules.Default) *Module {
	t0I := textinput.NewModel()
	a0I := textinput.NewModel()
	t1I := textinput.NewModel()
	a1I := textinput.NewModel()
	sI := textinput.NewModel()

	primaryPanel := panel.NewModel(
		true,
		true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(),
		style.GetInactiveColor(),
	)

	secondaryPanel := panel.NewModel(
		false,
		true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(),
		style.GetInactiveColor(),
	)

	logPanel := panel.NewModel(
		false,
		true,
		lipgloss.RoundedBorder(),
		style.GetActiveColor(),
		style.GetInactiveColor(),
	)

	infoPanel := panel.NewModel(
		false,
		false,
		lipgloss.Border{}, lipgloss.NoColor{}, lipgloss.NoColor{})

	spr := style.GetSpinnerPoints()

	return &Module{
		D: modules.Default{
			PrePipe:     module.PrePipe,
			Pipe:        module.Pipe,
			PostPipe:    module.PostPipe,
			ForkBackMsg: module.ForkBackMsg,
		},
		cancel: func() {},

		infoPanel:      infoPanel,
		primaryPanel:   primaryPanel,
		secondaryPanel: secondaryPanel,
		logPanel:       logPanel,
		spinner:        spr,

		token0Input:   t0I,
		token1Input:   t1I,
		amount0Input:  a0I,
		amount1Input:  a1I,
		slippageInput: sI,

		menuIndex:  int(amount0Choice),
		menuChoice: amount0Choice,

		tradeInfoC: make(chan tradeInfoResult, 1),
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "market module" }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.tradeCtx, m.tradeCancel = ctx.WithCancel(m.ctx)

	m.token0Input.Placeholder = c.Token0.GetSymbol()
	m.token0Input.Prompt = ""
	m.token0Input.CursorStyle = style.GetActiveCursor()

	m.amount0Input.Placeholder = "0.0"
	m.amount0Input.Prompt = ""
	m.amount0Input.CursorStyle = style.GetActiveCursor()

	m.token1Input.Placeholder = c.Token1.GetSymbol()
	m.token1Input.Prompt = ""
	m.token1Input.CursorStyle = style.GetActiveCursor()

	m.amount1Input.Placeholder = "0.0"
	m.amount1Input.Prompt = ""
	m.amount1Input.CursorStyle = style.GetActiveCursor()

	m.slippageInput.Placeholder = "1%"
	m.slippageInput.Prompt = ""
	m.slippageInput.CursorStyle = style.GetActiveCursor()

	// create a pseudo-target to prevent a nil pointer panic
	m.D.Ctx.Trade.SetBuyTargets(
		[]*database.Target{
			database.NewTargetWithDefaults(),
		},
	)

	return modules.Resize
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return tea.Quit
		}
		switch m.state {
		case stateReady, stateLoadingData:
			switch msg.String() {
			// Prev menu item
			case "esc":
				if m.D.ForkBackMsg != 0 {
					return func() tea.Msg { return m.D.ForkBackMsg }
				}
				return modules.Back

			case "up", "shift+tab":
				m.menuIndex--

				if m.menuIndex < m.primaryPanel.GetYOffset() {
					m.primaryPanel.LineUp(1)
				}
				if m.menuIndex < 0 {
					m.menuIndex = len(menuChoices) - 1
					m.primaryPanel.GotoBottom()
				}
				defer func() { m.menuChoice = menuChoices[m.menuIndex] }()
				if m.menuChoice == slippageChoice {
					m.valChanged = false
					return cmd
				}
				if m.menuChoice != swapChoice && m.valChanged {
					m.getNewTradeInfo()
					cmd = m.listenForTradeInfo(m.tradeInfoC)
					m.valChanged = false
					return tea.Batch(cmd, spinner.Tick)
				}
				m.valChanged = false
				return cmd

			case "down", "tab":
				m.menuIndex++
				if m.menuIndex >= m.primaryPanel.GetHeight() {
					m.primaryPanel.LineDown(1)
				}
				if m.menuIndex >= len(menuChoices) {
					m.menuIndex = 0
					m.primaryPanel.GotoTop()
				}
				defer func() { m.menuChoice = menuChoices[m.menuIndex] }()
				if m.menuChoice == slippageChoice {
					m.valChanged = false
					return cmd
				}
				if m.menuChoice != swapChoice && m.valChanged {
					m.getNewTradeInfo()
					cmd = m.listenForTradeInfo(m.tradeInfoC)
					m.valChanged = false
					return tea.Batch(cmd, spinner.Tick)
				}
				m.valChanged = false
				return cmd

			case "enter":
				switch m.menuChoice {
				case amount0Choice:
					m.D.Ctx.Trade.GetBuyTargets()[0].SetAmountMode(database.DefaultAmountModes.GetAmountIn())
				case amount1Choice:
					m.D.Ctx.Trade.GetBuyTargets()[0].SetAmountMode(database.DefaultAmountModes.GetAmountOut())
				case slippageChoice:
					return setSlippage(m.slippageInput.Value())
				case invertChoice:
					m.D.Ctx.Trade.InvertTokens()
					m.D.Ctx.Trade.GetBuyTargets()[0].InvertAmountMode()
					m.amount0Input, m.amount1Input = m.amount1Input, m.amount0Input
					m.token0Input, m.token1Input = m.token1Input, m.token0Input
					m.tradeInfo = new(uniswap.Trade)
					if m.D.Ctx.Trade.GetBuyTargets()[0].GetActualAmount() == nil && m.D.Ctx.Trade.GetBuyTargets()[0].GetAmountMinMax() == nil {
						return nil
					}
					m.getNewTradeInfo()
					cmd = m.listenForTradeInfo(m.tradeInfoC)
					return tea.Batch(cmd, spinner.Tick)

				case swapChoice:
					if m.D.Ctx.Trade.GetBuyTargets()[0].MarketSwapPossible(m.tradeInfo, m.D.Ctx.Trade.GetToken0()) {
						return m.triggerSwap()
					}
				}

				m.getNewTradeInfo()
				cmd = m.listenForTradeInfo(m.tradeInfoC)
				return tea.Batch(cmd, spinner.Tick)

			default:
				m.valChanged = true
				switch m.menuChoice {
				case token0Choice:
					m.token0Input, cmd = m.token0Input.Update(msg)

				case amount0Choice:
					switch {
					case containsOnly(msg.String(), "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", ".", "backspace", "left", "right", "%"):
						m.D.Ctx.Trade.GetBuyTargets()[0].SetAmountMode(database.DefaultAmountModes.GetAmountIn())
						m.amount0Input, cmd = m.amount0Input.Update(msg)
					}

				case token1Choice:
					m.token1Input, cmd = m.token1Input.Update(msg)

				case amount1Choice:
					switch {
					case containsOnly(msg.String(), "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", ".", "backspace", "left", "right", "%"):
						m.D.Ctx.Trade.GetBuyTargets()[0].SetAmountMode(database.DefaultAmountModes.GetAmountOut())
						m.amount1Input, cmd = m.amount1Input.Update(msg)
					}

				case slippageChoice:
					switch {
					case containsOnly(msg.String(), "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", ".", "backspace", "left", "right", "%"):
						m.slippageInput, cmd = m.slippageInput.Update(msg)
					}
				}
				return cmd
			}
		}

	case tea.WindowSizeMsg:
		m.infoPanel.SetSize(msg.Width, 17)
		m.primaryPanel.SetSize(msg.Width/2, msg.Height/2-18/2)
		m.secondaryPanel.SetSize(msg.Width/2, msg.Height/2-19/2)
		m.logPanel.SetSize(msg.Width/2-3, m.primaryPanel.GetHeight()+m.secondaryPanel.GetHeight()+4)

	// TODO: move functions which modify the model view to the View() method
	case tradeInfoMsg:
		m.state = stateReady
		m.tradeInfo = msg.t
		m.err = nil

		info := m.tradeInfo

		// set primary panel
		m.token0Input.Reset()
		m.token0Input.Placeholder = m.D.Ctx.Trade.GetToken0().GetSymbol()
		m.token0Input.TextStyle = lipgloss.NewStyle()

		m.token1Input.Reset()
		m.token1Input.Placeholder = m.D.Ctx.Trade.GetToken1().GetSymbol()
		m.token1Input.TextStyle = lipgloss.NewStyle()

		// set secondary panel
		switch m.D.Ctx.Trade.GetBuyTargets()[0].GetAmountMode().GetName() {
		case database.DefaultAmountModes.GetAmountIn().GetName():
			m.amount1Input.Reset()
			m.amount1Input.SetValue(info.OutputAmount().ToSignificant(6))

		case database.DefaultAmountModes.GetAmountOut().GetName():
			m.amount0Input.Reset()
			if info != nil {
				m.amount0Input.SetValue(info.InputAmount().ToSignificant(6))
			}
		}

		return tea.Batch(m.listenForTradeInfo(m.tradeInfoC), spinner.Tick)

	case errTradeInfo:
		m.state = stateReady
		m.tradeInfo = nil
		m.err = msg.err
		switch err := m.err.Error(); err {
		case ErrInvalidToken0.Error():
			m.token0Input.TextStyle = lipgloss.NewStyle().Foreground(style.GetErrColor())
		case ErrInvalidToken1.Error():
			m.token1Input.TextStyle = lipgloss.NewStyle().Foreground(style.GetErrColor())
		}

	case slippageMsg:
		m.D.Ctx.Trade.GetBuyTargets()[0].SetSlippage(&msg.slippage)
		m.slippageInput.Reset()
		m.slippageInput.TextStyle = lipgloss.NewStyle()
		m.slippageInput.Placeholder = fmt.Sprint(math.Floor(msg.slippage)/100) + "%" // support up to 2 decimal places
		m.setTradeInfoToWorkflow(m.tradeInfo)                                        // update slippage for min/max amount

	case errSlippageInput:
		m.slippageInput.TextStyle = lipgloss.NewStyle().Foreground(style.GetErrColor())

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd

	case errSwap:
		m.err = msg.err

	case swapMsg:
		if msg.tx != nil {
			m.err = fmt.Errorf("%s", msg.tx.Hash().String())
		}

	default:
		switch m.menuChoice {
		case token0Choice:
			m.token0Input, cmd = m.token0Input.Update(msg)
		case token1Choice:
			m.token1Input, cmd = m.token1Input.Update(msg)
		case amount0Choice:
			m.amount0Input, cmd = m.amount0Input.Update(msg)
		case amount1Choice:
			m.amount1Input, cmd = m.amount1Input.Update(msg)
		case slippageChoice:
			m.slippageInput, cmd = m.slippageInput.Update(msg)
		}
		return cmd
	}
	return nil
}

func (m *Module) getNewTradeInfo() {
	// stop previous trade info gathering
	m.tradeCancel()
	m.tradeCtx, m.tradeCancel = ctx.WithCancel(m.ctx)

	go m.getTradeInfo(
		m.D.Ctx.Trade.GetToken0(), m.D.Ctx.Trade.GetToken1(),
		m.D.Ctx.Trade.GetBuyTargets()[0].GetAmountMode(),
		m.D.Ctx.Trade.GetDex(), m.D.Ctx.Trade.GetNetwork().Connectors(),
		m.D.Ctx.Client, m.tradeInfoC, m.tradeCtx)
	m.state = stateLoadingData
}

func (m *Module) getTradeInfo(token0, token1 *database.Token, amountMode *database.AmountMode, dex *database.Dex, tokens []*database.Token, client *chain.Client, infoC chan<- tradeInfoResult, ctx ctx.Context) {
	type newToken struct {
		token0 string
		token1 string
	}
	newT := newToken{}

	token0Input := strings.TrimSpace(m.token0Input.Value())
	if token0Input != "" {
		if ethutils.IsValidAddress(token0Input) {
			newT.token0 = token0Input
		} else {
			go func() {
				select {
				case infoC <- tradeInfoResult{err: ErrInvalidToken0}:
				case <-ctx.Done():
					return
				}
			}()
			return
		}
	}
	token1Input := strings.TrimSpace(m.token1Input.Value())
	if token1Input != "" {
		if ethutils.IsValidAddress(token1Input) {
			newT.token1 = token1Input
		} else {
			go func() {
				select {
				case infoC <- tradeInfoResult{err: ErrInvalidToken1}:
				case <-ctx.Done():
					return
				}
			}()
			return
		}
	}

	var newTokens map[string]*database.Token
	var err error
	if newT.token0 != "" || newT.token1 != "" {
		newTokens, err = client.GetTokenInfo(newT.token0, newT.token1)
		if err != nil {
			go func() {
				select {
				case infoC <- tradeInfoResult{err: err}:
				case <-ctx.Done():
					return
				}
			}()
			return
		}
		t0, ok := newTokens[newT.token0]
		if ok {
			bal, err := client.GetBalanceOf(m.D.Ctx.Config.Wallet.GetWallet(), t0.GetContract())
			if err != nil {
				go func() {
					select {
					case infoC <- tradeInfoResult{err: ErrGettingBalance}:
					case <-ctx.Done():
						return
					}
				}()
				return
			}
			t0.SetBalance(bal)
			err = database.UpdateBalanceByContractAndNetworkID(t0.GetContract(), m.D.Ctx.Trade.GetNetwork().GetID(), bal)
			if err != nil {
				logging.Log.WithField("err", err).Error("Error saving balance to database")
				go func() {
					select {
					case infoC <- tradeInfoResult{err: err}:
					case <-ctx.Done():
						return
					}
				}()
				return
			}
			m.D.Ctx.Trade.SetToken0(t0)
		}
		t1, ok := newTokens[newT.token1]
		if ok {
			bal, err := client.GetBalanceOf(m.D.Ctx.Config.Wallet.GetWallet(), t1.GetContract())
			if err != nil {
				logging.Log.WithField("err", err).Error("Error saving token to database")
				go func() {
					select {
					case infoC <- tradeInfoResult{err: ErrGettingBalance}:
					case <-ctx.Done():
						return
					}
				}()
				return
			}
			t1.SetBalance(bal)
			err = database.UpdateBalanceByContractAndNetworkID(t1.GetContract(), m.D.Ctx.Trade.GetNetwork().GetID(), bal)
			if err != nil {
				logging.Log.WithField("err", err).Error("Error saving token to database")
				go func() {
					select {
					case infoC <- tradeInfoResult{err: err}:
					case <-ctx.Done():
						return
					}
				}()
				return
			}
			m.D.Ctx.Trade.SetToken1(t1)
		}
	}

	tradeC := make(chan tradeInfoResult, 1)
	go func() {
		select {
		case info := <-tradeC:
			infoC <- info
			return
		case <-ctx.Done():
			return
		}
	}()
	switch amountMode.GetName() {
	case database.DefaultAmountModes.GetAmountIn().GetName():
		amount, err := getAmountXInput(m.amount0Input.Value(), token0)
		if err != nil {
			go func() {
				select {
				case tradeC <- tradeInfoResult{err: err}:
				case <-ctx.Done():
					return
				}
			}()
			return
		}
		if amount == nil {
			// TODO handle empty input properly
			return
		}
		go func() {
			t, err := client.GetBestTradeExactIn(token0, token1, amount, dex, tokens, 5, m.D.Ctx.Trade.GetNetwork().GetWETH())
			select {
			case tradeC <- tradeInfoResult{t, err}:
			case <-ctx.Done():
				return
			}
		}()

	case database.DefaultAmountModes.GetAmountOut().GetName():
		amount, err := getAmountXInput(m.amount1Input.Value(), token1)
		if err != nil {
			go func() {
				select {
				case tradeC <- tradeInfoResult{err: err}:
				case <-ctx.Done():
					return
				}
			}()
			return
		}
		if amount == nil {
			// TODO handle empty input properly
			return
		}
		go func() {
			t, err := client.GetBestTradeExactOut(token0, token1, amount, dex, tokens, 5, m.D.Ctx.Trade.GetNetwork().GetWETH())
			select {
			case tradeC <- tradeInfoResult{t, err}:
			case <-ctx.Done():
				return
			}
		}()
	}
}

// getchain.AmountInput takes an amount as string with the associated token and returns the amount as *big.Int
// the input amount can either be a float or a percentage (e.g. "10.0" or "10%").
func getAmountXInput(amountXInput string, tokenX *database.Token) (*big.Int, error) {
	v := strings.TrimSpace(amountXInput)
	if v == "" {
		// TODO: handle empty input properly
		return nil, nil
	}

	hasSuffix := strings.HasSuffix(v, "%")
	if hasSuffix {
		percentage, err := strconv.ParseFloat(v[:len(v)-1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount")
		}
		if percentage > 100 {
			return nil, fmt.Errorf("invalid amount")
		}
		// support up to 5 decimal places
		normalizedPercentage := new(big.Int).SetInt64(int64(percentage * 1e5))
		return new(big.Int).Quo(new(big.Int).Mul(normalizedPercentage, tokenX.GetBalance()), big.NewInt(1e7)), nil
	}

	amount, err := decimal.NewFromString(v)
	if err != nil {
		return nil, fmt.Errorf("invalid amount")
	}
	return ethutils.ToWei(amount, tokenX.GetDecimals()), nil
}

func (m *Module) listenForTradeInfo(c <-chan tradeInfoResult) tea.Cmd {
	return func() tea.Msg {
		info, ok := <-c
		if !ok {
			return nil
		}
		if info.err != nil {
			return errTradeInfo{info.err}
		}

		m.setTradeInfoToWorkflow(info.trade)

		return tradeInfoMsg{info.trade}
	}
}

func (m *Module) setTradeInfoToWorkflow(info *uniswap.Trade) {
	if info == nil {
		return
	}
	switch m.D.Ctx.Trade.GetBuyTargets()[0].GetAmountMode().GetName() {
	case database.DefaultAmountModes.GetAmountIn().GetName():
		min, _ := info.MinimumAmountOut(uniswap.NewPercent(big.NewInt(int64(m.D.Ctx.Trade.GetBuyTargets()[0].GetSlippage())), big.NewInt(database.MaxSlippage))) // support up to two decimal places for the slippage
		m.D.Ctx.Trade.GetBuyTargets()[0].SetActualAmount(info.InputAmount().Raw())
		m.D.Ctx.Trade.GetBuyTargets()[0].SetAmountMinMax(min.Raw().String())
	case database.DefaultAmountModes.GetAmountOut().GetName():
		max, _ := info.MaximumAmountIn(uniswap.NewPercent(big.NewInt(int64(m.D.Ctx.Trade.GetBuyTargets()[0].GetSlippage())), big.NewInt(database.MaxSlippage))) // support up to two decimal places for the slippage
		m.D.Ctx.Trade.GetBuyTargets()[0].SetActualAmount(info.OutputAmount().Raw())
		m.D.Ctx.Trade.GetBuyTargets()[0].SetAmountMinMax(max.Raw().String())
	}
	m.D.Ctx.Trade.GetBuyTargets()[0].SetPath(info.Route.GetAddresses())
}

// containsOnly returns wether the given string contains only the given character set.
func containsOnly(s string, chars ...string) bool {
	switch s {
	case "backspace", "left", "right":
		return true
	default:
		for _, v := range s {
			if !contains(chars, string(v)) {
				return false
			}
		}
		return true
	}
}

// contains returns whether a given char is contained in the given slice.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func setSlippage(slippageInput string) tea.Cmd {
	return func() tea.Msg {
		slippageInput = strings.TrimSpace(slippageInput)
		slippageInput = strings.TrimSuffix(slippageInput, "%")
		slippage, err := strconv.ParseFloat(slippageInput, 64)
		if err != nil {
			return errSlippageInput{}
		}
		if slippage < 0 || slippage >= 50 {
			return errSlippageInput{}
		}
		slippage *= 100 // support up to 2 decimal places
		return slippageMsg{slippage}
	}
}

func (m *Module) triggerSwap() tea.Cmd {
	return func() tea.Msg {
		tx, err := m.D.Ctx.Client.Swap(m.D.Ctx.Config.Wallet, m.D.Ctx.Trade, m.D.Ctx.Trade.GetBuyTargets()[0])
		if err != nil {
			return errSwap{err}
		}
		return swapMsg{tx}
	}
}

func (m *Module) SetContentSize(width, height int) {
	if width%2 == 1 {
		width--
	}

	switch m.state {
	case stateReady, stateLoadingData:
		m.infoPanel.SetSize(width, lipgloss.Height(m.infoView()))
	}
	m.primaryPanel.SetSize(width/2, height/2-18/2)
	m.secondaryPanel.SetSize(width/2, height/2-19/2)
	m.logPanel.SetSize(width/2, m.primaryPanel.GetHeight()+m.secondaryPanel.GetHeight()+4)

	m.width = width
}

func (m *Module) Content() string {
	switch m.state {
	case stateReady, stateLoadingData:
		m.setInfoView()
		m.setReadyView()
		if m.state == stateLoadingData {
			m.setLoadingDataView()
		} else {
			m.setSecondaryView()
		}
		m.setLogView()

		return lipgloss.JoinVertical(
			lipgloss.Top,
			m.infoPanel.View(),
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				lipgloss.JoinVertical(
					lipgloss.Top,
					m.primaryPanel.View(),
					m.secondaryPanel.View(),
				),
				m.logPanel.View(),
			),
		)
	}
	return ""
}

func (m *Module) setInfoView() {
	m.infoPanel.SetContent(m.infoView())
}

func (m *Module) infoView() string {
	var s strings.Builder
	s.WriteString(style.GenLogo() + "\n\n")
	s.WriteString(common.KeyValueView(
		"Wallet", m.D.Ctx.Config.Wallet.GetWallet(),
		"Web3 Provider", m.D.Ctx.Trade.GetEndpoint().GetURL(),
		"Dex", m.D.Ctx.Trade.GetDex().GetName(),
		"Trading Pair", fmt.Sprintf("%s / %s", m.D.Ctx.Trade.GetToken0().GetSymbol(), m.D.Ctx.Trade.GetToken1().GetSymbol()),
		"Balance", fmt.Sprintf("%s / %s", m.D.Ctx.Trade.GetToken0().GetBalanceDecimal(m.D.Ctx.Trade.GetToken0().GetDecimals()).String(), m.D.Ctx.Trade.GetToken1().GetBalanceDecimal(m.D.Ctx.Trade.GetToken1().GetDecimals()).String()),
	))
	return s.String()
}

func (m *Module) setSecondaryView() {
	m.secondaryPanel.SetContent(m.secondaryView())
}

func (m *Module) secondaryView() string {
	var s strings.Builder
	info := m.tradeInfo

	if info == nil {
		return ""
	}
	if info.Route == nil {
		return ""
	}

	s.WriteString("Price: " + info.ExecutionPrice.Invert().ToSignificant(10) + "\n")

	switch m.D.Ctx.Trade.GetBuyTargets()[0].GetAmountMode().GetName() {
	case database.DefaultAmountModes.GetAmountIn().GetName():
		min, _ := info.MinimumAmountOut(uniswap.NewPercent(big.NewInt(int64(m.D.Ctx.Trade.GetBuyTargets()[0].GetSlippage())), big.NewInt(10000))) // support up to decimal places for the slippage
		s.WriteString("Minimum received: " + min.ToSignificant(6) + " ")
		s.WriteString(m.D.Ctx.Trade.GetToken1().GetSymbol())

	case database.DefaultAmountModes.GetAmountOut().GetName():
		max, _ := info.MaximumAmountIn(uniswap.NewPercent(big.NewInt(int64(m.D.Ctx.Trade.GetBuyTargets()[0].GetSlippage())), big.NewInt(10000))) // support up to decimal places for the slippage
		s.WriteString("Maximum sold: " + max.ToSignificant(6) + " ")
		s.WriteString(m.D.Ctx.Trade.GetToken0().GetSymbol())
	}
	s.WriteString("\n")

	priceImpact := chain.GetActualPriceImpact(info, m.D.Ctx.Trade.GetDex().GetFeeBigInt())
	if priceImpact < 0.01 {
		s.WriteString("Price Impact: < 0.01%\n")
	} else {
		s.WriteString(fmt.Sprintf("Price Impact: %f", priceImpact) + "%\n")
	}

	s.WriteString("Route: ")
	for i, v := range info.Route.Path {
		if i < len(info.Route.Path)-1 {
			s.WriteString(v.Symbol() + " -> ")
			continue
		}
		s.WriteString(v.Symbol() + "\n")
	}

	return s.String()
}

func (m *Module) setLoadingDataView() {
	m.secondaryPanel.SetContent(m.loadingDataView())
}

func (m *Module) loadingDataView() string {
	return m.spinner.View() + " Loading data..."
}

func (m *Module) setLogView() {
	m.logPanel.SetContent(m.logView())
	// m.logPanel.GotoBottom() // TODO: fix
}

func (m *Module) logView() string {
	if m.err != nil {
		m.errs += logstream.Format(m.err.Error(), logstream.INFO) // TODO separate error from info
	}
	m.err = nil
	return m.errs
}

func (m *Module) setReadyView() {
	m.primaryPanel.SetContent(m.readyView())
}

func (m *Module) readyView() string {
	var s strings.Builder

	p := prompt
	m.token0Input.Blur()
	if m.menuChoice == token0Choice {
		p = focusedPrompt
		m.token0Input.Focus()
	}
	s.WriteString(p + lipgloss.NewStyle().Width(10).Render("From: ") + m.token0Input.View())
	s.WriteString("\n")

	p = prompt
	m.amount0Input.Blur()
	if m.menuChoice == amount0Choice {
		p = focusedPrompt
		m.amount0Input.Focus()
	}
	s.WriteString(p + lipgloss.NewStyle().Width(10).Render("Amount: ") + m.amount0Input.View())
	s.WriteString("\n")

	p = prompt
	m.token1Input.Blur()
	if m.menuChoice == token1Choice {
		p = focusedPrompt
		m.token1Input.Focus()
	}
	s.WriteString(p + lipgloss.NewStyle().Width(10).Render("To: ") + m.token1Input.View())
	s.WriteString("\n")

	p = prompt
	m.amount1Input.Blur()
	if m.menuChoice == amount1Choice {
		p = focusedPrompt
		m.amount1Input.Focus()
	}
	s.WriteString(p + lipgloss.NewStyle().Width(10).Render("Amount: ") + m.amount1Input.View())
	s.WriteString("\n")

	p = prompt
	if m.menuChoice == swapChoice {
		p = focusedPrompt
	}
	if m.D.Ctx.Trade.GetBuyTargets()[0].MarketSwapPossible(m.tradeInfo, m.D.Ctx.Trade.GetToken0()) && m.state != stateLoadingData {
		s.WriteString(p + "Swap\n")
	} else {
		s.WriteString(p + common.Subtle("Swap\n"))
	}

	p = prompt
	if m.menuChoice == invertChoice {
		p = focusedPrompt
	}
	s.WriteString(p + "Invert\n")

	p = prompt
	m.slippageInput.Blur()
	if m.menuChoice == slippageChoice {
		p = focusedPrompt
		m.slippageInput.Focus()
	}
	s.WriteString(p + lipgloss.NewStyle().Width(10).Render("Slippage: ") + m.slippageInput.View())
	s.WriteString("\n")

	return s.String()
}

func (m Module) Wrap() bool { return false }

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
