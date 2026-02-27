package logger

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

func New(debug bool) *log.Logger {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.DateTime,
	})
	if debug {
		logger.SetLevel(log.DebugLevel)
	}
	return logger
}
