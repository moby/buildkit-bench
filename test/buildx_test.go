package test

import (
	"os"
	"os/exec"
	"testing"

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
	))
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
