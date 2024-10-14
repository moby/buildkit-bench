package main

import (
	"encoding/json"

	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-githubactions"
)

type listCmd struct {
	Config    string `kong:"name='config',required,default='testconfig.yml',help='Test config file.'"`
	Project   string `kong:"name='project',enum='buildkit,buildx',default='buildkit',help='Project type.'"`
	GhaOutput string `kong:"name='gha-output',help='Set GitHub Actions output parameter to be used as matrix includes.'"`
}

type listInclude struct {
	Test      string `json:"test"`
	Count     int    `json:"count"`
	BenchTime string `json:"benchtime"`
}

func (c *listCmd) Run(ctx *Context) error {
	tc, err := testutil.LoadTestConfig(c.Config)
	if err != nil {
		return err
	}
	if c.GhaOutput != "" {
		var includes []listInclude
		for rootName, benchmarks := range tc.Runs {
			for benchmarkName, benchmark := range benchmarks {
				if benchmark.Scope != "" && benchmark.Scope != c.Project {
					continue
				}
				count := benchmark.Count
				if count == 0 {
					count = tc.Defaults.Count
				}
				benchtime := benchmark.Benchtime
				if benchtime == "" {
					benchtime = tc.Defaults.Benchtime
				}
				includes = append(includes, listInclude{
					Test:      rootName + "/" + benchmarkName,
					Count:     count,
					BenchTime: benchtime,
				})
			}
		}
		dti, err := json.Marshal(includes)
		if err != nil {
			return errors.Wrap(err, "failed to marshal includes")
		}
		githubactions.SetOutput(c.GhaOutput, string(dti))
	}
	return nil
}
