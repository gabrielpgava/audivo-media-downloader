package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunnerCapturesStdoutAndStderrWithoutShell(t *testing.T) {
	if os.Getenv("AUDIVO_FAKE_PROCESS") == "1" {
		println("AUDIVO_PROGRESS\\t47.2\\t8.1 MiB/s\\t12")
		fmt.Fprintln(os.Stderr, "diagnostic")
		return
	}
	t.Setenv("AUDIVO_FAKE_PROCESS", "1")
	ctx := context.Background()
	var lines []Line
	err := NewRunner().Run(ctx, os.Args[0], []string{"-test.run=TestRunnerCapturesStdoutAndStderrWithoutShell", "--"}, t.TempDir(), func(line Line) {
		lines = append(lines, line)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Fatal("expected fake process output")
	}
	joined := make([]string, 0, len(lines))
	for _, line := range lines {
		joined = append(joined, line.Stream+":"+line.Message)
	}
	joinedOutput := strings.Join(joined, "\n")
	if !strings.Contains(joinedOutput, "AUDIVO_PROGRESS") || !strings.Contains(joinedOutput, "diagnostic") {
		t.Fatalf("unexpected lines: %v", joined)
	}
}

func TestRunnerReturnsFailureAndPreservesStderr(t *testing.T) {
	if os.Getenv("AUDIVO_FAKE_PROCESS_FAIL") == "1" {
		fmt.Fprintln(os.Stderr, "fake engine failed")
		os.Exit(7)
	}
	t.Setenv("AUDIVO_FAKE_PROCESS_FAIL", "1")
	var mu sync.Mutex
	var lines []Line
	err := NewRunner().Run(context.Background(), os.Args[0], []string{"-test.run=TestRunnerReturnsFailureAndPreservesStderr", "--"}, t.TempDir(), func(line Line) {
		mu.Lock()
		lines = append(lines, line)
		mu.Unlock()
	})
	if err == nil {
		t.Fatal("expected fake engine failure")
	}
	joined := make([]string, 0, len(lines))
	for _, line := range lines {
		joined = append(joined, line.Stream+":"+line.Message)
	}
	if !strings.Contains(strings.Join(joined, "\n"), "stderr:fake engine failed") {
		t.Fatalf("stderr was not retained: %v", joined)
	}
}

func TestRunnerCancellationReturnsContextError(t *testing.T) {
	if os.Getenv("AUDIVO_FAKE_PROCESS") == "1" {
		time.Sleep(5 * time.Second)
		return
	}
	t.Setenv("AUDIVO_FAKE_PROCESS", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	err := NewRunner().Run(ctx, os.Args[0], []string{"-test.run=TestRunnerCancellationReturnsContextError", "--"}, t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestRunnerStopsControlledChildProcess(t *testing.T) {
	if os.Getenv("AUDIVO_CHILD_MODE") == "1" {
		marker := os.Getenv("AUDIVO_CHILD_MARKER")
		time.Sleep(2 * time.Second)
		_ = os.WriteFile(marker, []byte("survived"), 0o600)
		return
	}
	if os.Getenv("AUDIVO_FAKE_PROCESS_CHILD") == "1" {
		marker := os.Getenv("AUDIVO_CHILD_MARKER")
		child := exec.Command(os.Args[0], "-test.run=TestRunnerStopsControlledChildProcess", "--")
		child.Env = append(os.Environ(), "AUDIVO_CHILD_MODE=1", "AUDIVO_CHILD_MARKER="+marker)
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		_ = os.WriteFile(marker, []byte("started"), 0o600)
		time.Sleep(10 * time.Second)
		return
	}

	marker := filepath.Join(t.TempDir(), "child-marker")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Setenv("AUDIVO_FAKE_PROCESS_CHILD", "1")
	t.Setenv("AUDIVO_CHILD_MARKER", marker)
	done := make(chan error, 1)
	go func() {
		done <- NewRunner().Run(ctx, os.Args[0], []string{"-test.run=TestRunnerStopsControlledChildProcess", "--"}, t.TempDir(), nil)
	}()
	deadline := time.Now().Add(time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("fake child did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	if err := <-done; err == nil {
		t.Fatal("expected cancellation error")
	}
	time.Sleep(700 * time.Millisecond)
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "survived" {
		t.Fatal("controlled child survived cancellation")
	}
}
