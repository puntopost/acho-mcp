package cli

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/puntopost/acho-mcp/internal/cli/config"
)

const maxLogSize = 1 << 20 // 1 MB

// InitLogger sets up the global slog logger to write to ~/.acho/acho.log.
// The current log is rotated to acho.log.1 when it exceeds 1 MB.
func InitLogger() {
	path := filepath.Join(config.DefaultDir(), "acho.log")

	w := &truncWriter{path: path}
	h := slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))
}

type truncWriter struct {
	path string
	mu   sync.Mutex
}

func (w *truncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	info, err := os.Stat(w.path)
	if err == nil && info.Size() > maxLogSize {
		if err := rotateLog(w.path); err != nil {
			return 0, err
		}
	}

	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Write(p)
}

func rotateLog(path string) error {
	prevPath := path + ".1"
	if err := os.Remove(prevPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(path, prevPath); err != nil {
		return err
	}
	return nil
}

var _ io.Writer = (*truncWriter)(nil)
