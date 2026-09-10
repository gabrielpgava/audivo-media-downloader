//go:build !windows

package platform

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

func ConfigureCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func KillProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	pid := cmd.Process.Pid
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
		return err
	}
	timer := time.NewTimer(750 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	case <-processExited(cmd.Process):
	}
	return nil
}

func processExited(process *os.Process) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			if err := process.Signal(syscall.Signal(0)); err != nil {
				close(done)
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
	}()
	return done
}
