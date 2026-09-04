package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	maxLogSize  = 2 * 1024 * 1024
	logFileName = "monogit.log"
)

var (
	Logger *slog.Logger
	mu     sync.Mutex
	writer io.WriteCloser
)

func Init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config")
	}
	logDir := filepath.Join(configDir, "monogit")

	if err := os.MkdirAll(logDir, 0700); err != nil {
		Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		return
	}

	logPath := filepath.Join(logDir, logFileName)

	fi, statErr := os.Stat(logPath)
	if statErr == nil && fi.Size() > maxLogSize {
		oldPath := logPath + ".old"
		_ = os.Rename(logPath, oldPath)
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		return
	}

	mu.Lock()
	writer = f
	mu.Unlock()

	level := slog.LevelInfo
	if os.Getenv("MONOGIT_DEBUG") != "" {
		level = slog.LevelDebug
	}

	Logger = slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: level,
	}))
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if writer != nil {
		_ = writer.Close()
		writer = nil
	}
}

func Debug(msg string, args ...any) {
	if l := Logger; l != nil {
		l.Debug(msg, args...)
	}
}

func Info(msg string, args ...any) {
	if l := Logger; l != nil {
		l.Info(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if l := Logger; l != nil {
		l.Error(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if l := Logger; l != nil {
		l.Warn(msg, args...)
	}
}
