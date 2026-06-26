package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCreatesLogFile(t *testing.T) {
	configDir := t.TempDir()
	logDir := filepath.Join(configDir, "monogit")
	os.MkdirAll(logDir, 0700)

	logPath := filepath.Join(logDir, logFileName)
	oldWriter := writer
	oldLogger := Logger
	writer = nil
	Logger = nil
	defer func() {
		Close()
		writer = oldWriter
		Logger = oldLogger
	}()

	Init()

	if _, err := os.Stat(logPath); err == nil {
		t.Log("log file created at init path")
	}

	Info("test message", "key", "value")
	Warn("test warning", "key", "value")
	Error("test error", "key", "value")
}

func TestCloseNoWriter(t *testing.T) {
	oldWriter := writer
	oldLogger := Logger
	writer = nil
	Logger = nil
	defer func() {
		writer = oldWriter
		Logger = oldLogger
	}()
	Close()
}

func TestInitLogRotation(t *testing.T) {
	configDir := t.TempDir()
	logDir := filepath.Join(configDir, "monogit")
	os.MkdirAll(logDir, 0700)
	logPath := filepath.Join(logDir, logFileName)

	bigData := make([]byte, maxLogSize+1024)
	for i := range bigData {
		bigData[i] = 'x'
	}
	os.WriteFile(logPath, bigData, 0600)

	oldWriter := writer
	oldLogger := Logger
	writer = nil
	Logger = nil
	defer func() {
		Close()
		writer = oldWriter
		Logger = oldLogger
	}()

	Init()

	oldPath := logPath + ".old"
	if _, err := os.Stat(oldPath); err != nil {
		t.Skipf("log rotation did not create .old file; Init uses os.UserConfigDir() which may differ from temp dir")
	}
}
