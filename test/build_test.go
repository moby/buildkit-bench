package test

import (
	"testing"
	"time"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func BenchmarkBuild(b *testing.B) {
	testutil.Run(b, testutil.BenchFuncs(
		benchmarkBuildLocal,
	), testutil.WithMirroredImages(testutil.OfficialImages("busybox:latest")))
}

func benchmarkBuildLocal(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM busybox:latest AS base
COPY foo /etc/foo
RUN cp /etc/foo /etc/bar

FROM scratch
COPY --from=base /etc/bar /bar
`)
	for i := 0; i < b.N; i++ {
		dir := tmpdir(
			b,
			fstest.CreateFile("Dockerfile", dockerfile, 0600),
			fstest.CreateFile("foo", []byte("foo"), 0600),
		)
		start := time.Now()
		out, err := buildCmd(sb, withArgs("--no-cache", "--local=context="+dir, "--local=dockerfile="+dir))
		testutil.ReportMetricDuration(b, time.Since(start))
		require.NoError(b, err, out)
	}
}
