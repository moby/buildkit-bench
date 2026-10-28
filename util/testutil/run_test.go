package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type failingWorker string

func (w failingWorker) Name() string { return string(w) }
func (w failingWorker) New(context.Context, *BackendConfig) (Backend, func() error, error) {
	fmt.Println("attempted worker " + w.Name())
	return nil, nil, fmt.Errorf("worker %s failed", w)
}

func TestRunContinuesAfterFailure(t *testing.T) {
	if kind := os.Getenv("BENCH_FAILURE_HELPER"); kind != "" {
		defaultWorkers = []Worker{failingWorker("first"), failingWorker("second")}
		if kind == "benchmark" {
			failed := false
			testing.Benchmark(func(b *testing.B) {
				Run(b, BenchFuncs(func(b *testing.B, sb Sandbox) {}))
				failed = b.Failed()
			})
			require.True(t, failed)
		} else {
			Run(t, TestFuncs(func(t *testing.T, sb Sandbox) {}))
		}
		return
	}
	for _, kind := range []string{"test", "benchmark"} {
		t.Run(kind, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRunContinuesAfterFailure$", "-test.v")
			cmd.Env = append(os.Environ(), "BENCH_FAILURE_HELPER="+kind, "BUILDKIT_REF_RANDOM=", "TEST_BENCH_RUN=1")
			output, err := cmd.CombinedOutput()
			if kind == "test" {
				require.Error(t, err)
			} else {
				require.NoError(t, err, string(output))
			}
			require.Contains(t, string(output), "attempted worker first")
			require.Contains(t, string(output), "attempted worker second")
		})
	}
}

func TestOfficialImages(t *testing.T) {
	namespace := runtime.GOARCH
	if namespace == "arm64" {
		namespace = "arm64v8"
	} else if namespace != "amd64" {
		namespace = "library"
	}
	for _, tc := range []struct {
		name     string
		images   []string
		expected map[string]string
	}{
		{
			name:   "host architecture",
			images: []string{"busybox:latest"},
			expected: map[string]string{
				"library/busybox:latest": "docker.io/" + namespace + "/busybox:latest@" + pins["busybox:latest"][namespace],
			},
		},
		{
			name:   "multi-platform override does not affect subsequent images",
			images: []string{"library/busybox:latest", "alpine:latest"},
			expected: map[string]string{
				"library/busybox:latest": "docker.io/library/busybox:latest@" + pins["busybox:latest"]["library"],
				"library/alpine:latest":  "docker.io/" + namespace + "/alpine:latest@" + pins["alpine:latest"][namespace],
			},
		},
		{
			name:   "unpinned explicit namespace",
			images: []string{"library/busybox:unlisted"},
			expected: map[string]string{
				"library/busybox:unlisted": "docker.io/library/busybox:unlisted",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, OfficialImages(tc.images...))
		})
	}
}
