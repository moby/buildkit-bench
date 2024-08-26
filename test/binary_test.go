package test

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestBinary(t *testing.T) {
	testutil.Run(t, testutil.TestFuncs(
		testBinaryVersion,
	))
}

func BenchmarkBinary(b *testing.B) {
	testutil.Run(b, testutil.BenchFuncs(
		benchmarkBinaryVersion,
		benchmarkBinarySize,
	))
}

func testBinaryVersion(t *testing.T, sb testutil.Sandbox) {
	buildkitdPath := path.Join(sb.BinsDir(), sb.Name(), "buildkitd")

	output, err := exec.Command(buildkitdPath, "--version").Output()
	require.NoError(t, err)

	versionParts := strings.Fields(string(output))
	require.Len(t, versionParts, 4)
	require.Equal(t, "buildkitd", versionParts[0])
	t.Log("repo:", versionParts[1])
	t.Log("version:", versionParts[2])
	t.Log("commit:", versionParts[3])
}

func benchmarkBinaryVersion(b *testing.B, sb testutil.Sandbox) {
	for i := 0; i < b.N; i++ {
		buildkitdPath := path.Join(sb.BinsDir(), sb.Name(), "buildkitd")
		require.NoError(b, exec.Command(buildkitdPath, "--version").Run())
	}
}

func benchmarkBinarySize(b *testing.B, sb testutil.Sandbox) {
	buildkitdPath := path.Join(sb.BinsDir(), sb.Name(), "buildkitd")
	fi, err := os.Stat(buildkitdPath)
	require.NoError(b, err)
	b.ResetTimer()
	testutil.ReportMetric(b, float64(fi.Size()), "bytes")
}
