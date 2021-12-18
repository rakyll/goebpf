//go:build !linux
// +build !linux

package goebpf

import "time"

// KtimeToTime converts kernel time (nanoseconds since boot) to time.Time
func KtimeToTime(ktime uint64) time.Time {
	panic("not implemented")
}
