package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestReportOutputFlags(t *testing.T) {
	var args struct {
		Gen genCmd `kong:"cmd"`
	}
	parser, err := kong.New(&args)
	require.NoError(t, err)
	_, err = parser.Parse([]string{"gen", "input.json"})
	require.NoError(t, err)
	require.Equal(t, "./bin/gen/benchmarks.html", args.Gen.HTMLOutput)
	require.Equal(t, "./bin/gen/benchmarks.json", args.Gen.JSONOutput)
	_, err = parser.Parse([]string{"gen", "--html-output=custom.html", "input.json"})
	require.NoError(t, err)
	require.Equal(t, "custom.html", args.Gen.HTMLOutput)
	require.Equal(t, "./bin/gen/benchmarks.json", args.Gen.JSONOutput)
	_, err = parser.Parse([]string{"gen", "--html-output=custom.html", "--json-output=custom.json", "input.json"})
	require.NoError(t, err)
	require.Equal(t, "custom.json", args.Gen.JSONOutput)
}

func TestGenerateJSONReport(t *testing.T) {
	for _, project := range []string{"buildkit", "buildx"} {
		t.Run(project, func(t *testing.T) {
			dir := t.TempDir()
			write := func(name, content string) string {
				t.Helper()
				file := filepath.Join(dir, name)
				require.NoError(t, os.WriteFile(file, []byte(content), 0644))
				return file
			}
			config := write("config.yml", `defaults:
  count: 3
  benchtime: 1x
runs:
  BenchmarkBuild:
    BenchmarkSimple:
      description: Simple build
      count: 4
      metrics:
        duration:
          description: Time (s)
          chart: boxplot
        alloc:
          description: Allocated memory (bytes)
          chart: boxplot
`)
			input := write("input.json", `{
  "Execution": {"id":"execution-a","go_version":"go1.26.8","settings":{"TEST_BENCH_RUN":"2"},"environment":{"ImageVersion":"runner-a"},"binaries":[{"project":"buildkit","ref":"master","name":"buildkitd","sha256":"actual-binary-hash","version":"buildkitd actual-build"}]},
  "BenchmarkInfo": {"OS":"linux","Architecture":"amd64","CPU":"cpu-a","Package":"bench"},
  "Benchmarks": {
    "simple": {
      "Name":"BenchmarkBuild/BenchmarkSimple", "Ref":"master",
      "Runs":[
        {"Ref":"master","Buildx":"v1.0.0","Run":1,"N":2,"Threads":8,
         "Time":"2026-10-13T03:00:00Z","NsPerOp":1500000000,"Measured":15,
         "AllocedBytesPerOp":123,"AllocsPerOp":0,"MBPerS":4.5,
         "Extra":{"alloc":100}},
        {"Ref":"master","Buildx":"v1.0.0","Run":2,"N":1,"Threads":8,
         "Time":"2026-10-13T03:01:00Z","NsPerOp":9000000000,
         "Extra":{"duration":2.5,"alloc":0}}
      ]
    }
  }
}`)
			candidateFile := write("candidates.json", `{"refs":{
  "master":{"sha":"buildkit-sha","date":"2026-10-12T00:00:00Z"},
  "v1.0.0":{"sha":"buildx-sha","date":"2026-10-11T00:00:00Z"}
}}`)
			envs := write("env.txt", "GITHUB_REPOSITORY=moby/buildkit-bench\nGITHUB_SHA=suite-sha\nGITHUB_RUN_ID=123\nGITHUB_RUN_ATTEMPT=2\nGITHUB_EVENT_NAME=schedule\nSECRET=never-export\n")
			cmd := &genCmd{
				Files:          []string{input},
				Name:           "20261013-030000",
				Project:        project,
				Config:         config,
				HTMLOutput:     filepath.Join(dir, "report", project+".html"),
				JSONOutput:     filepath.Join(dir, "report", project+".json"),
				Candidates:     candidateFile,
				Envs:           envs,
				ValidationMode: "strict",
			}
			require.NoError(t, cmd.Run(&Context{}))
			html, err := os.ReadFile(cmd.HTMLOutput)
			require.NoError(t, err)
			require.Contains(t, string(html), "Simple build")
			dt, err := os.ReadFile(filepath.Join(dir, "report", project+".json"))
			require.NoError(t, err)
			require.NotContains(t, string(dt), "never-export")
			var report map[string]any
			require.NoError(t, json.Unmarshal(dt, &report))
			require.Equal(t, float64(1), report["version"])
			require.Equal(t, project, report["project"])
			metadata := report["metadata"].(map[string]any)
			require.Equal(t, "suite-sha", metadata["commit"])
			require.Equal(t, "https://github.com/moby/buildkit-bench/actions/runs/123/attempts/2", metadata["run_url"])
			refs := report["candidates"].(map[string]any)[project].([]any)
			require.Equal(t, "buildx-sha", refs[0].(map[string]any)["sha"])
			benchmarks := report["benchmarks"].([]any)
			require.Len(t, benchmarks, 1)
			bm := benchmarks[0].(map[string]any)
			require.Equal(t, float64(4), bm["count"])
			require.Equal(t, "1x", bm["benchtime"])
			metrics := bm["metrics"].(map[string]any)
			require.Equal(t, "s", metrics["duration"].(map[string]any)["unit"])
			require.Equal(t, "bytes", metrics["alloc"].(map[string]any)["unit"])
			samples := bm["samples"].([]any)
			require.Len(t, samples, 2)
			first := samples[0].(map[string]any)
			require.Equal(t, "master", first["buildkit"])
			require.Equal(t, "v1.0.0", first["buildx"])
			require.Equal(t, "2026-10-13T03:00:00Z", first["time"])
			require.Equal(t, "cpu-a", first["cpu"])
			require.Equal(t, float64(2), first["iterations"])
			require.Equal(t, float64(1.5), first["values"].(map[string]any)["duration"])
			raw := first["measurements"].(map[string]any)
			require.Equal(t, float64(1500000000), raw["ns/op"])
			require.Equal(t, float64(123), raw["B/op"])
			require.Equal(t, float64(0), raw["allocs/op"])
			require.Equal(t, float64(4.5), raw["MB/s"])
			require.Equal(t, float64(100), raw["alloc"])
			inputBytes, err := os.ReadFile(input)
			require.NoError(t, err)
			inputHash := sha256.Sum256(inputBytes)
			sourceID := hex.EncodeToString(inputHash[:])
			require.Equal(t, sourceID, first["source_id"])
			source := report["sources"].(map[string]any)[sourceID].(map[string]any)
			execution := source["execution"].(map[string]any)
			require.Equal(t, "execution-a", execution["id"])
			require.Equal(t, "2", execution["settings"].(map[string]any)["TEST_BENCH_RUN"])
			require.Equal(t, "actual-binary-hash", execution["binaries"].([]any)[0].(map[string]any)["sha256"])
			second := samples[1].(map[string]any)
			require.Equal(t, float64(2.5), second["values"].(map[string]any)["duration"])
			require.Equal(t, float64(0), second["values"].(map[string]any)["alloc"])
			require.NotContains(t, second["measurements"].(map[string]any), "ns/op")
			configBytes, err := os.ReadFile(config)
			require.NoError(t, err)
			configHash := sha256.Sum256(configBytes)
			require.Equal(t, hex.EncodeToString(configHash[:]), report["generator"].(map[string]any)["config_sha256"])

			// Older inputs have no timestamp; repeated keys from separate runners
			// must retain their measurements and provenance in the JSON report.
			legacy := write("legacy.json", `{"BenchmarkInfo":{"CPU":"cpu-b"},"Benchmarks":{
  "simple":{"Name":"BenchmarkBuild/BenchmarkSimple","Ref":"master","Runs":[
    {"Ref":"master","Buildx":"v1.0.0","Run":1,"N":1,"NsPerOp":3000000000,"Extra":{"alloc":200}}
  ]}
}}`)
			cmd.Files = []string{input, legacy, input}
			cmd.JSONOutput = filepath.Join(dir, "custom", "data.json")
			combined, err := cmd.jsonReport()
			require.NoError(t, err)
			require.NoError(t, cmd.writeJSON(combined))
			updated, err := os.ReadFile(cmd.JSONOutput)
			require.NoError(t, err)
			var decoded benchmarkReport
			require.NoError(t, json.Unmarshal(updated, &decoded))
			require.Len(t, decoded.Benchmarks[0].Samples, 3)
			require.Equal(t, "cpu-b", decoded.Benchmarks[0].Samples[2].CPU)
			require.True(t, decoded.Benchmarks[0].Samples[2].Time.IsZero())
			legacySource := decoded.Benchmarks[0].Samples[2].SourceID
			require.Nil(t, decoded.Sources[legacySource].Execution)
			require.NotContains(t, string(updated), "0001-01-01")
			combined, err = cmd.jsonReport()
			require.NoError(t, err)
			require.NoError(t, cmd.writeJSON(combined))
			repeated, err := os.ReadFile(cmd.JSONOutput)
			require.NoError(t, err)
			require.Equal(t, updated, repeated)

			cmd.JSONOutput = cmd.HTMLOutput
			require.ErrorContains(t, cmd.Run(&Context{}), "output paths must differ")
			unchanged, err := os.ReadFile(cmd.HTMLOutput)
			require.NoError(t, err)
			require.Equal(t, html, unchanged)

			cmd.Envs = ""
			cmd.Candidates = ""
			local, err := cmd.jsonReport()
			require.NoError(t, err)
			require.Empty(t, local.Candidates)
			require.Empty(t, local.Metadata)
			cmd.BuildkitCandidates = candidateFile
			cmd.BuildxCandidates = candidateFile
			withPeers, err := cmd.jsonReport()
			require.NoError(t, err)
			require.Len(t, withPeers.Candidates["buildkit"], 2)
			require.Len(t, withPeers.Candidates["buildx"], 2)
			cmd.Candidates = write("override.json", `{"refs":{"master":{"sha":"override-sha"}}}`)
			withOverride, err := cmd.jsonReport()
			require.NoError(t, err)
			require.Len(t, withOverride.Candidates[project], 1)
			require.Equal(t, "override-sha", withOverride.Candidates[project][0].SHA)

			cmd.Files = []string{write("missing.json", `{"Benchmarks":{
  "simple":{"Name":"BenchmarkBuild/BenchmarkSimple","Runs":[{"Ref":"master","Buildx":"v1.0.0","NsPerOp":100}]}
}}`)}
			_, err = cmd.jsonReport()
			require.ErrorContains(t, err, `missing metric "alloc"`)
			require.ErrorContains(t, err, "buildx v1.0.0")
			cmd.Files = []string{write("unknown.json", `{"Benchmarks":{
  "simple":{"Name":"BenchmarkBuild/BenchmarkSimple","Runs":[{"Ref":"master","NsPerOp":100,"Extra":{"alloc":1,"unknown":2}}]}
}}`), input}
			cmd.JSONOutput = filepath.Join(dir, "unknown-output.json")
			require.ErrorContains(t, cmd.Run(&Context{}), `unknown metric "unknown"`)
			unchanged, err = os.ReadFile(cmd.HTMLOutput)
			require.NoError(t, err)
			require.Equal(t, html, unchanged)
		})
	}
}

