package tuihandler

import (
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/middleware"
	"github.com/jon4hz/deadshot/internal/middleware/skip"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"

	tea "github.com/charmbracelet/bubbletea"
)

func Init(action middleware.TUIAction) middleware.TUIAction {
	return func(ctx *context.Context) tea.Cmd {
		cmd := action(ctx)
		return cmd
	}
}

func Skip(skipper any, next middleware.TUIAction) middleware.TUIAction {
	if skipper, ok := skipper.(skip.Skipper); ok {
		return func(ctx *context.Context) tea.Cmd {
			if skipper.Skip(ctx) {
				logging.Log.Debugf("skipped %s", skipper.String())
				return nil
			}
			cmd := next(ctx)
			return cmd
		}
	}
	return next
}

func SetSizes(module any, headerWidth, contentWidth, contentHeight, footerWidth int, next middleware.TUIAction) middleware.TUIAction {
	switch module := module.(type) {
	case simpleview.SimpleViewer:
		module.SetHeaderWidth(headerWidth)
		module.SetContentSize(contentWidth, contentHeight)
		module.SetFooterWidth(footerWidth)
	case simpleview.ContentViewer:
		module.SetContentSize(contentWidth, contentHeight)
	}
	return next
}
