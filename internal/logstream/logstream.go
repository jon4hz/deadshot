package logstream

import (
	"fmt"
	"time"
)

type logType string

const (
	INFO logType = "Info"
	WARN logType = "Warn"
	ERR  logType = "Erro"
)

func Format(log string, logType logType) string {
	return fmt.Sprintf("%s %s: %s\n", time.Now().Format("15:04:05"), logType, log)
}
