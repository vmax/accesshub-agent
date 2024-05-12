package internal

import (
	"log/slog"
	"os"
)

var (
	alog *slog.Logger
)

func init() {
	alog = slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func Info(msg string, keysAndValues ...interface{}) {
	alog.Info(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	alog.Error(msg, keysAndValues...)
}
