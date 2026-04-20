package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Setup configures the default slog logger based on the given level name.
// Valid levels are: debug, info, warn, error. Logs go to stderr so they
// don't interfere with the primary stdout output (reports, JSON).
func Setup(level string) error {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		return fmt.Errorf("unknown log level %q (expected debug|info|warn|error)", level)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(handler))
	return nil
}
