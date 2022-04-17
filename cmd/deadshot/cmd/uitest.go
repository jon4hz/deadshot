package cmd

import (
	"strconv"
	"strings"

	"github.com/jon4hz/deadshot/internal/config"
	"github.com/jon4hz/deadshot/internal/database"

	uiPassword "github.com/jon4hz/deadshot/internal/ui/bubbles/password"
	uiTarget "github.com/jon4hz/deadshot/internal/ui/bubbles/target"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var uitestFlags struct {
	module    string
	altscreen bool
}

var uitestCmd = &cobra.Command{
	Use:    "uitest",
	Short:  "test ui modules",
	Long:   `test ui modules`,
	Hidden: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := database.InitDB(); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return uiTest(uitestFlags.module, uitestFlags.altscreen)
	},
}

func init() {
	uitestCmd.Flags().BoolVar(&uitestFlags.altscreen, "altscreen", true, "use altscreen")
	uitestCmd.Flags().StringVarP(&uitestFlags.module, "module", "m", "", "module to test")
	err := uitestCmd.MarkFlagRequired("module")
	if err != nil {
		panic(err)
	}
}

var quitTextStyle = lipgloss.NewStyle().Margin(0, 0, 0, 0)

const (
	target   = "target"
	password = "password"
	input    = "input"
)

type uiTestModel struct {
	target     *uiTarget.Model
	password   *uiPassword.Model
	req        string
	inputModel textinput.Model
	quitting   string
}

func uiTest(module string, altscreen bool) error {
	m, err := newUITestModel(strings.ToLower(module))
	if err != nil {
		return err
	}
	var opts []tea.ProgramOption
	if altscreen {
		opts = append(opts, tea.WithAltScreen())
	}
	return tea.NewProgram(m, opts...).Start()
}

func newUITestModel(req string) (tea.Model, error) {
	cfg, err := config.Get(true)
	if err != nil {
		return nil, err
	}

	m := uiTestModel{
		req: req,
	}

	switch req {
	case target:
		t0 := database.NewToken("0x7ceb23fd6bc0add59e62ac25578270cff1b9f619", "WETH", 18, false, nil)
		t1 := database.NewToken("0x2791bca1f2de4661ed88a30c99a7a9449aa84174", "USDC", 6, false, nil)
		m.target = uiTarget.NewModel(cfg.Runtime, uiTarget.TargetTypeBuy, t0, t1, nil, 0, false, true, true)
	case password:
		m.password = uiPassword.NewModel(true)
	case input:
		m.inputModel = textinput.NewModel()
		m.inputModel.Focus()
	}
	return m, nil
}

func (m uiTestModel) Init() tea.Cmd {
	switch m.req {
	case target:
		return m.target.Init()
	case password:
		return m.password.Init()
	case input:
		return nil
	}
	return nil
}

func (m uiTestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	switch m.req {
	case target:
		return m, m.target.Update(msg)
	case password:
		return m, m.password.Update(msg)
	case input:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				m.quitting = strings.TrimSpace(m.inputModel.Value())
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.inputModel, cmd = m.inputModel.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m uiTestModel) View() string {
	switch m.req {
	case target:
		return m.target.View("")
	case password:
		return m.password.View()
	case input:
		if m.quitting != "" {
			var s strings.Builder
			for _, r := range m.quitting {
				s.WriteString(strconv.QuoteRuneToASCII(r))
				s.WriteByte('\n')
			}
			return quitTextStyle.Render(s.String())
		}
		return m.inputModel.View()
	}
	return "invalid module"
}
