package logger

import (
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/middleware"

	tea "github.com/charmbracelet/bubbletea"
)

func Log(title string, next middleware.Action) middleware.Action {
	return func(ctx *context.Context) error {
		logging.Log.Info(title)
		return next(ctx)
	}
}

func LogTUI(title string, next middleware.TUIAction) middleware.TUIAction {
	return func(ctx *context.Context) tea.Cmd {
		logging.Log.Info(title)
		return next(ctx)
	}
}
