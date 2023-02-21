package resolver

import (
	"bytes"
	"net"
	"os/exec"
	"strings"
	"time"
)

// IsHostAlive checks if a given host is alive by first checking if the host is resolvable.
//
// Then it executes a ping command with a timeout.
func IsHostAlive(host string, timeout time.Duration) bool {
	var out bytes.Buffer
	if !isHostResolvable(host, timeout) {
		return false
	}
	cmd := exec.Command("timeout", "0.3", "ping", "-c", "1", host)
	cmd.Stdout = &out

	// Check if the ping command was successful or if it returned a response
	isAlive := cmd.Run() == nil || strings.Contains(out.String(), "bytes from")
	return isAlive
}

// isHostResolvable checks if a given host can be resolved to an IP address within a given timeout.
func isHostResolvable(host string, timeout time.Duration) bool {
	// Start a timer with the given timeout
	t := time.After(timeout)
	// Create a channel to receive the result of the lookup operation
	result := make(chan error)

	// Execute the lookup operation in a separate goroutine
	go func() {
		_, err := net.LookupHost(host)
		result <- err
	}()

	// Wait for the result or for the timeout to expire
	select {
	case <-t:
		return false
	case err := <-result:
		return err == nil
	}
}
