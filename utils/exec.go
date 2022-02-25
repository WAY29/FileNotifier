package utils

import (
	"os/exec"
	"runtime"
)

var (
	IsWindows   = false
	ShellPath   string
	ShellRunArg string
)

func init() {
	if runtime.GOOS == "windows" {
		IsWindows = true
		ShellPath, _ = exec.LookPath("cmd.exe")
		ShellRunArg = "/c"
	} else {
		ShellPath, _ = exec.LookPath("sh")
		ShellRunArg = "-c"
	}
}

func ExecCommand(command string) (*exec.Cmd, error) {
	args := []string{ShellRunArg, command}

	return exec.Command(ShellPath, args...), nil

}
