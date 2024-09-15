package main

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/moby/buildkit-bench/util/candidates"
	"github.com/moby/buildkit-bench/util/gotest"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/montanaflynn/stats"
	"github.com/pkg/errors"
	"github.com/zeebo/xxh3"
)

type genCmd struct {
	Files []string `kong:"arg='',name='files',required,help='Benchmark results files generated by parse command.'"`

	Config     string `kong:"name='config',required,default='testconfig.yml',help='Test config file.'"`
	Output     string `kong:"name='output',default='./bin/gen/benchmarks.html',help='File to write the HTML report to.'"`
	Candidates string `kong:"name='candidates',help='Candidates file.'"`
}

func (c *genCmd) Run(ctx *Context) error {
	benchmarks, err := gotest.MergeBenchmarks(c.Files)
	if err != nil {
		return err
	}
	if ctx.Debug {
		b, _ := json.MarshalIndent(benchmarks, "", "  ")
		log.Printf("%s", string(b))
	}
	if err := c.validateBenchmarks(benchmarks); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.Output), 0755); err != nil {
		return errors.Wrap(err, "failed to create output directory")
	}
	return c.writeHTML(benchmarks)
}

func (c *genCmd) validateBenchmarks(benchmarks map[string]gotest.Benchmark) error {
	log.Printf("Validating %d benchmark results based on %s", len(benchmarks), c.Config)
	tc, err := testutil.LoadTestConfig(c.Config)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{})
	for _, benchmark := range benchmarks {
		if _, ok := seen[benchmark.Name]; !ok {
			seen[benchmark.Name] = struct{}{}
		}
		bm, err := tc.BenchmarkConfig(benchmark.Name)
		if err != nil {
			return err
		}
		for _, run := range benchmark.Runs {
			for unit := range run.Extra {
				if _, ok := bm.Metrics[unit]; !ok {
					return errors.Errorf("unknown metric %q for benchmark %q", unit, benchmark.Name)
				}
			}
		}
	}
	for rootName, bms := range tc.Runs {
		for testName := range bms {
			if _, ok := seen[rootName+"/"+testName]; !ok {
				return errors.Errorf("missing benchmark result for %q", rootName+"/"+testName)
			}
		}
	}
	return nil
}

func (c *genCmd) writeHTML(benchmarks map[string]gotest.Benchmark) error {
	log.Printf("Generating HTML report to %s", c.Output)
	benchmarksRuns := make(map[string][]gotest.BenchmarkRun)
	for _, benchmark := range benchmarks {
		br, ok := benchmarksRuns[benchmark.Name]
		if ok {
			benchmarksRuns[benchmark.Name] = append(br, benchmark.Runs...)
		} else {
			benchmarksRuns[benchmark.Name] = benchmark.Runs
		}
	}

	tc, err := testutil.LoadTestConfig(c.Config)
	if err != nil {
		return err
	}

	var sortedRefs []candidates.Ref
	if c.Candidates != "" {
		cds, err := candidates.Load(c.Candidates)
		if err != nil {
			return errors.Wrapf(err, "failed to load candidates from %s", c.Candidates)
		}
		sortedRefs = cds.Sorted()
	}

	var cps []components.Charter

	benchmarkKeys := make([]string, 0, len(benchmarksRuns))
	for k := range benchmarksRuns {
		benchmarkKeys = append(benchmarkKeys, k)
	}
	sort.Strings(benchmarkKeys)
	for _, name := range benchmarkKeys {
		runs := benchmarksRuns[name]
		bc, err := tc.BenchmarkConfig(name)
		if err != nil {
			return err
		}

		metrics := make(map[string]map[string][]float64)
		for _, run := range runs {
			for unit := range bc.Metrics {
				if _, ok := metrics[unit]; !ok {
					metrics[unit] = make(map[string][]float64)
				}
				if v, ok := run.Extra[unit]; ok {
					metrics[unit][run.Ref] = append(metrics[unit][run.Ref], v)
				} else if unit == "duration" {
					metrics[unit][run.Ref] = append(metrics[unit][run.Ref], time.Duration(run.NsPerOp).Seconds())
				} else {
					return errors.Errorf("missing metric %q for run %s", unit, run.Ref)
				}
			}
		}

		for unit, values := range metrics {
			globalOptions := []charts.GlobalOpts{
				charts.WithTitleOpts(opts.Title{
					Title:    bc.Description,
					Subtitle: name,
				}),
				charts.WithDataZoomOpts(opts.DataZoom{
					Type: "slider",
				}),
			}
			switch bc.Metrics[unit].Chart {
			case types.ChartBar:
				chart, err := chartBar(globalOptions, bc.Metrics[unit], sortedRefs, name, unit, values)
				if err != nil {
					return err
				}
				cps = append(cps, chart)
			case types.ChartBoxPlot:
				chart, err := chartBoxPlot(globalOptions, bc.Metrics[unit], sortedRefs, name, unit, values)
				if err != nil {
					return err
				}
				cps = append(cps, chart)
			default:
				return errors.Errorf("unknown chart type %q for metric %q", bc.Metrics[unit].Chart, unit)
			}
		}
	}

	page := components.NewPage()
	page.PageTitle = "BuildKit Benchmarks"
	page.Layout = components.PageFlexLayout
	page.AddCharts(cps...)

	f, err := os.Create(c.Output)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer f.Close()

	if err := page.Render(f); err != nil {
		return errors.Wrap(err, "failed to render page")
	}

	return nil
}

