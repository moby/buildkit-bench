package main

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/moby/buildkit-bench/util/gotest"
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
