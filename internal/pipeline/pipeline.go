package pipeline

import (
	"fmt"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/istty"
	"github.com/jon4hz/deadshot/internal/pipe/ui"
)

type Piper interface {
	fmt.Stringer

	// Run the pipe
	Run(ctx *context.Context) error
}

var NewPipeline = func() []Piper {
	return []Piper{
		istty.Pipe{},
		ui.Pipe{},
	}
}
