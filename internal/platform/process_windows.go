//go:build windows

package platform

import (
	"os/exec"
	"strconv"
)

func ConfigureCommand(_ *exec.Cmd) {}

func KillProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return exec.Command("taskkill.exe", "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F").Run()
}
