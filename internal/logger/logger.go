package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level is logging level.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "?"
	}
}

func parseLevel(s string) Level {
	switch s {
	case "debug", "DEBUG":
		return DEBUG
	case "info", "INFO":
		return INFO
	case "warn", "WARN", "warning":
		return WARN
	case "error", "ERROR":
		return ERROR
	default:
		return INFO
	}
}

var (
	mu       sync.Mutex
	minLevel Level = INFO
	file     *os.File
)

// Init initializes the logger: creates logDir, opens topology.log for append, sets level from env TOPOLOGY_LOG_LEVEL (default INFO).
func Init(logDir string) error {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		return nil
	}
	if logDir == "" {
		return nil
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	logPath := filepath.Join(logDir, "topology.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	file = f
	if s := os.Getenv("TOPOLOGY_LOG_LEVEL"); s != "" {
		minLevel = parseLevel(s)
	}
	return nil
}

// SetLevel sets the minimum log level.
func SetLevel(l Level) {
	mu.Lock()
	defer mu.Unlock()
	minLevel = l
}

// Close closes the log file.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
		file = nil
	}
}

func logf(level Level, format string, args ...interface{}) {
	mu.Lock()
	if level < minLevel || file == nil {
		mu.Unlock()
		return
	}
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), level.String(), msg)
	_, _ = io.WriteString(file, line)
	mu.Unlock()
}

func Debug(format string, args ...interface{}) { logf(DEBUG, format, args...) }
func Info(format string, args ...interface{})  { logf(INFO, format, args...) }
func Warn(format string, args ...interface{})  { logf(WARN, format, args...) }
func Error(format string, args ...interface{}) { logf(ERROR, format, args...) }
