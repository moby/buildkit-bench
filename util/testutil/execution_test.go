package testutil

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/moby/buildkit-bench/util/gotest"
	"github.com/stretchr/testify/require"
)

func TestWriteExecutionMetadata(t *testing.T) {
	dir := t.TempDir()
	oldBuildkit, oldBuildx := buildkitBinsDir, buildxBinsDir
	buildkitBinsDir, buildxBinsDir = filepath.Join(dir, "buildkit"), filepath.Join(dir, "buildx")
	t.Cleanup(func() { buildkitBinsDir, buildxBinsDir = oldBuildkit, oldBuildx })
	require.NoError(t, os.MkdirAll(filepath.Join(buildkitBinsDir, "master"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(buildxBinsDir, "v1.0.0"), 0755))
	// Non-executable fixtures still have identities; unavailable version output
	// must remain absent instead of being inferred from a candidate's label.
	data := []byte("binary fixture")
	require.NoError(t, os.WriteFile(filepath.Join(buildkitBinsDir, "master", "buildkitd"), data, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(buildxBinsDir, "v1.0.0", "buildx"), data, 0644))
	t.Setenv("TEST_BENCH_RUN", "2")
	t.Setenv("TEST_KEEP_CACHE", "0")
	t.Setenv("BENCH_CACHE_REUSED", "1")
	t.Setenv("GITHUB_RUN_ID", "123")
	t.Setenv("ImageVersion", "runner-image-version")
	t.Setenv("BENCH_IMAGE_ID", "sha256:test-image")
	t.Setenv("SECRET", "must-not-be-exported")
	output := filepath.Join(dir, "execution.json")
	require.NoError(t, WriteExecutionMetadata(output))
	dt, err := os.ReadFile(output)
	require.NoError(t, err)
	require.NotContains(t, string(dt), "must-not-be-exported")
	var execution gotest.Execution
	require.NoError(t, json.Unmarshal(dt, &execution))
	require.NotEmpty(t, execution.ID)
	require.False(t, execution.CapturedAt.IsZero())
	require.NotEmpty(t, execution.GoVersion)
	require.Positive(t, execution.NumCPU)
	require.Equal(t, "2", execution.Settings["TEST_BENCH_RUN"])
	require.Equal(t, "0", execution.Settings["TEST_KEEP_CACHE"])
	require.Equal(t, "1", execution.Settings["BENCH_CACHE_REUSED"])
	require.NotContains(t, execution.Environment, "suite_modified")
	require.Equal(t, flag.Lookup("test.benchtime").Value.String(), execution.Settings["test.benchtime"])
	require.Equal(t, "runner-image-version", execution.Environment["ImageVersion"])
	require.Equal(t, "sha256:test-image", execution.Environment["BENCH_IMAGE_ID"])
	require.Len(t, execution.Binaries, 2)
	hash := sha256.Sum256(data)
	require.Equal(t, hex.EncodeToString(hash[:]), execution.Binaries[0].SHA256)
	require.Equal(t, "buildkit", execution.Binaries[0].Project)
	require.Equal(t, "master", execution.Binaries[0].Ref)
	require.Empty(t, execution.Binaries[0].Version)
	require.Equal(t, "buildx", execution.Binaries[1].Project)
	require.Equal(t, "v1.0.0", execution.Binaries[1].Ref)
}
