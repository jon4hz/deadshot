package pipe

import "github.com/jon4hz/deadshot/internal/context"

type Pipe struct{}

func (Pipe) String() string { return "default" }

func (Pipe) Run(ctx *context.Context) error {
	return nil
}
