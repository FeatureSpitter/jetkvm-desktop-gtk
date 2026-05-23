package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultLevel = zerolog.InfoLevel
	envLogLevel  = "JETKVM_DESKTOP_LOG_LEVEL"
	maxLogSizeMB = 100
	maxBackups   = 1
)

var (
	mu               sync.RWMutex
	base             zerolog.Logger
	logPath          string
	fileOut          *lumberjack.Logger
	resolveLogPathFunc = resolveLogPath
)

func init() {
	base = buildLogger(defaultLevel)
}

func Configure(levelText string) error {
	level := defaultLevel
	if value := strings.TrimSpace(levelText); value != "" {
		parsed, err := zerolog.ParseLevel(strings.ToLower(value))
		if err != nil {
			return fmt.Errorf("invalid log level %q: %w", value, err)
		}
		level = parsed
	} else if value := strings.TrimSpace(os.Getenv(envLogLevel)); value != "" {
		parsed, err := zerolog.ParseLevel(strings.ToLower(value))
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", envLogLevel, value, err)
		}
		level = parsed
	}

	mu.Lock()
	base = buildLogger(level)
	mu.Unlock()
	return nil
}

func Subsystem(name string) zerolog.Logger {
	mu.RLock()
	logger := base.With().Str("component", name).Logger()
	mu.RUnlock()
	return logger
}

// LogFilePath returns the path to the current log file, or "" if file
// logging could not be initialised.
func LogFilePath() string {
	mu.RLock()
	defer mu.RUnlock()
	return logPath
}

func buildLogger(level zerolog.Level) zerolog.Logger {
	console := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	fileConsole := initFileWriter()

	var zerologOut io.Writer
	if fileConsole != nil {
		zerologOut = zerolog.MultiLevelWriter(console, fileConsole)
	} else {
		zerologOut = console
	}

	return zerolog.New(zerologOut).Level(level).With().Timestamp().Logger()
}

func initFileWriter() *zerolog.ConsoleWriter {
	if fileOut != nil {
		redirectStdlib(fileOut)
		w := zerolog.ConsoleWriter{
			Out:        fileOut,
			TimeFormat: time.RFC3339,
			NoColor:    true,
		}
		return &w
	}

	p, err := resolveLogPathFunc()
	if err != nil {
		return nil
	}

	fileOut = &lumberjack.Logger{
		Filename:   p,
		MaxSize:    maxLogSizeMB,
		MaxBackups: maxBackups,
		Compress:   false,
	}
	logPath = p

	redirectStdlib(fileOut)
	w := zerolog.ConsoleWriter{
		Out:        fileOut,
		TimeFormat: time.RFC3339,
		NoColor:    true,
	}
	return &w
}

func redirectStdlib(w io.Writer) {
	combined := io.MultiWriter(os.Stderr, w)
	log.SetOutput(combined)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}

func resolveLogPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "jetkvm-desktop")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "jetkvm-desktop.log"), nil
}
