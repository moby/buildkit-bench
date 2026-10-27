package test

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/containerd/platforms"
	"github.com/moby/buildkit-bench/util/archiveutil"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

const dockerfileImagePin = "docker/dockerfile:1.9.0"

var contextDirApplier fstest.Applier

func BenchmarkBuild(b *testing.B) {
	mirroredImages := testutil.OfficialImages(
		"library/busybox:latest", // Include the emulated platform using the multi-platform pin.
		"golang:1.22-alpine",
		"python:latest",
	)
	mirroredImages[dockerfileImagePin] = "docker.io/" + dockerfileImagePin

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
		benchmarkBuildFileTransferCachedWithIteration,
		benchmarkBuildFileTransferIncremental,
		benchmarkBuildLocalCacheRestore,
		benchmarkBuildBakeSharedTargets,
		benchmarkBuildAddTarManyFiles,
		benchmarkBuildAddTarLargeFile,
		benchmarkBuildEmulator,
		benchmarkBuildEmulatorNative,
		benchmarkBuildExportUncompressed,
		benchmarkBuildExportGzip,
		benchmarkBuildExportEstargz,
		//benchmarkBuildExportZstd, https://github.com/moby/buildkit-bench/pull/146#discussion_r1771519112
	),
		testutil.WithMirroredImages(mirroredImages),
	)
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

	// warmup
	for i := 0; i < 5; i++ {
		dir := tmpdir(
			b,
			fstest.CreateFile("Dockerfile", dockerfile, 0600),
			fstest.CreateFile("foo", []byte("foo"), 0600),
		)
		out, err := buildxBuildCmd(sb, withArgs("--output=type=image", dir))
		require.NoError(b, err, out)
	}

	dirs := make([]string, n)
	for i := 0; i < n; i++ {
		dirs[i] = tmpdir(
			b,
			fstest.CreateFile("Dockerfile", dockerfile, 0600),
			fstest.CreateFile("foo", []byte("foo"), 0600),
		)
	}
	outputs := make([]string, n)
	reportBuildkitdAlloc(b, sb, func() {
		var builds errgroup.Group
		b.StartTimer()
		for i, dir := range dirs {
			builds.Go(func() error {
				out, err := buildxBuildCmd(sb, withArgs("--output=type=image", dir))
				outputs[i] = out
				return err
			})
		}
		err := builds.Wait()
		b.StopTimer()
		for i, out := range outputs {
			sb.WriteLogFile(b, fmt.Sprintf("buildx-%d", i), []byte(out))
		}
		require.NoError(b, err)
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

func benchmarkBuildFileTransferCachedWithIteration(b *testing.B, sb testutil.Sandbox) {
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

	// first build to warmup cache
	out, err := buildxBuildCmd(sb, withArgs(dir))
	require.NoError(b, err, out)

	for i := 0; i < b.N; i++ {
		reportBuildkitdAlloc(b, sb, func() {
			b.StartTimer()
			out, err := buildxBuildCmd(sb, withArgs(dir))
			b.StopTimer()
			sb.WriteLogFile(b, "buildx", []byte(out))
			require.NoError(b, err, out)
		})
	}
}

func benchmarkBuildFileTransferIncremental(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM busybox:latest AS build
COPY subdir1 /src
RUN sha256sum /src/file1.txt > /checksum
FROM scratch
COPY --from=build /checksum /checksum
`)
	dir := tmpdir(b,
		fstest.CreateFile("Dockerfile", dockerfile, 0600),
		contextDirApplier,
	)
	out, err := buildxBuildCmd(sb, withArgs(dir))
	sb.WriteLogFile(b, "warmup", []byte(out))
	require.NoError(b, err, out)

	// Change the size as well as contents so metadata-based differs see the edit.
	content := []byte("updated source file\n")
	require.NoError(b, os.WriteFile(filepath.Join(dir, "subdir1", "file1.txt"), content, 0600))
	dest := b.TempDir()
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs("--output=type=local,dest="+dest, dir))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
	checksum, err := os.ReadFile(filepath.Join(dest, "checksum"))
	require.NoError(b, err)
	require.Equal(b, fmt.Sprintf("%x  /src/file1.txt\n", sha256.Sum256(content)), string(checksum))
}

func benchmarkBuildLocalCacheRestore(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(`
FROM busybox:latest AS build
RUN dd if=/dev/zero of=/payload bs=1048576 count=16 && \
    sha256sum /payload > /checksum && \
    cat /proc/sys/kernel/random/uuid > /cache-marker
FROM scratch
COPY --from=build /payload /payload
COPY --from=build /checksum /checksum
COPY --from=build /cache-marker /cache-marker
`)
	// Generate the payload during cache export, so import cannot be dominated
	// by retransferring a large local context after pruning the source cache.
	dir := tmpdir(b, fstest.CreateFile("Dockerfile", dockerfile, 0600))
	cache, before, after := b.TempDir(), b.TempDir(), b.TempDir()
	out, err := buildxBuildCmd(sb, withArgs(
		"--cache-to=type=local,mode=max,dest="+cache,
		"--output=type=local,dest="+before, dir,
	))
	sb.WriteLogFile(b, "cache-export", []byte(out))
	require.NoError(b, err, out)

	// Solves release references asynchronously. Like upstream ensurePruneAll,
	// retry pruning and require an empty daemon cache before measuring import.
	var records []byte
	for attempt := 0; attempt < 20; attempt++ {
		out, err := buildctlCmd(sb, withArgs("prune", "--all")).CombinedOutput()
		require.NoError(b, err, string(out))
		records, err = buildctlCmd(sb, withArgs("du", "--format={{len .}}")).Output()
		require.NoError(b, err)
		if strings.TrimSpace(string(records)) == "0" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.Equal(b, "0", strings.TrimSpace(string(records)), "local cache must be empty before external cache import")
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(
			"--cache-from=type=local,src="+cache,
			"--output=type=local,dest="+after, dir,
		))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
	for _, name := range []string{"checksum", "cache-marker"} {
		expected, err := os.ReadFile(filepath.Join(before, name))
		require.NoError(b, err)
		require.NotEmpty(b, expected)
		actual, err := os.ReadFile(filepath.Join(after, name))
		require.NoError(b, err)
		require.Equal(b, expected, actual, name)
	}
	payload, err := os.ReadFile(filepath.Join(after, "payload"))
	require.NoError(b, err)
	require.Len(b, payload, 16*1024*1024)
	checksum, err := os.ReadFile(filepath.Join(after, "checksum"))
	require.NoError(b, err)
	require.Equal(b, fmt.Sprintf("%x  /payload\n", sha256.Sum256(payload)), string(checksum))
}

func benchmarkBuildBakeSharedTargets(b *testing.B, sb testutil.Sandbox) {
	const targets = 16
	var bake strings.Builder
	bake.WriteString("group \"default\" {\n  targets = [")
	for i := 0; i < targets; i++ {
		fmt.Fprintf(&bake, "\"leaf%d\",", i)
	}
	bake.WriteString("]\n}\n")
	bake.WriteString(`
target "base" {
  output = ["type=cacheonly"]
  dockerfile-inline = "FROM scratch\nCOPY payload /payload\n"
}
`)
	payload := []byte(strings.Repeat("shared build context\n", 50000))
	files := []fstest.Applier{
		fstest.CreateFile("payload", payload, 0600),
		fstest.CreateFile(".dockerignore", []byte("out\n"), 0600),
	}
	for i := 0; i < targets; i++ {
		fmt.Fprintf(&bake, `
target "leaf%d" {
  contexts = { base = "target:base" }
  output = ["type=local,dest=out/leaf%d"]
  dockerfile-inline = "FROM base\nCOPY leaf%d /leaf\n"
}
`, i, i, i)
		files = append(files, fstest.CreateFile(fmt.Sprintf("leaf%d", i), []byte(fmt.Sprint(i)), 0600))
	}
	files = append(files, fstest.CreateFile("docker-bake.hcl", []byte(bake.String()), 0600))
	dir := tmpdir(b, files...)
	reportBuildkitdAlloc(b, sb, func() {
		cmd := buildxCmd(sb, withDir(dir), withArgs("bake", "-f", "docker-bake.hcl"))
		b.StartTimer()
		out, err := cmd.CombinedOutput()
		b.StopTimer()
		sb.WriteLogFile(b, "bake", out)
		require.NoError(b, err, string(out))
	})
	for i := 0; i < targets; i++ {
		dest := filepath.Join(dir, "out", fmt.Sprintf("leaf%d", i))
		actual, err := os.ReadFile(filepath.Join(dest, "payload"))
		require.NoError(b, err)
		require.Equal(b, payload, actual)
		leaf, err := os.ReadFile(filepath.Join(dest, "leaf"))
		require.NoError(b, err)
		require.Equal(b, fmt.Sprint(i), string(leaf))
	}
}

func benchmarkBuildAddTarManyFiles(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(fmt.Sprintf(`
FROM busybox:latest
ADD payload.tar /src/
RUN test -f /src/%s
`, archiveutil.ManyFilesLastFileName()))
	dir := tmpdir(b, fstest.CreateFile("Dockerfile", dockerfile, 0600))
	require.NoError(b, archiveutil.WriteManyFilesTar(filepath.Join(dir, "payload.tar")))

	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(dir))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}

func benchmarkBuildAddTarLargeFile(b *testing.B, sb testutil.Sandbox) {
	dockerfile := []byte(fmt.Sprintf(`
FROM busybox:latest
ADD payload.tar /src/
RUN test "$(stat -c %%s /src/payload/blob.bin)" = "%d"
`, archiveutil.LargeFileSizeBytes))
	dir := tmpdir(b, fstest.CreateFile("Dockerfile", dockerfile, 0600))
	require.NoError(b, archiveutil.WriteLargeFileTar(filepath.Join(dir, "payload.tar"), archiveutil.LargeFileSizeBytes))

	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(dir))
		b.StopTimer()
		sb.WriteLogFile(b, "buildx", []byte(out))
		require.NoError(b, err, out)
	})
}

func benchmarkBuildEmulator(b *testing.B, sb testutil.Sandbox) {
	platform := platforms.Normalize(platforms.DefaultSpec())
	switch platform.Architecture {
	case "amd64":
		platform.Architecture = "arm64"
	case "arm64":
		platform.Architecture = "amd64"
	default:
		b.Skip("emulator workload requires an amd64 or arm64 host")
	}
	platform.Variant = ""
	benchmarkBuildEmulatorWorkload(b, sb, platform)
}

func benchmarkBuildEmulatorNative(b *testing.B, sb testutil.Sandbox) {
	benchmarkBuildEmulatorWorkload(b, sb, platforms.Normalize(platforms.DefaultSpec()))
}

func benchmarkBuildEmulatorWorkload(b *testing.B, sb testutil.Sandbox, platform platforms.Platform) {
	// Keep CPU work, filesystem operations and child execution in separate RUNs.
	// Do not enable QEMU_STRACE: logging would distort the workload timings.
	dockerfile := []byte(`
FROM busybox:latest AS input
WORKDIR /work
COPY payload .

FROM input AS workload
RUN sha256sum payload > payload.sha256 && gzip -1 -c payload > payload.gz && \
    gzip -d -c payload.gz > restored && cmp payload restored
RUN mkdir chunks extracted && split -b 65536 payload chunks/part- && \
    tar -cf chunks.tar chunks && tar -xf chunks.tar -C extracted && \
    diff -r chunks extracted/chunks
RUN set -eu; for file in chunks/*; do sha256sum "$file"; done > chunks.sha256; \
    sha256sum -c chunks.sha256 && sha256sum -c payload.sha256
`)
	// Fixed pseudo-random bytes avoid measuring compression of an all-zero file.
	payload := make([]byte, 8*1024*1024)
	_, err := rand.New(rand.NewSource(0)).Read(payload)
	require.NoError(b, err)
	dir := tmpdir(b,
		fstest.CreateFile("Dockerfile", dockerfile, 0600),
		fstest.CreateFile("payload", payload, 0600),
	)
	// Populate only the input stage outside timing. The three RUNs remain cold.
	out, err := buildxBuildCmd(sb, withArgs("--platform", platforms.Format(platform), "--target=input", dir))
	sb.WriteLogFile(b, "warmup", []byte(out))
	require.NoError(b, err, out)
	reportBuildkitdAlloc(b, sb, func() {
		b.StartTimer()
		out, err := buildxBuildCmd(sb, withArgs(
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
	benchmarkBuildExport(b, sb, "estargz")
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
