package test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
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

func withEnv(env ...string) cmdOpt {
	return func(cmd *exec.Cmd) {
		cmd.Env = append(cmd.Env, env...)
	}
}

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
	cmd := exec.Command("buildx")
	cmd.Env = append([]string{}, os.Environ()...)
	for _, opt := range opts {
		opt(cmd)
	}
	if buildxDir := sb.BuildxDir(); buildxDir != "" {
		cmd.Env = append(cmd.Env, "BUILDX_CONFIG="+buildxDir)
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

func buildctlCmd(sb testutil.Sandbox, opts ...cmdOpt) *exec.Cmd {
	cmd := exec.Command(path.Join(sb.BinsDir(), sb.Name(), "buildctl"))
	cmd.Args = append(cmd.Args, "--debug")
	if buildkitAddr := sb.Address(); buildkitAddr != "" {
		cmd.Args = append(cmd.Args, "--addr", buildkitAddr)
	}
	cmd.Env = append([]string{}, os.Environ()...)
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

func buildctlBuildCmd(sb testutil.Sandbox, opts ...cmdOpt) (string, error) {
	opts = append([]cmdOpt{withArgs("build", "--frontend=dockerfile.v0")}, opts...)
	cmd := buildctlCmd(sb, opts...)
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
	resp, err := client.Get(fmt.Sprintf("http://%s/debug/pprof/heap?gc=1", sb.DebugAddress()))
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
