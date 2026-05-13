package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/google/pprof/profile"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func tmpdir(tb testing.TB, appliers ...fstest.Applier) string {
	tb.Helper()
	tmpdir := tb.TempDir()
	err := fstest.Apply(appliers...).Apply(tmpdir)
	require.NoError(tb, err)
	return tmpdir
}

type cmdOpt func(*exec.Cmd)

func withArgs(args ...string) cmdOpt {
	return func(cmd *exec.Cmd) {
		cmd.Args = append(cmd.Args, args...)
	}
}

func withDir(dir string) cmdOpt {
	return func(cmd *exec.Cmd) {
		cmd.Dir = dir
	}
}

func buildxCmd(sb testutil.Sandbox, opts ...cmdOpt) *exec.Cmd {
	cmd := exec.CommandContext(context.Background(), sb.BuildxBin()) //nolint:gosec // test utility
	cmd.Env = append([]string{}, os.Environ()...)
	for _, opt := range opts {
		opt(cmd)
	}
	if buildxConfigDir := sb.BuildxConfigDir(); buildxConfigDir != "" {
		cmd.Env = append(cmd.Env, "BUILDX_CONFIG="+buildxConfigDir)
	}
	if builderName := sb.BuilderName(); builderName != "" {
		cmd.Env = append(cmd.Env, "BUILDX_BUILDER="+builderName)
	}
	return cmd
}

func buildxBuildCmd(sb testutil.Sandbox, opts ...cmdOpt) (string, error) {
	opts = append([]cmdOpt{withArgs("build")}, opts...)
	cmd := buildxCmd(sb, opts...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func reportBuildkitdAlloc(b *testing.B, sb testutil.Sandbox, cb func()) {
	beforeAlloc, errb := buildkitdAlloc(sb)
	cb()
	afterAlloc, erra := buildkitdAlloc(sb)
	testutil.ReportAlloc(b, afterAlloc-beforeAlloc)
	require.NoError(b, errb)
	require.NoError(b, erra)
}

func buildkitdAlloc(sb testutil.Sandbox) (int64, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("http://%s/debug/pprof/heap?gc=1", sb.DebugAddress()), nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	p, err := profile.Parse(resp.Body)
	if err != nil {
		return 0, err
	}

	return func(prof *profile.Profile, value func(v []int64) int64) int64 {
		var total, diff int64
		for _, sample := range prof.Sample {
			var v int64
			v = value(sample.Value)
			if v < 0 {
				v = -v
			}
			total += v
			if sample.DiffBaseSample() {
				diff += v
			}
		}
		if diff > 0 {
			total = diff
		}
		return total
	}(p, func(v []int64) int64 {
		return v[1] // sample index of alloc_space in heap profiles
	}), nil
}
