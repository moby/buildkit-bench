package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/moby/buildkit-bench/util/gotest"
	"github.com/pkg/errors"
)

type parseCmd struct {
	Output    string `kong:"name='output',help='File to write the JSON output to.'"`
	Execution string `kong:"name='execution',help='Optional metadata file captured by the benchmark process.'"`
}

func (c *parseCmd) Run(ctx *Context) error {
	res, ex, err := gotest.Parse(gotest.ParseConfig{
		Stdout: os.Stdin,
		Logger: log.Writer(),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to scan test output")
	}
	if c.Execution != "" {
		dt, err := os.ReadFile(c.Execution)
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to read execution metadata")
		}
		if err == nil {
			var execution gotest.Execution
			if err := json.Unmarshal(dt, &execution); err != nil {
				return errors.Wrap(err, "failed to decode execution metadata")
			}
			res.Execution = &execution
		}
	}

	if ctx.Debug {
		b, _ := json.MarshalIndent(res, "", "  ")
		log.Printf("%s", string(b))
	}

	if len(ex.Failed()) > 0 {
		return errors.Errorf("%d test(s) failed", len(ex.Failed()))
	}

	if c.Output != "" {
		if err := os.MkdirAll(filepath.Dir(c.Output), 0755); err != nil {
			return errors.Wrap(err, "failed to create output file directory")
		}
		dt, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return errors.Wrap(err, "failed to marshal result")
		}
		if err := os.WriteFile(c.Output, dt, 0644); err != nil {
			return errors.Wrap(err, "failed to write result to output file")
		}
	}
	return nil
}
