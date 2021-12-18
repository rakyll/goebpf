//go:build linux
// +build linux

package goebpf

import (
	"syscall"
	"time"
)

// KtimeToTime converts kernel time (nanoseconds since boot) to time.Time
func KtimeToTime(ktime uint64) time.Time {
	si := &syscall.Sysinfo_t{}
	syscall.Sysinfo(si)
	boot := time.Now().Add(-time.Duration(si.Uptime) * time.Second)
	return boot.Add(time.Duration(ktime) * time.Nanosecond)
}
