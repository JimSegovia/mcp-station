//go:build !windows

package opencode

import "os/exec"

func setProcessGroup(cmd *exec.Cmd) {}
