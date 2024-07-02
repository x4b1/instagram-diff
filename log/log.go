package log

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

var _ Logger = (*logger)(nil)

type logger struct {
	*slog.Logger
}

func (l *logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}

func NewLogger() Logger {
	return &logger{slog.New(tint.NewHandler(os.Stderr, nil))}
}

type Logger interface {
	Info(msg string, args ...any)
	Fatal(msg string, args ...any)
}
