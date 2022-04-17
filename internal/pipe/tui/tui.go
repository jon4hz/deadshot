package tui

import (
	ctx "context"
	"errors"
	"sync"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/middleware/logger"
	"github.com/jon4hz/deadshot/internal/middleware/skip"
	"github.com/jon4hz/deadshot/internal/middleware/tuihandler"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"

	tea "github.com/charmbracelet/bubbletea"
)

type postPipeDoneMsg struct{}

type Tui struct {
	cm                 modules.Module
	modules            []modules.Module
	index              int
	ctx                *context.Context
	pipeMsgChan        chan string
	pipeCancelFuncChan chan ctx.CancelFunc

	height int
	width  int
	view   *simpleview.Model
	mu     *sync.Mutex
}

func newTui(c *context.Context) Tui {
	t := Tui{
		modules:            newTuiPipeline(),
		ctx:                c,
		view:               simpleview.NewModel(),
		pipeMsgChan:        make(chan string),
		pipeCancelFuncChan: make(chan ctx.CancelFunc),
		mu:                 &sync.Mutex{},
	}
	t.index = 1
	t.cm = t.modules[t.index]
	return t
}

func (t Tui) Init() tea.Cmd {
	return tea.Batch(
		pipeMsgListener(t.pipeMsgChan),
		pipeCancelFuncListener(t.pipeCancelFuncChan),
		modules.Init,
	)
}

func pipeMsgListener(pipeMsgChan chan string) tea.Cmd {
	return func() tea.Msg {
		// TODO: add context to make it possible to cancel the listener
		return modules.PipeMsg(<-pipeMsgChan)
	}
}

func pipeCancelFuncListener(pipeCancelFuncChan chan ctx.CancelFunc) tea.Cmd {
	return func() tea.Msg {
		// TODO: add context to make it possible to cancel the listener
		return modules.PipeCancelFuncMsg(<-pipeCancelFuncChan)
	}
}