func chartBar(globalOpts []charts.GlobalOpts, cfg testutil.TestConfigMetric, sortedRefs []candidates.Ref, name, unit string, values map[string][]float64) (components.Charter, error) {
	var refs []string
	var data []opts.BarData
	var allv []float64
	var total float64
	if len(sortedRefs) == 0 {
		for ref, v := range values {
			allv = append(allv, v...)
			m, err := stats.Median(v)
			if err != nil {
				return nil, errors.Wrap(err, "failed to calculate median")
			}
			total += m
			refs = append(refs, ref)
			mr, err := stats.Round(m, 5)
			if err != nil {
				return nil, errors.Wrap(err, "failed to round median")
			}
			data = append(data, opts.BarData{Value: mr})
		}
	} else {
		for _, ref := range sortedRefs {
			v, ok := values[ref.Name]
			if !ok {
				continue
			}
			allv = append(allv, v...)
			m, err := stats.Median(v)
			if err != nil {
				return nil, errors.Wrap(err, "failed to calculate median")
			}
			total += m
			refs = append(refs, ref.Name)
			mr, err := stats.Round(m, 5)
			if err != nil {
				return nil, errors.Wrap(err, "failed to round median")
			}
			data = append(data, opts.BarData{Value: mr})
		}
	}

	chart := charts.NewBar()
	chart.ChartID = chartIdentity(name, unit)
	chart.SetGlobalOptions(globalOpts...)

	chart.SetXAxis(refs).AddSeries(cfg.Description, data)

	if cfg.Average {
		avgv := total / float64(len(refs))
		avgdt := make([]opts.LineData, len(refs))
		for i := 0; i < len(refs); i++ {
			avgdt[i] = opts.LineData{Value: avgv}
		}
		avgl := charts.NewLine()
		avgl.SetXAxis(refs).AddSeries("Average", avgdt)
		chart.Overlap(avgl)
	}

	return chart, nil
}

func chartBoxPlot(globalOpts []charts.GlobalOpts, cfg testutil.TestConfigMetric, sortedRefs []candidates.Ref, name, unit string, values map[string][]float64) (components.Charter, error) {
	var refs []string
	var data []opts.BoxPlotData
	var allv []float64
	if len(sortedRefs) == 0 {
		for ref, v := range values {
			allv = append(allv, v...)
			refs = append(refs, ref)
			plotData, err := createBoxPlotData(v)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create box plot data")
			}
			data = append(data, opts.BoxPlotData{Value: plotData})
		}
	} else {
		for _, ref := range sortedRefs {
			v, ok := values[ref.Name]
			if !ok {
				continue
			}
			allv = append(allv, v...)
			refs = append(refs, ref.Name)
			plotData, err := createBoxPlotData(v)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create box plot data")
			}
			data = append(data, opts.BoxPlotData{Value: plotData})
		}
	}

	chart := charts.NewBoxPlot()
	chart.ChartID = chartIdentity(name, unit)
	chart.SetGlobalOptions(globalOpts...)

	chart.SetXAxis(refs).AddSeries(cfg.Description, data)

	return chart, nil
}

func chartIdentity(name, unit string) string {
	h := xxh3.New()
	h.WriteString(name)
	h.Write([]byte{0})
	h.WriteString(unit)
	h.Write([]byte{0})
	return strconv.FormatUint(h.Sum64(), 10)
}

func createBoxPlotData(data []float64) ([]float64, error) {
	minv, err := stats.Min(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate min")
	}
	maxv, err := stats.Max(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate max")
	}
	q, err := stats.Quartile(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate quartiles")
	}
	values := []float64{
		minv,
		q.Q1,
		q.Q2,
		q.Q3,
		maxv,
	}
	// replace NaN values with the previous or next value
	for i, v := range values {
		if math.IsNaN(v) {
			if i == 0 {
				values[i] = values[i+1]
			} else {
				values[i] = values[i-1]
			}
		}
	}
	return values, nil
}
