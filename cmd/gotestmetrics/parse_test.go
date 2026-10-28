package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/moby/buildkit-bench/util/gotest"
	"github.com/stretchr/testify/require"
)

func TestParseExecutionMetadata(t *testing.T) {
	for _, tc := range []struct {
		name      string
		metadata  string
		missing   bool
		wantError bool
	}{
		{
			name:     "captured",
			metadata: `{"id":"execution-1","go_version":"go1.26.8","settings":{"TEST_BENCH_RUN":"2"}}`,
		},
		{
			name:    "custom package without metadata",
			missing: true,
		},
		{
			name:      "malformed metadata",
			metadata:  `{`,
			wantError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "events.json")
			require.NoError(t, os.WriteFile(input, []byte(`{"Action":"output","Package":"bench","Output":"BenchmarkBuild/Simple/ref=master/run=1-4\t1\t123 ns/op\n"}`+"\n"), 0644))
			f, err := os.Open(input)
			require.NoError(t, err)
			stdin := os.Stdin
			os.Stdin = f
			t.Cleanup(func() { os.Stdin = stdin; f.Close() })
			metadataFile := filepath.Join(dir, "execution.json")
			if !tc.missing {
				require.NoError(t, os.WriteFile(metadataFile, []byte(tc.metadata), 0644))
			}
			cmd := &parseCmd{Output: filepath.Join(dir, "parsed.json"), Execution: metadataFile}
			err = cmd.Run(&Context{})
			if tc.wantError {
				require.ErrorContains(t, err, "failed to decode execution metadata")
				return
			}
			require.NoError(t, err)
			dt, err := os.ReadFile(cmd.Output)
			require.NoError(t, err)
			var result gotest.Result
			require.NoError(t, json.Unmarshal(dt, &result))
			require.Len(t, result.Benchmarks, 1)
			require.Empty(t, result.OutputLog)
			logs, err := filepath.Glob(filepath.Join(dir, "*.log.gz"))
			require.NoError(t, err)
			require.Empty(t, logs)
			if tc.missing {
				require.Nil(t, result.Execution)
			} else {
				require.Equal(t, "execution-1", result.Execution.ID)
				require.Equal(t, "2", result.Execution.Settings["TEST_BENCH_RUN"])
			}
		})
	}
}

func TestParsePreservesFailedResults(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "events.json")
	require.NoError(t, os.WriteFile(input, []byte(`{"Action":"output","Package":"bench","Output":"BenchmarkBuild/Simple/ref=master/run=1-4\t1\t123 ns/op\n"}
{"Action":"output","Package":"bench","Output":"--- FAIL: BenchmarkBuild/Simple/ref=old/run=1\n"}
{"Action":"output","Package":"bench","Output":"    build failed\n"}
{"Action":"fail","Package":"bench"}
`), 0644))
	events, err := os.ReadFile(input)
	require.NoError(t, err)
	fullOutput := "assertion failed\n" + strings.Repeat("daemon log\n", 2000) + "middle diagnostic\n" + strings.Repeat("daemon log\n", 2000) + "last message\n"
	outputEvent, err := json.Marshal(map[string]string{"Action": "output", "Package": "bench", "Output": fullOutput})
	require.NoError(t, err)
	events = []byte(strings.Replace(string(events), `{"Action":"fail"`, string(outputEvent)+"\n"+`{"Action":"fail"`, 1))
	require.NoError(t, os.WriteFile(input, events, 0644))
	f, err := os.Open(input)
	require.NoError(t, err)
	stdin := os.Stdin
	os.Stdin = f
	t.Cleanup(func() { os.Stdin = stdin; f.Close() })
	cmd := &parseCmd{Output: filepath.Join(dir, "parsed.json")}
	require.ErrorContains(t, cmd.Run(&Context{}), "results have been preserved")
	dt, err := os.ReadFile(cmd.Output)
	require.NoError(t, err)
	var result gotest.Result
	require.NoError(t, json.Unmarshal(dt, &result))
	require.Len(t, result.Benchmarks, 1)
	require.Len(t, result.Failures, 1)
	require.Contains(t, result.Failures[0].Output, "build failed")
	require.Contains(t, result.Failures[0].Output, "output truncated")
	require.NotContains(t, result.Failures[0].Output, "middle diagnostic")
	require.NotEmpty(t, result.OutputLog)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(dir, result.OutputLog))
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0644), info.Mode().Perm())
	}
	logFile, err := os.Open(filepath.Join(dir, result.OutputLog))
	require.NoError(t, err)
	defer logFile.Close()
	compressed, err := gzip.NewReader(logFile)
	require.NoError(t, err)
	defer compressed.Close()
	output, err := io.ReadAll(compressed)
	require.NoError(t, err)
	require.Contains(t, string(output), fullOutput)
	require.NoError(t, compressed.Close())
	require.NoError(t, logFile.Close())
	// An unavailable sidecar destination must not discard primary results.
	logPath := filepath.Join(dir, result.OutputLog)
	require.NoError(t, os.Remove(logPath))
	require.NoError(t, os.Mkdir(logPath, 0755))
	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)
	require.ErrorContains(t, cmd.Run(&Context{}), "results have been preserved")
	dt, err = os.ReadFile(cmd.Output)
	require.NoError(t, err)
	var withoutLog gotest.Result
	require.NoError(t, json.Unmarshal(dt, &withoutLog))
	require.Empty(t, withoutLog.OutputLog)
	require.Len(t, withoutLog.Failures, 1)
	require.Len(t, withoutLog.Benchmarks, 1)
}
