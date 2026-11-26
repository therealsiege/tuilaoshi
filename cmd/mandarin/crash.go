package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// crashLog is where a crash's details go, in the wiki folder (ignored by
// git): when a screen panics, Bubble Tea restores the terminal and writes
// the stack to stderr, which would scroll away with the window.
const crashLog = "crash.log"

// logStderr copies everything written to stderr into the wiki's crash log
// as well as the terminal, the file only opened once something is written.
// The returned func stops the copy and waits for it to finish.
func logStderr(dir string) (stop func()) {
	r, w, err := os.Pipe()
	if err != nil {
		return func() {}
	}
	terminal := os.Stderr
	os.Stderr = w
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		var log *os.File
		for {
			n, err := r.Read(buf)
			if n > 0 {
				_, _ = terminal.Write(buf[:n]) //nolint:errcheck // the terminal is best effort
				if log == nil {
					log, _ = os.OpenFile(filepath.Join(dir, crashLog), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600) //nolint:errcheck // no log beats no game
					if log != nil {
						fmt.Fprintf(log, "\n== %s ==\n", time.Now().Format(time.RFC3339)) //nolint:errcheck // best effort
					}
				}
				if log != nil {
					_, _ = log.Write(buf[:n]) //nolint:errcheck // best effort
				}
			}
			if err == io.EOF || err != nil {
				break
			}
		}
		if log != nil {
			_ = log.Close() //nolint:errcheck // best effort
		}
	}()
	return func() {
		os.Stderr = terminal
		_ = w.Close() //nolint:errcheck // ends the copy
		<-done
	}
}
