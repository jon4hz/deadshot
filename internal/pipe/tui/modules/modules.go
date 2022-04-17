package modules

import (
	ctx "context"
	"fmt"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/ui/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Module interface {
	fmt.Stringer

	Init(ctx *context.Context) tea.Cmd
	Update(tea.Msg) tea.Cmd
	Cancel()
}

type ForkBackMsg int

type ForkMsg int

const (
	ForkMsgNone ForkMsg = iota
	ForkMsgQuit
	ForkMsgTrade
	ForkMsgSwap
	ForkMsgOrder
	ForkMsgSettings
	ForkMsgWalletSettings
	ForkMsgWalletSettingsNew
	ForkMsgWalletSettingsDerivation
	ForkMsgCustomEndpoint
)

type (
	InitMsg           struct{}
	BackMsg           struct{}
	NextMsg           struct{}
	PipeMsg           string
	PipeDoneMsg       struct{}
	ErrMsg            error
	ResizeMsg         struct{}
	PipeCancelFuncMsg ctx.CancelFunc
)

type Error struct {
	Message string
	Help    string
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Render(width int) string {
	head := style.ErrStyle.Render(fmt.Sprintf("%s. ", e.Message))
	body := style.SubtleStyle.Width(width - lipgloss.Width(head)).Render(e.Help)
	return lipgloss.NewStyle().Width(width).Render(head + body)
}

type Piper interface {
	fmt.Stringer

	Run(ctx *context.Context) error
}

type PipeMessenger interface {
	Message() string
}

type ModulePiper interface {
	PrePipe() []Piper
	Pipe() []Piper
	PostPipe() []Piper
}

type Canceler interface {
	CancelFunc() ctx.CancelFunc
}

type Forker interface {
	Fork() ForkMsg
}

type Default struct {
	ForkBackMsg ForkBackMsg
	Ctx         *context.Context
	PrePipe     []Piper
	Pipe        []Piper
	PostPipe    []Piper
}

// None is a placeholder to be returned from the module's init function.
// If the modules doesn't need to do anything special,
// it must return None instead of nil.
var None = func() tea.Msg { return nil }

// Init initialized the first module of the pipeline.
var Init = func() tea.Msg { return InitMsg{} }

// Next loads the next module in the pipeline.
var Next = func() tea.Msg { return NextMsg{} }

// Back loads the previous module in the pipeline.
var Back = func() tea.Msg { return BackMsg{} }

// Resize can be used, to manually trigger a resize of the tui components.
var Resize = func() tea.Msg { return ResizeMsg{} }
