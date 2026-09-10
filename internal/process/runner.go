package process

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os/exec"
	"sync"

	"audivo-media-downloader/internal/platform"
)

type Line struct {
	Stream  string
	Message string
}

type Runner struct{}

func NewRunner() *Runner { return &Runner{} }

// Run executes one external process with direct arguments. It never invokes a
// shell. stdout and stderr are scanned independently so structured output can
// be parsed without losing diagnostic lines.
func (r *Runner) Run(ctx context.Context, executable string, args []string, dir string, onLine func(Line)) error {
	if executable == "" {
		return errors.New("executable is required")
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	platform.ConfigureCommand(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	finished := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = platform.KillProcessTree(cmd)
		case <-finished:
		}
	}()

	var readGroup sync.WaitGroup
	var callbackMu sync.Mutex
	emitLine := func(line Line) {
		if onLine == nil {
			return
		}
		// stdout and stderr are consumed concurrently, but callers should not
		// need their own lock just to collect the diagnostic stream.
		callbackMu.Lock()
		defer callbackMu.Unlock()
		onLine(line)
	}
	readGroup.Add(2)
	read := func(stream string, reader io.Reader) {
		defer readGroup.Done()
		scanner := bufio.NewScanner(reader)
		// Engine diagnostics can contain long JSON/URLs. Keep the scanner
		// bounded but substantially above its default 64 KiB limit.
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			emitLine(Line{Stream: stream, Message: scanner.Text()})
		}
	}
	go read("stdout", stdout)
	go read("stderr", stderr)

	waitErr := cmd.Wait()
	close(finished)
	readGroup.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return waitErr
}
