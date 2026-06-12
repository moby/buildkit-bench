package main

import (
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

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
