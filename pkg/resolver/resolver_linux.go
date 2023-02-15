//go:build linux

package resolver

import (
	pinger "github.com/prometheus-community/pro-bing"
	"time"
)

func IsHostAlive(host string) bool {
	if !IsHostResolvable(host) {
		return false
	}
	p, err := pinger.NewPinger(host)
	if err != nil {
		return false
	}
	p.Count = 1
	p.Interval = 300 * time.Millisecond
	p.Timeout = 300 * time.Millisecond

	if p.Run() != nil {
		return false
	}
	return p.Statistics().PacketsRecv >= 1
}
