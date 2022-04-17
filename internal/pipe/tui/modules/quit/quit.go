package quit

import (
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	_ modules.Module           = (*Module)(nil)
	_ modules.ModulePiper      = (*Module)(nil)
	_ simpleview.ContentViewer = (*Module)(nil)
)

type Module struct {
	D modules.Default
}

func NewModule(module *modules.Default) modules.Module {
	return &Module{
		D: modules.Default{
			PrePipe:  module.PrePipe,
			Pipe:     module.Pipe,
			PostPipe: module.PostPipe,
		},
	}
}

func (m *Module) String() string                    { return "quit module" }
func (m *Module) Init(ctx *context.Context) tea.Cmd { return tea.Quit }
func (m *Module) Update(msg tea.Msg) tea.Cmd        { return nil }
func (m *Module) SetContentSize(width, height int)  {}
func (m *Module) Content() string                   { return "" }
func (m *Module) Wrap() bool                        { return true }
func (m *Module) PrePipe() []modules.Piper          { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper             { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper         { return m.D.PostPipe }
func (m *Module) Cancel()                           {}