func (t Tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case modules.InitMsg:
		t.mu.Lock()

		err := t.runPrePipeline()
		if err != nil {
			if !errors.Is(err, ErrNoPipeline) {
				logging.Log.WithField("err", err).Error("Error running pre pipeline")
				cmds = append(cmds, func() tea.Msg { return modules.ErrMsg(err) })
			}
		}

		if cmd := tuihandler.Skip( // nolint: dupl
			t.cm,

			tuihandler.SetSizes(
				t.cm,
				t.view.GetHeaderWidth(),
				t.view.GetContentWidth(),
				t.view.GetContentHeight(t.cm),
				t.view.GetFooterWidth(),

				tuihandler.Init(
					logger.LogTUI(
						t.cm.String(),
						t.cm.Init,
					),
				),
			),
		)(t.ctx); cmd != nil {
			cmds = append(cmds, cmd)
			cmds = append(cmds,
				func() tea.Msg {
					defer t.mu.Unlock()
					if err := t.runPipeline(); err != nil {
						if !errors.Is(err, ErrNoPipeline) {
							logging.Log.WithField("err", err).Error("Error running pipeline")
							return modules.ErrMsg(err)
						}
						return nil
					}
					return modules.PipeDoneMsg{}
				},
			)
		} else {
			defer t.mu.Unlock()
			cmds = append(cmds, modules.Next)
		}

		return t, tea.Batch(cmds...)

	case modules.BackMsg:
		t.mu.Lock()

		logging.Log.WithField("from", t.cm.String()).Debug("BackMsg")
		t.cm.Cancel()
		if m := t.moduleBack(); m != nil {
			t.cm = m
		} else {
			return t, nil
		}

		err := t.runPrePipeline()
		if err != nil {
			if !errors.Is(err, ErrNoPipeline) {
				logging.Log.WithField("err", err).Error("Error running pre pipeline")
				cmds = append(cmds, func() tea.Msg { return modules.ErrMsg(err) })
			}
		}

		if cmd := tuihandler.Skip( // nolint: dupl
			t.cm,

			tuihandler.SetSizes(
				t.cm,
				t.view.GetHeaderWidth(),
				t.view.GetContentWidth(),
				t.view.GetContentHeight(t.cm),
				t.view.GetFooterWidth(),

				tuihandler.Init(
					logger.LogTUI(
						t.cm.String(),
						t.cm.Init,
					),
				),
			),
		)(t.ctx); cmd != nil {
			cmds = append(cmds, cmd)

			cmds = append(cmds,
				func() tea.Msg {
					defer t.mu.Unlock()
					if err := t.runPipeline(); err != nil {
						if !errors.Is(err, ErrNoPipeline) {
							logging.Log.WithField("err", err).Error("Error running pipeline")
							return modules.ErrMsg(err)
						}
						return nil
					}
					return modules.PipeDoneMsg{}
				},
			)
		} else {
			defer t.mu.Unlock()
			cmds = append(cmds, modules.Back)
		}

		return t, tea.Batch(cmds...)

	case modules.NextMsg:
		t.mu.Lock()

		logging.Log.WithField("from", t.cm.String()).Debug("NextMsg")

		cmds = append(cmds, func() tea.Msg {
			defer t.mu.Unlock()
			if err := t.runPostPipeline(); err != nil {
				if !errors.Is(err, ErrNoPipeline) {
					logging.Log.WithField("err", err).Error("Error running post pipeline")
					return modules.ErrMsg(err)
				}
			}
			return postPipeDoneMsg{}
		})

		return t, tea.Batch(cmds...)

	case postPipeDoneMsg:
		t.mu.Lock()

		logging.Log.WithField("from", t.cm.String()).Debug("postPipeDoneMsg")

		t.cm.Cancel()
		if m := t.moduleForward(); m != nil {
			t.cm = m
		} else {
			return t, nil
		}

		err := t.runPrePipeline()
		if err != nil {
			if !errors.Is(err, ErrNoPipeline) {
				logging.Log.WithField("err", err).Error("Error running pre pipeline")
				cmds = append(cmds, func() tea.Msg { return modules.ErrMsg(err) })
			}
		}

		if cmd := tuihandler.Skip( // nolint: dupl
			t.cm,

			tuihandler.SetSizes(
				t.cm,
				t.view.GetHeaderWidth(),
				t.view.GetContentWidth(),
				t.view.GetContentHeight(t.cm),
				t.view.GetFooterWidth(),

				tuihandler.Init(
					logger.LogTUI(
						t.cm.String(),
						t.cm.Init,
					),
				),
			),
		)(t.ctx); cmd != nil {
			cmds = append(cmds, cmd)
			cmds = append(cmds,
				func() tea.Msg {
					defer t.mu.Unlock()
					if err := t.runPipeline(); err != nil {
						if !errors.Is(err, ErrNoPipeline) {
							logging.Log.WithField("err", err).Error("Error running pipeline")
							return modules.ErrMsg(err)
						}
						return nil
					}
					return modules.PipeDoneMsg{}
				},
			)
		} else {
			defer t.mu.Unlock()
			cmds = append(cmds, modules.Next)
		}

		return t, tea.Batch(cmds...)

	case modules.ForkMsg:
		if cmd := t.handleFork(msg); cmd != nil {
			return t, cmd
		}

	case modules.ForkBackMsg:
		logging.Log.WithField("ForkBackMsg", msg).Debug("return pipeline")
		t.modules = t.modules[:len(t.modules)-int(msg)]
		return t, modules.Back

	case modules.PipeMsg:
		cmds = append(cmds, pipeMsgListener(t.pipeMsgChan))

	case modules.PipeCancelFuncMsg:
		cmds = append(cmds, pipeCancelFuncListener(t.pipeCancelFuncChan))

	case tea.WindowSizeMsg:
		top, right, bottom, left := style.DocStyle.GetMargin()
		t.height = msg.Height - top - bottom
		t.width = msg.Width - right - left
		t.view.SetSize(t.width, t.height)

		t.setSimpleviewerSizes()

	case modules.ResizeMsg:
		t.setSimpleviewerSizes()
	}

	cmds = append(cmds, t.cm.Update(msg))
	return t, tea.Batch(cmds...)
}

