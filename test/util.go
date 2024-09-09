package test

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/containerd/continuity/fs/fstest"
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

func buildCmd(sb testutil.Sandbox, opts ...cmdOpt) (string, error) {
	opts = append([]cmdOpt{withArgs("build", "--frontend=dockerfile.v0")}, opts...)
	cmd := buildctlCmd(sb, opts...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
