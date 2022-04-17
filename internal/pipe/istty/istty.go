package istty

import (
	"errors"

	"github.com/jon4hz/deadshot/internal/context"
)

type Pipe struct{}

func (Pipe) String() string {
	return "checking if program runs in a terminal"
}

func (Pipe) Run(ctx *context.Context) error {
	if !isTTY() {
		return errors.New("please execute in a terminal")
	}
	return nil
}
