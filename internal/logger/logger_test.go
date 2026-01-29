package logger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLogger(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer Close()

	SetLevel(DEBUG)
	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")

	logPath := filepath.Join(dir, "topology.log")
	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	s := string(b)
	if len(s) == 0 {
		t.Error("expected non-empty log file")
	}
	for _, sub := range []string{"DEBUG", "INFO", "WARN", "ERROR", "debug msg", "info msg", "warn msg", "error msg"} {
		if !contains(s, sub) {
			t.Errorf("log missing %q", sub)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestParseLevel(t *testing.T) {
	if parseLevel("INFO") != INFO {
		t.Error("INFO")
	}
	if parseLevel("debug") != DEBUG {
		t.Error("debug")
	}
	if parseLevel("") != INFO {
		t.Error("default")
	}
}
