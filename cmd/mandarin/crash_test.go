package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStderrIsKeptInTheCrashLog(t *testing.T) {
	dir := t.TempDir()
	stop := logStderr(dir)
	stop()
	if _, err := os.Stat(filepath.Join(dir, crashLog)); !os.IsNotExist(err) {
		t.Fatalf("a run with nothing on stderr made a crash log: %v", err)
	}
	stop = logStderr(dir)
	fmt.Fprintln(os.Stderr, "Caught panic: boom")
	stop()
	data, err := os.ReadFile(filepath.Join(dir, crashLog))
	if err != nil || !strings.Contains(string(data), "Caught panic: boom") {
		t.Errorf("crash log %q %v", data, err)
	}
}
