package testutil

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/moby/buildkit-bench/util/gotest"
	"github.com/moby/buildkit/identity"
	"github.com/pkg/errors"
)

// WriteExecutionMetadata records provenance before any benchmark timers start.
// The caller must parse testing flags first so overrides reflect actual settings.
func WriteExecutionMetadata(filename string) error {
	metadata := gotest.Execution{
		ID: identity.NewID(), CapturedAt: time.Now().UTC(),
		GoVersion: runtime.Version(), NumCPU: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0),
		Settings: make(map[string]string), Environment: make(map[string]string),
		Binaries: make([]gotest.BinaryIdentity, 0),
	}
	for _, name := range []string{"test.bench", "test.benchtime", "test.count", "test.cpu", "test.parallel", "test.run", "test.timeout", "test.benchmem"} {
		if f := flag.Lookup(name); f != nil {
			metadata.Settings[name] = f.Value.String()
		}
	}
	for _, name := range []string{"TEST_BENCH_RUN", "TEST_KEEP_CACHE", "BENCH_CACHE_REUSED", "BUILDKIT_REF_RANDOM", "GOGC", "GOMEMLIMIT", "GODEBUG", "GOMAXPROCS"} {
		if value, ok := os.LookupEnv(name); ok {
			metadata.Settings[name] = value
		}
	}
	// Host metadata is explicitly forwarded by hack/test. Never copy the full
	// environment, or inspect the later report job's environment for these fields.
	for _, name := range []string{"GITHUB_REPOSITORY", "GITHUB_SHA", "GITHUB_REF", "GITHUB_RUN_ID", "GITHUB_RUN_ATTEMPT", "GITHUB_JOB", "GITHUB_EVENT_NAME", "RUNNER_ENVIRONMENT", "ImageOS", "ImageVersion", "BENCH_DOCKER_VERSION", "BENCH_IMAGE_ID"} {
		if value := os.Getenv(name); value != "" {
			metadata.Environment[name] = value
		}
	}
	for name, file := range map[string]string{
		"kernel":                "/proc/sys/kernel/osrelease",
		"os_release":            "/etc/os-release",
		"memory_max":            "/sys/fs/cgroup/memory.max",
		"memory_limit_in_bytes": "/sys/fs/cgroup/memory/memory.limit_in_bytes",
		"cpu_max":               "/sys/fs/cgroup/cpu.max",
		"cpu_quota_us":          "/sys/fs/cgroup/cpu/cpu.cfs_quota_us",
		"cpu_period_us":         "/sys/fs/cgroup/cpu/cpu.cfs_period_us",
	} {
		if dt, err := os.ReadFile(file); err == nil {
			metadata.Environment[name] = strings.TrimSpace(string(dt))
		}
	}
	if dt, err := os.ReadFile("/proc/meminfo"); err == nil {
		for _, line := range strings.Split(string(dt), "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				metadata.Environment["memory_total"] = strings.TrimSpace(strings.TrimPrefix(line, "MemTotal:"))
			}
		}
	}
	if revision := executionCommand("git", "rev-parse", "HEAD"); revision != "" {
		metadata.Environment["suite_revision"] = revision
	}
	for _, project := range []struct {
		name     string
		dir      string
		binaries []string
	}{
		{
			name:     "buildkit",
			dir:      buildkitBinsDir,
			binaries: []string{"buildkitd", "buildctl", "buildkit-runc"},
		},
		{
			name:     "buildx",
			dir:      buildxBinsDir,
			binaries: []string{"buildx"},
		},
	} {
		entries, err := os.ReadDir(project.dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return errors.Wrap(err, "failed to list benchmark binaries")
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			for _, name := range project.binaries {
				file := filepath.Join(project.dir, entry.Name(), name)
				f, err := os.Open(file)
				if os.IsNotExist(err) {
					continue
				}
				if err != nil {
					return err
				}
				hash := sha256.New()
				_, err = io.Copy(hash, f)
				f.Close()
				if err != nil {
					return errors.Wrapf(err, "failed to hash %s", file)
				}
				arg := "--version"
				if name == "buildx" {
					arg = "version"
				}
				metadata.Binaries = append(metadata.Binaries, gotest.BinaryIdentity{
					Project: project.name, Ref: entry.Name(), Name: name,
					SHA256: hex.EncodeToString(hash.Sum(nil)), Version: executionCommand(file, arg),
				})
			}
		}
	}
	dt, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, dt, 0644)
}

// Version and OS details are optional; a missing tool must not invent metadata.
func executionCommand(name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
