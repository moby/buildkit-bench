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
		benchmarkBuildLocalSecret,
		benchmarkBuildRemoteBuildme,
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
	for i := 0; i < b.N; i++ {
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
}

// https://github.com/docker/buildx/issues/2479
func benchmarkBuildLocalSecret(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM python:latest
RUN --mount=type=secret,id=SECRET cat /run/secrets/SECRET
`)
	for i := 0; i < b.N; i++ {
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
}

func benchmarkBuildRemoteBuildme(b *testing.B, sb testutil.Sandbox) {
	for i := 0; i < b.N; i++ {
		start := time.Now()
		out, err := buildCmd(sb, withArgs(
			"--opt=context=https://github.com/dvdksn/buildme.git#eb6279e0ad8a10003718656c6867539bd9426ad8",
			"--opt=build-arg:BUILDKIT_SYNTAX=docker/dockerfile:1.9.0", // pin dockerfile syntax
		))
		testutil.ReportMetricDuration(b, time.Since(start))
		require.NoError(b, err, out)
	}
}
