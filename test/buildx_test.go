package test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestBuildx(t *testing.T) {
	testutil.Run(t, testutil.TestFuncs(
		testBuildxVersion,
	))
}

func BenchmarkBuildx(b *testing.B) {
	testutil.Run(b, testutil.BenchFuncs(
		benchmarkBuildxVersion,
		benchmarkBuildxSize,
		benchmarkBuildxBakeMatrix,
	))
}

func benchmarkBuildxBakeMatrix(b *testing.B, sb testutil.Sandbox) {
	const bakefile = `
target "base" {
  context = "."
  dockerfile = "Dockerfile"
  labels = { "bench" = "matrix" }
}
target "default" {
  inherits = ["base"]
  matrix = {
    app = range(32)
    mode = ["debug", "release"]
    arch = ["amd64", "arm64"]
  }
  name = "app-${app}-${mode}-${arch}"
  args = { APP = "${app}", MODE = mode }
  platforms = ["linux/${arch}"]
}
`
	dir := tmpdir(b, fstest.CreateFile("docker-bake.hcl", []byte(bakefile), 0600))
	cmd := buildxCmd(sb, withDir(dir), withArgs("bake", "--print", "-f", "docker-bake.hcl"))
	b.StartTimer()
	out, err := cmd.Output()
	b.StopTimer()
	sb.WriteLogFile(b, "bake", out)
	require.NoError(b, err)
	var result struct {
		Target map[string]struct {
			Args      map[string]string `json:"args"`
			Labels    map[string]string `json:"labels"`
			Platforms []string          `json:"platforms"`
		} `json:"target"`
	}
	require.NoError(b, json.Unmarshal(out, &result))
	require.Len(b, result.Target, 128)
	for app := 0; app < 32; app++ {
		for _, mode := range []string{"debug", "release"} {
			for _, arch := range []string{"amd64", "arm64"} {
				target, ok := result.Target[fmt.Sprintf("app-%d-%s-%s", app, mode, arch)]
				require.True(b, ok)
				require.Equal(b, fmt.Sprint(app), target.Args["APP"])
				require.Equal(b, mode, target.Args["MODE"])
				require.Equal(b, "matrix", target.Labels["bench"])
				require.Equal(b, []string{"linux/" + arch}, target.Platforms)
			}
		}
	}
}

func testBuildxVersion(t *testing.T, sb testutil.Sandbox) {
	output, err := exec.Command(sb.BuildxBin(), "version").Output()
	require.NoError(t, err)
	t.Log(string(output))
}

func benchmarkBuildxVersion(b *testing.B, sb testutil.Sandbox) {
	b.StartTimer()
	err := exec.Command(sb.BuildxBin(), "version").Run()
	b.StopTimer()
	require.NoError(b, err)
}

func benchmarkBuildxSize(b *testing.B, sb testutil.Sandbox) {
	fi, err := os.Stat(sb.BuildxBin())
	require.NoError(b, err)
	testutil.ReportMetric(b, float64(fi.Size()), testutil.MetricBytes)
}
