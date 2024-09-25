package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/containerd/platforms"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

const dockerfileImagePin = "docker/dockerfile:1.9.0"

var contextDirApplier fstest.Applier

func BenchmarkBuild(b *testing.B) {
	mirroredImages := testutil.OfficialImages(
		"busybox:latest",
		"golang:1.22-alpine",
		"python:latest",
	)
	mirroredImages[dockerfileImagePin] = "docker.io/" + dockerfileImagePin
	mirroredImages["amd64/busybox:latest"] = "docker.io/amd64/busybox:latest"
	mirroredImages["arm64v8/busybox:latest"] = "docker.io/arm64v8/busybox:latest"

	var contextDirAppliers []fstest.Applier
	contextDirAppliers = append(contextDirAppliers,
		fstest.CreateDir("subdir1", 0755),
		fstest.CreateFile("subdir1/file1.txt", []byte("foo"), 0600),
		fstest.CreateFile("subdir1/file2.txt", make([]byte, 1024*1024), 0600), // 1MB file
		fstest.CreateDir("subdir1/subdir2", 0755),
		fstest.CreateFile("subdir1/subdir2/file3.txt", []byte("bar"), 0600),
		fstest.CreateFile("subdir1/subdir2/file4.txt", make([]byte, 1024*1024*10), 0600), // 10MB file
	)
	for i := 0; i < 5000; i++ {
		contextDirAppliers = append(contextDirAppliers, fstest.CreateFile(fmt.Sprintf("subdir1/file%d.txt", i+5), []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."), 0600))
	}
	contextDirAppliers = append(contextDirAppliers,
		fstest.CreateFile("subdir1/largefile1.txt", make([]byte, 1024*1024*50), 0600),  // 50MB file
		fstest.CreateFile("subdir1/largefile2.txt", make([]byte, 1024*1024*100), 0600), // 100MB file
	)
	contextDirApplier = fstest.Apply(contextDirAppliers...)

	testutil.Run(b, testutil.BenchFuncs(
		benchmarkBuildSimple,
		benchmarkBuildMultistage,
		benchmarkBuildSecret,
		benchmarkBuildRemote,
		benchmarkBuildHighParallelization16x,
		benchmarkBuildHighParallelization32x,
		benchmarkBuildHighParallelization64x,
		benchmarkBuildHighParallelization128x,
		benchmarkBuildFileTransfer,
		benchmarkBuildEmulator,
		benchmarkBuildExportUncompressed,
		benchmarkBuildExportGzip,
		benchmarkBuildExportEstargz,
		//benchmarkBuildExportZstd, https://github.com/moby/buildkit-bench/pull/146#discussion_r1771519112
	), testutil.WithMirroredImages(mirroredImages))
}

func benchmarkBuildSimple(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM busybox:latest
COPY foo /etc/foo
RUN cp /etc/foo /etc/bar
`)
	dir := tmpdir(
		b,
		fstest.CreateFile("Dockerfile", dockerfile, 0600),
		fstest.CreateFile("foo", []byte("foo"), 0600),
	)
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(dir))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}

func benchmarkBuildMultistage(b *testing.B, sb testutil.Sandbox) {
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
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(dir))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}

// https://github.com/docker/buildx/issues/2479
func benchmarkBuildSecret(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM python:latest
RUN --mount=type=secret,id=SECRET cat /run/secrets/SECRET
`)
	dir := tmpdir(
		b,
		fstest.CreateFile("Dockerfile", dockerfile, 0600),
		fstest.CreateFile("secret.txt", []byte("mysecret"), 0600),
	)
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withDir(dir), withArgs("--secret=id=SECRET,src=secret.txt", "."))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}

func benchmarkBuildRemote(b *testing.B, sb testutil.Sandbox) {
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(
			"--build-arg=BUILDKIT_SYNTAX="+dockerfileImagePin,
			"https://github.com/dvdksn/buildme.git#eb6279e0ad8a10003718656c6867539bd9426ad8",
		))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}

func benchmarkBuildHighParallelization16x(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildHighParallelization(b, sb, 16)
}
func benchmarkBuildHighParallelization32x(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildHighParallelization(b, sb, 32)
}
func benchmarkBuildHighParallelization64x(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildHighParallelization(b, sb, 64)
}
func benchmarkBuildHighParallelization128x(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildHighParallelization(b, sb, 128)
}
func benchmarkBuildHighParallelization(b *testing.B, sb testutil.Sandbox, n int) {
	dockerfile := []byte(`
FROM busybox:latest AS base
COPY foo /etc/foo
RUN cp /etc/foo /etc/bar
`)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dir := tmpdir(
				b,
				fstest.CreateFile("Dockerfile", dockerfile, 0600),
				fstest.CreateFile("foo", []byte("foo"), 0600),
			)
			out, err := buildxBuildCmd(sb, withArgs("--output=type=image", dir))
			// TODO: use sb.WriteLogFile to write buildx logs in a defer with a
			//  semaphore using a buffered channel to limit the number of
			//  concurrent goroutines. This might affect timing.
			require.NoError(b, err, out)
		}()
	}
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		wg.Wait()
		b.StopTimer()
	})
}

func benchmarkBuildFileTransfer(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM busybox:latest
WORKDIR /src
COPY . .
RUN du -sh . && tree .
`)
	dir := tmpdir(b,
		fstest.CreateFile("Dockerfile", dockerfile, 0600),
		contextDirApplier,
	)
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(dir))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}

// https://github.com/moby/buildkit/pull/4949
func benchmarkBuildEmulator(b *testing.B, sb testutil.Sandbox) {
	var busyboxImage string
	var platform platforms.Platform
	defaultPlatform := platforms.Normalize(platforms.DefaultSpec())
	if strings.HasPrefix(defaultPlatform.Architecture, "arm") {
		busyboxImage = "amd64/busybox:latest"
		platform = platforms.Normalize(platforms.Platform{
			OS:           defaultPlatform.OS,
			Architecture: "amd64",
		})
	} else {
		busyboxImage = "arm64v8/busybox:latest"
		platform = platforms.Normalize(platforms.Platform{
			OS:           defaultPlatform.OS,
			Architecture: "arm64",
		})
	}

	dockerfile := []byte(fmt.Sprintf(`
FROM %s
ENV QEMU_STRACE=1
RUN uname -a
`, busyboxImage))

	dir := tmpdir(b, fstest.CreateFile("Dockerfile", dockerfile, 0600))
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(
			"--build-arg", "BUILDKIT_DOCKERFILE_CHECK=skip=all", // skip all checks (for InvalidBaseImagePlatform): https://docs.docker.com/build/checks/#skip-checks
			"--platform", platforms.Format(platform),
			dir,
		))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}

func benchmarkBuildExportUncompressed(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildExport(b, sb, "uncompressed")
}
func benchmarkBuildExportGzip(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildExport(b, sb, "gzip")
}
func benchmarkBuildExportEstargz(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildExport(b, sb, "gzip")
}
func benchmarkBuildExportZstd(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildExport(b, sb, "zstd")
}
func benchmarkBuildExport(b *testing.B, sb testutil.Sandbox, compression string) {
	dockerfile := []byte(`
FROM python:latest
WORKDIR /src
COPY . .
`)
	dir := tmpdir(b,
		fstest.CreateFile("Dockerfile", dockerfile, 0600),
		contextDirApplier,
	)
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs("--output=type=image,compression="+compression, dir))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}
