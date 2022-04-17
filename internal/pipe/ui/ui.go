package ui

import (
	"fmt"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/middleware/logger"
	"github.com/jon4hz/deadshot/internal/pipe/tui"
)

type Pipe struct{}

func (Pipe) String() string {
	return "selecting ui"
}

func (Pipe) Run(ctx *context.Context) error {
	tui := tui.Pipe{}
	if err := logger.Log(
		tui.String(),
		tui.Run,
	)(ctx); err != nil {
		return fmt.Errorf("%s: failed to run ui: %w", tui.String(), err)
	}
	return nil
}