func (t *Tui) handleFork(msg modules.ForkMsg) tea.Cmd {
	switch msg {
	case modules.ForkMsgTrade:
		logging.Log.WithField("ForkMsg", "trade").Debug("new pipeline")
		t.modules = append(t.modules, newTradePipeline()...)
		return modules.Next

	case modules.ForkMsgOrder:
		logging.Log.WithField("ForkMsg", "order").Debug("new pipeline")
		t.modules = append(t.modules, newOrderPipeline()...)
		return modules.Next

	case modules.ForkMsgSwap:
		logging.Log.WithField("ForkMsg", "swap").Debug("new pipeline")
		t.modules = append(t.modules, newSwapPipeline()...)
		return modules.Next

	case modules.ForkMsgSettings:
		logging.Log.WithField("ForkMsg", "settings").Debug("new pipeline")
		t.modules = append(t.modules, newSettingsPipeline()...)
		return modules.Next

	case modules.ForkMsgCustomEndpoint:
		logging.Log.WithField("ForkMsg", "settings endpoint").Debug("new pipeline")
		t.modules = append(t.modules, newSettingsEndpointPipeline()...)
		return modules.Next

	case modules.ForkMsgWalletSettings:
		logging.Log.WithField("ForkMsg", "settings wallet").Debug("new pipeline")
		t.modules = append(t.modules, newSettingsWalletPipeline()...)
		return modules.Next

	case modules.ForkMsgWalletSettingsNew:
		logging.Log.WithField("ForkMsg", "settings wallet new").Debug("new pipeline")
		t.modules = append(t.modules, newSettingsWalletNewPipeline()...)
		return modules.Next

	case modules.ForkMsgWalletSettingsDerivation:
		logging.Log.WithField("ForkMsg", "settings wallet derivation").Debug("new pipeline")
		t.modules = append(t.modules, newSettingsWalletDerivationPipeline()...)
		return modules.Next

	case modules.ForkMsgQuit:
		logging.Log.WithField("ForkMsg", "quit").Debug("new pipeline")
		t.modules = append(t.modules, newQuit())
		return modules.Next
	}

	return nil
}

func (t *Tui) setSimpleviewerSizes() {
	switch module := t.cm.(type) {
	case simpleview.SimpleViewer:
		module.SetHeaderWidth(t.view.GetHeaderWidth())
		module.SetContentSize(t.view.GetContentWidth(), t.view.GetContentHeight(t.cm))
		module.SetFooterWidth(t.view.GetFooterWidth())
	case simpleview.ContentViewer:
		module.SetContentSize(t.view.GetContentWidth(), t.view.GetFullContentHeight())
	}
}

func (t *Tui) moduleForward() modules.Module {
	if t.index < len(t.modules)-1 {
		t.index++
		return t.modules[t.index]
	}
	return nil
}

func (t *Tui) moduleBack() modules.Module {
	if t.index > 0 {
		t.index--
		return t.modules[t.index]
	}
	return nil
}

func (t *Tui) runPrePipeline() error {
	pipe, ok := t.cm.(modules.ModulePiper)
	if !ok {
		return ErrNoPipeline
	}
	return t.runPipe(pipe.PrePipe())
}

var ErrNoPipeline = errors.New("no pipeline")

func (t *Tui) runPipeline() error {
	pipe, ok := t.cm.(modules.ModulePiper)
	if !ok {
		return ErrNoPipeline
	}
	return t.runPipe(pipe.Pipe())
}

func (t *Tui) runPostPipeline() error {
	pipe, ok := t.cm.(modules.ModulePiper)
	if !ok {
		return ErrNoPipeline
	}
	return t.runPipe(pipe.PostPipe())
}

func (t *Tui) runPipe(pipeline []modules.Piper) error {
	if pipeline == nil || len(pipeline) == 0 {
		return ErrNoPipeline
	}
	for _, pipe := range pipeline {
		messenger, ok := pipe.(modules.PipeMessenger)
		if ok {
			if msg := messenger.Message(); msg != "" {
				t.pipeMsgChan <- msg
			}
		}
		cancel, ok := pipe.(modules.Canceler)
		if ok {
			if cancel := cancel.CancelFunc(); cancel != nil {
				t.pipeCancelFuncChan <- cancel
			}
		}
		if err := skip.Maybe(
			pipe,
			logger.Log(
				pipe.String(),
				pipe.Run,
			),
		)(t.ctx); err != nil {
			return err
		}
	}
	return nil
}

func (t Tui) View() string {
	return style.DocStyle.Render(t.view.View(t.cm))
}
