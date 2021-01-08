package logging

import (
	"os"

	"go.uber.org/zap"
)

var (
	sugaredLogger *zap.SugaredLogger
	zlogger       *zap.Logger
)

func init() {
	mode := os.Getenv("BADGER_MODE")
	switch mode {
	case "test":
		zlogger = zap.NewExample()
	case "release":
		zlogger, _ = zap.NewProduction()
	default:
		zlogger, _ = zap.NewDevelopment()
	}
	sugaredLogger = zlogger.Sugar()
}

// Logger return a default logger to using
func Logger() *zap.SugaredLogger {
	return sugaredLogger
}

// Clean flushing log
func Clean() {
	_ = zlogger.Sync()
}
