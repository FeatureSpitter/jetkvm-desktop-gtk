package logging

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureDefaultLevel(t *testing.T) {
	if err := Configure(""); err != nil {
		t.Fatalf("Configure empty: %v", err)
	}
}

func TestConfigureExplicitLevel(t *testing.T) {
	if err := Configure("debug"); err != nil {
		t.Fatalf("Configure debug: %v", err)
	}
}

func TestConfigureInvalidLevel(t *testing.T) {
	if err := Configure("bogus"); err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestSubsystem(t *testing.T) {
	logger := Subsystem("test-component")
	if logger.GetLevel() > base.GetLevel() {
		t.Error("subsystem logger has unexpected level")
	}
}

func resetFileOut(t *testing.T, origResolve func() (string, error)) {
	t.Helper()
	resolveLogPathFunc = origResolve
	if fileOut != nil {
		fileOut.Close()
		fileOut = nil
	}
	logPath = ""
}

func TestFileLogging_WritesAndAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	origResolve := resolveLogPathFunc
	resolveLogPathFunc = func() (string, error) { return path, nil }
	t.Cleanup(func() { resetFileOut(t, origResolve) })

	fileOut = nil
	logPath = ""
	if err := Configure("info"); err != nil {
		t.Fatalf("Configure: %v", err)
	}

	logger := Subsystem("test")
	logger.Info().Msg("first message")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "first message") {
		t.Errorf("log file missing 'first message', got:\n%s", data)
	}

	logger.Info().Msg("second message")
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "first message") {
		t.Error("first message disappeared after second write (not appending)")
	}
	if !strings.Contains(string(data), "second message") {
		t.Error("second message not in log file")
	}
}

func TestFileLogging_StdlibAlsoWritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	origResolve := resolveLogPathFunc
	resolveLogPathFunc = func() (string, error) { return path, nil }
	t.Cleanup(func() { resetFileOut(t, origResolve) })

	fileOut = nil
	logPath = ""
	if err := Configure("info"); err != nil {
		t.Fatalf("Configure: %v", err)
	}

	log.Printf("stdlib test line xyz123")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "xyz123") {
		t.Errorf("stdlib log.Printf not in file, got:\n%s", data)
	}
}

func TestFileLogging_SurvivesReconfigure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	origResolve := resolveLogPathFunc
	resolveLogPathFunc = func() (string, error) { return path, nil }
	t.Cleanup(func() { resetFileOut(t, origResolve) })

	fileOut = nil
	logPath = ""

	if err := Configure("info"); err != nil {
		t.Fatal(err)
	}
	la := Subsystem("a")
	la.Info().Msg("before reconfig")

	if err := Configure("debug"); err != nil {
		t.Fatal(err)
	}
	lb := Subsystem("b")
	lb.Debug().Msg("after reconfig")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "before reconfig") {
		t.Error("missing 'before reconfig'")
	}
	if !strings.Contains(content, "after reconfig") {
		t.Error("missing 'after reconfig'")
	}
}

func TestLogFilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	origResolve := resolveLogPathFunc
	resolveLogPathFunc = func() (string, error) { return path, nil }
	t.Cleanup(func() { resetFileOut(t, origResolve) })

	fileOut = nil
	logPath = ""
	Configure("info")

	got := LogFilePath()
	if got != path {
		t.Errorf("LogFilePath() = %q, want %q", got, path)
	}
}
