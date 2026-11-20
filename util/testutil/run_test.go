package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pkg/errors"
)

type failingWorker string

func (w failingWorker) Name() string { return string(w) }
func (w failingWorker) New(context.Context, *BackendConfig) (Backend, func() error, error) {
	fmt.Println("attempted worker " + w.Name())
	return nil, nil, errors.Errorf("worker %s failed", w)
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
			cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestRunContinuesAfterFailure$", "-test.v") //nolint:gosec // Re-execute this test binary to isolate expected failures.
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

type sandboxLifetimeWorker struct {
	active, peak, created, closed int
}

func (w *sandboxLifetimeWorker) Name() string { return "lifetime" }
func (w *sandboxLifetimeWorker) New(context.Context, *BackendConfig) (Backend, func() error, error) {
	w.active++
	w.created++
	w.peak = max(w.peak, w.active)
	return backend{}, func() error {
		w.active--
		w.closed++
		return nil
	}, nil
}

func TestRunBenchmarkSandboxCleanup(t *testing.T) {
	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^$", "-test.bench=^BenchmarkRunSandboxCleanup$", "-test.benchtime=3x") //nolint:gosec // Re-execute this test binary to run the benchmark with a fixed iteration count.
	cmd.Env = append(os.Environ(), "BENCH_SANDBOX_CLEANUP_HELPER=1", "TEST_BENCH_RUN=1", "BUILDKIT_REF_RANDOM=")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	require.Contains(t, string(output), "BenchmarkSandboxLifetime/ref=lifetime/buildx=probe/run=1")
	require.Contains(t, string(output), "BenchmarkSandboxLifetimeWithIteration/ref=lifetime/buildx=probe/run=1")
}

func BenchmarkRunSandboxCleanup(b *testing.B) {
	if os.Getenv("BENCH_SANDBOX_CLEANUP_HELPER") != "1" {
		b.Skip("subprocess helper for TestRunBenchmarkSandboxCleanup")
	}
	workers, output, bins := defaultWorkers, outDir, buildxBinsDir
	b.Cleanup(func() { defaultWorkers, outDir, buildxBinsDir = workers, output, bins })
	outDir, buildxBinsDir = b.TempDir(), b.TempDir()
	for _, tc := range []struct {
		name string
		fn   func(*testing.B, Sandbox)
	}{
		{
			name: "ordinary",
			fn:   benchmarkSandboxLifetime,
		},
		{
			name: "WithIteration",
			fn:   benchmarkSandboxLifetimeWithIteration,
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			w := &sandboxLifetimeWorker{}
			defaultWorkers = []Worker{w}
			Run(b, BenchFuncs(tc.fn), WithMatrix("buildx", map[string]any{"probe": "unused"}))
			require.Equal(b, 1, w.peak)
			require.Equal(b, w.created, w.closed)
			require.Zero(b, w.active)
		})
	}
}

func benchmarkSandboxLifetime(b *testing.B, _ Sandbox) {
	b.StartTimer()
	b.StopTimer()
}

func benchmarkSandboxLifetimeWithIteration(b *testing.B, _ Sandbox) {
	for range b.N {
		b.StartTimer()
		b.StopTimer()
	}
}

func TestRunSandboxSurvivesTestCleanup(t *testing.T) {
	workers, output, bins := defaultWorkers, outDir, buildxBinsDir
	t.Cleanup(func() { defaultWorkers, outDir, buildxBinsDir = workers, output, bins })
	outDir, buildxBinsDir = t.TempDir(), t.TempDir()
	w := &sandboxLifetimeWorker{}
	defaultWorkers = []Worker{w}
	Run(t, TestFuncs(func(t *testing.T, _ Sandbox) {
		t.Cleanup(func() { require.Equal(t, 1, w.active) })
	}), WithMatrix("buildx", map[string]any{"probe": "unused"}))
	require.Equal(t, 1, w.created)
	require.Equal(t, w.created, w.closed)
	require.Zero(t, w.active)
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
