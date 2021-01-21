package log

import (
	"os"

	"go.uber.org/zap"
)

var (
	// DefaultLogger is the default logger
	DefaultLogger Logger
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
	DefaultLogger = zlogger.Sugar()
}

// Logger interface of badger
type Logger interface {

	// Log a debug statement
	Debugf(format string, v ...interface{})

	// Log a info statement
	Infof(format string, v ...interface{})

	// Log a warning statement
	Warnf(format string, v ...interface{})

	// Log an error
	Errorf(format string, v ...interface{})

	// Log a fatal error
	Fatalf(format string, v ...interface{})
}

// Clean flushing log
func Clean() {
	_ = zlogger.Sync()
}
