package logger

import (
	"testing"

	"github.com/jon4hz/deadshot/internal/context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	require.NoError(t, Log("foo", func(ctx *context.Context) error {
		return nil
	})(nil))
}

func TestTUILogger(t *testing.T) {
	require.Nil(t, LogTUI("foo", func(ctx *context.Context) tea.Cmd {
		return nil
	})(nil))
}
