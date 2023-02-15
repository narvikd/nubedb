package resolver

import (
	"net"
	"time"
)

func IsHostResolvable(host string) bool {
	timeout := time.After(200 * time.Millisecond)
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
