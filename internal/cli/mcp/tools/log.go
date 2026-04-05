package tools

import (
	"log/slog"
	"time"
)

func logToolStart(name string, attrs ...any) time.Time {
	slog.Info("mcp tool started", append([]any{"tool", name}, attrs...)...)
	return time.Now()
}

func logToolSuccess(name string, start time.Time, attrs ...any) {
	slog.Info("mcp tool completed", append([]any{"tool", name, "duration_ms", time.Since(start).Milliseconds()}, attrs...)...)
}

func logToolError(name string, start time.Time, err error, attrs ...any) {
	slog.Error("mcp tool failed", append([]any{"tool", name, "duration_ms", time.Since(start).Milliseconds(), "error", err}, attrs...)...)
}
