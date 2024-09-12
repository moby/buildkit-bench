package test

import (
	"sync"
	"testing"
	"time"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func BenchmarkBuild(b *testing.B) {
	testutil.Run(b, testutil.BenchFuncs(
		benchmarkBuildLocal,
		benchmarkBuildLocalSecret,
		benchmarkBuildRemoteBuildme,
		benchmarkBuildBreaker16,
		benchmarkBuildBreaker32,
		benchmarkBuildBreaker64,
		benchmarkBuildBreaker128,
	), testutil.WithMirroredImages(testutil.OfficialImages(
		"busybox:latest",
		"golang:1.22-alpine",
		"python:latest",
	)))
}

func benchmarkBuildLocal(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM busybox:latest AS base
COPY foo /etc/foo
RUN cp /etc/foo /etc/bar

FROM scratch
COPY --from=base /etc/bar /bar
`)
	dir := tmpdir(
		b,
		fstest.CreateFile("Dockerfile", dockerfile, 0600),
		fstest.CreateFile("foo", []byte("foo"), 0600),
	)
	start := time.Now()
	out, err := buildCmd(sb, withDir(dir), withArgs(
		"--local=context=.",
		"--local=dockerfile=.",
	))
	testutil.ReportMetricDuration(b, time.Since(start))
	require.NoError(b, err, out)
}

// https://github.com/docker/buildx/issues/2479
func benchmarkBuildLocalSecret(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM python:latest
RUN --mount=type=secret,id=SECRET cat /run/secrets/SECRET
`)
	dir := tmpdir(
		b,
		fstest.CreateFile("Dockerfile", dockerfile, 0600),
		fstest.CreateFile("secret.txt", []byte("mysecret"), 0600),
	)
	start := time.Now()
	out, err := buildCmd(sb, withDir(dir), withArgs(
		"--local=context=.",
		"--local=dockerfile=.",
		"--secret=id=SECRET,src=secret.txt",
	))
	testutil.ReportMetricDuration(b, time.Since(start))
	require.NoError(b, err, out)
}

func benchmarkBuildRemoteBuildme(b *testing.B, sb testutil.Sandbox) {
	start := time.Now()
	out, err := buildCmd(sb, withArgs(
		"--opt=context=https://github.com/dvdksn/buildme.git#eb6279e0ad8a10003718656c6867539bd9426ad8",
		"--opt=build-arg:BUILDKIT_SYNTAX=docker/dockerfile:1.9.0", // pin dockerfile syntax
	))
	testutil.ReportMetricDuration(b, time.Since(start))
	require.NoError(b, err, out)
}

func benchmarkBuildBreaker16(b *testing.B, sb testutil.Sandbox) {
	buildBreaker(b, sb, 16)
}

func benchmarkBuildBreaker32(b *testing.B, sb testutil.Sandbox) {
	buildBreaker(b, sb, 32)
}

func benchmarkBuildBreaker64(b *testing.B, sb testutil.Sandbox) {
	buildBreaker(b, sb, 64)
}

func benchmarkBuildBreaker128(b *testing.B, sb testutil.Sandbox) {
	buildBreaker(b, sb, 128)
}

func buildBreaker(b *testing.B, sb testutil.Sandbox, n int) {
	type buildRecord struct {
		id  int
		d   time.Duration
		out string
		err error
	}

	var wg sync.WaitGroup
	records := make(chan buildRecord, 5)

	for i := 0; i < n; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			start := time.Now()
			out, err := buildCmd(sb, withArgs(
				"--opt=context=https://github.com/dvdksn/buildme.git#eb6279e0ad8a10003718656c6867539bd9426ad8",
				"--opt=build-arg:BUILDKIT_SYNTAX=docker/dockerfile:1.9.0", // pin dockerfile syntax
			))
			d := time.Since(start)
			b.Logf("build %d: %fs", i, d.Seconds())
			record := buildRecord{
				id:  i,
				d:   d,
				out: out,
				err: err,
			}
			records <- record
			require.NoError(b, err, out)
		}()
	}

	go func() {
		wg.Wait()
		close(records)
	}()

	var avgd time.Duration
	var totald time.Duration
	for record := range records {
		if record.err != nil {
			b.Fatalf("build %d failed: %v", record.id, record.err)
		}
		totald += record.d
	}
	avgd = totald / time.Duration(n)
	testutil.ReportMetricDuration(b, avgd)
}
