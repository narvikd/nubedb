//go:build darwin

package resolver

import (
	"bytes"
	"os/exec"
	"strings"
)

func IsHostAlive(host string) bool {
	var out bytes.Buffer
	if !IsHostResolvable(host) {
		return false
	}
	cmd := exec.Command("timeout", "0.3", "ping", "-c", "1", host)
	cmd.Stdout = &out
	isIt := cmd.Run() == nil || strings.Contains(out.String(), "bytes from")
	return isIt
}