func TestBenchmarkChartComponentsCombinesDurationAndAlloc(t *testing.T) {
	name := "BenchmarkBuild/BenchmarkBuildSimple"
	bc := &testutil.TestConfigBenchmark{
		Description: "Simple build",
		Metrics: map[string]testutil.TestConfigMetric{
			string(testutil.MetricDuration): {
				Description: "Time (s)",
				Chart:       types.ChartBoxPlot,
			},
			string(testutil.MetricAlloc): {
				Description: "Allocated memory (bytes)",
				Chart:       types.ChartBoxPlot,
			},
		},
	}
	metrics := map[string]map[string][]float64{
		string(testutil.MetricDuration): {
			"ref-b": {2, 3, 4},
			"ref-a": {1, 2, 3},
		},
		string(testutil.MetricAlloc): {
			"ref-b": {200, 300, 400},
			"ref-a": {100, 200, 300},
		},
	}

	cps, err := benchmarkChartComponents(chartGlobalOptions(bc.Description, name), bc, nil, name, metrics)
	require.NoError(t, err)
	require.Len(t, cps, 1)

	chart, ok := cps[0].(*charts.BoxPlot)
	require.True(t, ok)
	chart.Validate()
	options := chart.JSON()

	series, ok := options["series"].(charts.MultiSeries)
	require.True(t, ok)
	require.Len(t, series, 2)
	require.Equal(t, "Time (s)", series[0].Name)
	require.Equal(t, 0, series[0].YAxisIndex)
	require.Equal(t, "Allocated memory (bytes)", series[1].Name)
	require.Equal(t, 1, series[1].YAxisIndex)

	durationData, ok := series[0].Data.([]opts.BoxPlotData)
	require.True(t, ok)
	require.Len(t, durationData, 2)
	allocData, ok := series[1].Data.([]opts.BoxPlotData)
	require.True(t, ok)
	require.Len(t, allocData, 2)

	xAxis, ok := options["xAxis"].([]opts.XAxis)
	require.True(t, ok)
	require.Len(t, xAxis, 1)
	require.Equal(t, []string{"ref-a", "ref-b"}, xAxis[0].Data)

	yAxis, ok := options["yAxis"].([]opts.YAxis)
	require.True(t, ok)
	require.Len(t, yAxis, 2)
	require.Equal(t, "Time (s)", yAxis[0].Name)
	require.Equal(t, "middle", yAxis[0].NameLocation)
	require.Equal(t, 45, yAxis[0].NameGap)
	require.Equal(t, "left", yAxis[0].Position)
	require.Equal(t, "Allocated memory (bytes)", yAxis[1].Name)
	require.Equal(t, "middle", yAxis[1].NameLocation)
	require.Equal(t, 45, yAxis[1].NameGap)
	require.Equal(t, "right", yAxis[1].Position)
}
