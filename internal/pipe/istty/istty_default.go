//go:build !windows
// +build !windows

package istty

import (
	"os"

	"github.com/mattn/go-isatty"
)

func isTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
