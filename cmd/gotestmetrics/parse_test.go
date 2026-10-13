package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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
			if tc.missing {
				require.Nil(t, result.Execution)
			} else {
				require.Equal(t, "execution-1", result.Execution.ID)
				require.Equal(t, "2", result.Execution.Settings["TEST_BENCH_RUN"])
			}
		})
	}
}
