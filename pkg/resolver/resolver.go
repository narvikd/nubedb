package resolver

import (
	"net"
	"time"
)

const timeoutTimer = 300 * time.Millisecond

func IsHostResolvable(host string) bool {
	timeout := time.After(timeoutTimer)
	result := make(chan error)
	go func() {
		_, err := net.LookupHost(host)
		result <- err
	}()
	select {
	case <-timeout:
		return false
	case err := <-result:
		return err == nil
	}
}
