package testutil

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWaitSocketProcessExit(t *testing.T) {
	if os.Getenv("BENCH_SOCKET_EXIT_HELPER") == "1" {
		_, _ = io.Copy(io.Discard, os.Stdin)
		return
	}

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWaitSocketProcessExit$") //nolint:gosec // Re-execute this test binary as the child process.
	cmd.Env = append(os.Environ(), "BENCH_SOCKET_EXIT_HELPER=1")
	cmd.Stderr = io.Discard
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stop, stopped, err := startCmd(cmd, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	// Let the child exit while waitSocket is polling for a socket it never opens.
	require.NoError(t, stdin.Close())
	err = waitSocket(filepath.Join(t.TempDir(), "missing.sock"), 10*time.Second, cmd, stopped)
	require.EqualError(t, err, "process exited: "+cmd.String())
	select {
	case <-stopped:
	default:
		t.Fatal("process exit was not signaled")
	}
}
