package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-githubactions"
	"gopkg.in/yaml.v3"
)

type listCmd struct {
	Config    string `kong:"name='config',required,default='testconfig.yml',help='Test config file.'"`
	GhaOutput string `kong:"name='gha-output',help='Set GitHub Actions output parameter to be used as matrix includes.'"`
}

type listInclude struct {
	Test      string `json:"test"`
	Count     int    `json:"count"`
	BenchTime string `json:"benchtime"`
}

func (c *listCmd) Run(ctx *Context) error {
	dt, err := os.ReadFile(c.Config)
	if err != nil {
		return errors.Wrap(err, "failed to read test config")
	}
	var testconfig testutil.TestConfig
	if err := yaml.Unmarshal(dt, &testconfig); err != nil {
		return errors.Wrap(err, "failed to decode test config")
	}
	if ctx.Debug {
		b, _ := json.MarshalIndent(testconfig, "", "  ")
		log.Printf("%s", string(b))
	}
	if c.GhaOutput != "" {
		var includes []listInclude
		for rootName, benchmarks := range testconfig.Runs {
			for benchmarkName, benchmark := range benchmarks {
				count := benchmark.Count
				if count == 0 {
					count = testconfig.Defaults.Count
				}
				benchtime := benchmark.Benchtime
				if benchtime == "" {
					benchtime = testconfig.Defaults.Benchtime
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
