// Copyright (c) 2019 Dropbox, Inc.
// Full license can be found in the LICENSE file.

package itest

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/rakyll/goebpf"
	"github.com/stretchr/testify/require"
)

// Runs XDP program and listen for perf events generated by program
func TestPerfEvents(t *testing.T) {
	// Read ELF, find map, load/attach program
	eb := goebpf.NewDefaultEbpfSystem()
	err := eb.LoadElf(xdpProgramFilename)
	require.NoError(t, err)
	perfMap := eb.GetMapByName("perf_map")
	require.NotNil(t, perfMap)

	program := eb.GetProgramByName("xdp_perf")
	require.NotNil(t, program)

	// Load / Attach program to "lo" interface
	err = program.Load()
	require.NoError(t, err)
	err = program.Attach("lo")
	require.NoError(t, err)
	defer program.Detach()

	// Setup/Start perf events
	perfEvents, err := goebpf.NewPerfEvents(perfMap)
	require.NoError(t, err)
	perfCh, err := perfEvents.StartForAllProcessesAndCPUs(4096)
	require.NoError(t, err)

	// Send dummy UDP packet to localhost so XDP program
	// will catch it and emit perf event
	conn, err := net.Dial("udp", "127.0.0.1:4444")
	require.NoError(t, err)
	conn.Write([]byte("test"))

	// Read Perf Events in a separated goroutine
	var firstEventData []byte
	doneChan := make(chan struct{})
	go func() {
		for {
			data, ok := <-perfCh
			if doneChan != nil {
				firstEventData = data
				close(doneChan)
				doneChan = nil
			}
			if !ok {
				break
			}
		}
	}()

	// Wait until first perf event (up to 1 sec)
	select {
	case <-doneChan:
		break
	case <-time.After(1 * time.Second):
		require.Fail(t, "timeout while waiting for perf event")
	}
	perfEvents.Stop()

	// Verify first received event
	packetSize := binary.LittleEndian.Uint32(firstEventData)
	require.True(t, packetSize > 20)
}
