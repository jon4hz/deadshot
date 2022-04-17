package endpoint

import (
	ctx "context"
	"strconv"
	"strings"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/keyvalue"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/common"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	benchmarkDoneMsg struct{}
	latencyResultMsg struct{ l *context.LatencyResult }
)

type state int

const (
	stateUnknown state = iota
	stateBenchmarking
	stateDone
)

var (
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

type Module struct {
	ctx        ctx.Context
	cancel     ctx.CancelFunc
	D          modules.Default
	err        error
	state      state
	spinner    spinner.Model
	results    []context.LatencyResult
	pipeCancel ctx.CancelFunc

	help help.Model
	kv   keyvalue.Model
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
		help:    help.NewModel(),
		kv:      keyvalue.New(),
		spinner: style.GetSpinnerPoints(),
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "endpoint module" }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil
	m.results = make([]context.LatencyResult, 0)

	return tea.Batch(
		m.listenForResults(),
		m.spinner.Tick,
	)
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Back):
			if m.state == stateBenchmarking {
				m.pipeCancel()
				m.D.Ctx.LatencyResultDone <- struct{}{}
			}
			return modules.Back

		case key.Matches(msg, defaultKeyMap.Quit):
			return tea.Quit
		default:
			if m.state == stateDone {
				return modules.Next
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd

	case modules.ErrMsg:
		m.err = msg

	case latencyResultMsg:
		if msg.l != nil {
			m.results = append(m.results, *msg.l)
			return tea.Batch(
				m.listenForResults(),
			)
		}
		return nil

	case modules.PipeDoneMsg:
		m.state = stateDone

	case modules.PipeCancelFuncMsg:
		m.pipeCancel = ctx.CancelFunc(msg)
	}
	return nil
}

func (m *Module) listenForResults() tea.Cmd {
	return func() tea.Msg {
		select {
		case r, ok := <-m.D.Ctx.LatencyResultChan:
			if !ok {
				return nil
			}
			return latencyResultMsg{&r}

		case <-m.ctx.Done():
			return nil
		}
	}
}

func (m *Module) SetHeaderWidth(width int) { m.kv.SetWidth(width) }
func (m *Module) Header() string {
	var s strings.Builder
	s.WriteString(style.GenLogo())
	s.WriteString("\n\n")
	switch m.state {
	case stateBenchmarking:
		s.WriteString(m.kv.View(
			keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet()),
		))
	case stateDone:
		kvs := []keyvalue.KeyValue{
			keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet()),
		}
		if m.D.Ctx.Endpoint != nil {
			kvs = append(kvs,
				keyvalue.NewKV("Endpoint", m.D.Ctx.Endpoint.GetURL()),
				keyvalue.NewKV("Latency", strconv.Itoa(m.D.Ctx.BestLatency)+"ms"))
		}
		s.WriteString(m.kv.View(kvs...))
	}
	return s.String()
}

func (m *Module) SetContentSize(width, height int) {}
func (m *Module) MinContentHeight() int {
	switch m.state {
	case stateBenchmarking:
		return 5 // TODO: don't hardcode that value
	case stateDone:
		return 0
	}
	return 0
}

func (m *Module) Content() string {
	var s strings.Builder
	switch m.state {
	case stateBenchmarking:
		s.WriteString(m.benchmarkView())
	}
	return s.String()
}

func (m *Module) benchmarkView() string {
	var s strings.Builder
	s.WriteString("\n" + m.spinner.View() + " Benchmarking...\n")
	for _, v := range m.results {
		s.WriteString("\n" + common.KeyValueView(v.URL, strconv.Itoa(v.Latency)+"ms"))
	}
	return s.String()
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	return m.help.View(defaultKeyMap)
}

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
