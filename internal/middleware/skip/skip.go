package skip

import (
	"fmt"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/middleware"
)

// Skipper defines a method to skip an entire Piper.
type Skipper interface {
	// Skip returns true if the Piper should be skipped.
	Skip(ctx *context.Context) bool
	fmt.Stringer
}

// Maybe returns an action that skips immediately if the given p is a Skipper
// and its Skip method returns true.
func Maybe(skipper any, next middleware.Action) middleware.Action {
	if skipper, ok := skipper.(Skipper); ok {
		return func(ctx *context.Context) error {
			if skipper.Skip(ctx) {
				logging.Log.Debugf("skipped %s", skipper.String())
				return nil
			}
			return next(ctx)
		}
	}
	return next
}
