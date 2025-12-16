package log

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Lgr     *zap.Logger
	Verbose bool
	LogPath string
)

// Init intializes global logger, and returns with the cleanup functon.
func Init() func() {
	// Open the log file once, to be shared by both loggers
	logFile, err := os.OpenFile(LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	Must("open log file", err)

	// Configure zap to use the same file handle
	cfg := zap.NewDevelopmentConfig()
	if !Verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	// Create a core that writes to our file handle instead of letting zap open it
	ws := zapcore.AddSync(logFile)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg.EncoderConfig),
		ws,
		cfg.Level,
	)
	Lgr = zap.New(core)
	zap.ReplaceGlobals(Lgr)

	// Redirect standard library log package to the same file
	log.SetOutput(logFile)

	return func() {
		Must("log sync", Lgr.Sync())
		Must("close log file", logFile.Close())
	}
}

// Must panics if the given error is not nil, with the given description. For initializations only.
func Must(descr string, err error) {
	if err != nil {
		panic(fmt.Errorf("%s: %w", descr, err))
	}
}
