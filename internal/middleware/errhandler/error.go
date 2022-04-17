package errhandler

import (
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/middleware"
	"github.com/jon4hz/deadshot/internal/pipe"
)

// Handle handles an action error, ignoring and logging pipe skipped
// errors.
func Handle(action middleware.Action) middleware.Action {
	return func(ctx *context.Context) error {
		err := action(ctx)
		if err == nil {
			return nil
		}
		if pipe.IsSkip(err) {
			logging.Log.WithError(err).Warn("pipe skipped")
			return nil
		}
		return err
	}
}
