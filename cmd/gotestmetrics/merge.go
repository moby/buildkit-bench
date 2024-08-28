package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/moby/buildkit-bench/util/gotest"
	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/pkg/errors"
)

type mergeCmd struct {
	Config string `kong:"name='config',required,default='testconfig.yml',help='Test config file.'"`
	Dir    string `kong:"arg='',name='dir',required,help='Directory containing benchmark results to merge.'"`
	Format string `kong:"name='format',default='json',help='Format of the benchmark results.'"`
	Output string `kong:"name='output',default='./bin/benchmarks',help='Directory to write the merged results to.'"`
}

func (c *mergeCmd) Run(ctx *Context) error {
	benchmarks, err := gotest.MergeBenchmarks(c.Dir)
	if err != nil {
		return err
	}
	if ctx.Debug {
		b, _ := json.MarshalIndent(benchmarks, "", "  ")
		log.Printf("%s", string(b))
	}
	switch c.Format {
	case "json":
		return c.writeBenchmarksJSON(benchmarks)
	case "html":
		return c.writeBenchmarksHTML(benchmarks)
	default:
		return errors.Errorf("unsupported format: %s", c.Format)
	}
}

func (c *mergeCmd) writeBenchmarksJSON(benchmarks map[string]gotest.Benchmark) error {
	b, err := json.MarshalIndent(benchmarks, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal benchmarks")
	}
	if err := os.MkdirAll(c.Output, 0755); err != nil {
		return errors.Wrap(err, "failed to create output file directory")
	}
	return os.WriteFile(path.Join(c.Output, "benchmarks.json"), b, 0644)
}

func (c *mergeCmd) writeBenchmarksHTML(benchmarks map[string]gotest.Benchmark) error {
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

	var cps []components.Charter

	for name, runs := range benchmarksRuns {
		bc, err := tc.BenchmarkConfig(name)
		if err != nil {
			return err
		}
		metrics := make(map[string]map[string]float64)
		for _, run := range runs {
			for unit := range bc.Metrics {
				if _, ok := metrics[unit]; !ok {
					metrics[unit] = make(map[string]float64)
				}
				if v, ok := run.Extra[unit]; ok {
					metrics[unit][run.Ref] = v
				} else {
					return errors.Errorf("missing metric %q for run %s", unit, run.Ref)
				}
			}
			if _, ok := metrics["ref_timestamp"]; !ok {
				metrics["ref_timestamp"] = make(map[string]float64)
			}
			if v, ok := run.Extra["ref_timestamp"]; ok {
				metrics["ref_timestamp"][run.Ref] = v
			} else {
				return errors.Errorf("missing ref_timestamp for run %s", run.Ref)
			}
		}

		refTimestamps := make([]struct {
			Ref   string
			Value float64
		}, 0, len(metrics["ref_timestamp"]))
		for ref, value := range metrics["ref_timestamp"] {
			refTimestamps = append(refTimestamps, struct {
				Ref   string
				Value float64
			}{Ref: ref, Value: value})
		}
		sort.Slice(refTimestamps, func(i, j int) bool {
			return refTimestamps[i].Value < refTimestamps[j].Value
		})

		for unit, values := range metrics {
			if unit == "ref_timestamp" {
				continue
			}
			var refs []string
			var data []opts.BarData
			chart := charts.NewBar() // TODO: chart type should be in testconfig
			chart.SetGlobalOptions(
				charts.WithTitleOpts(opts.Title{Title: bc.Description}),
				charts.WithXAxisOpts(opts.XAxis{
					AxisLabel: &opts.AxisLabel{
						Interval: "0", // Ensure all labels are displayed
					},
				}),
			)
			for _, entry := range refTimestamps {
				refs = append(refs, entry.Ref)
				data = append(data, opts.BarData{Value: values[entry.Ref]})
			}
			chart.SetXAxis(refs).AddSeries(bc.Metrics[unit].Description, data)
			cps = append(cps, chart)
		}
	}

	page := components.NewPage()
	page.AddCharts(cps...)

	if err := os.MkdirAll(c.Output, 0755); err != nil {
		return errors.Wrap(err, "failed to create output file directory")
	}
	f, err := os.Create(path.Join(c.Output, "index.html"))
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer f.Close()

	if err := page.Render(f); err != nil {
		return errors.Wrap(err, "failed to render page")
	}

	return nil
}
