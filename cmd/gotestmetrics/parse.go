package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/moby/buildkit-bench/util/gotest"
	"github.com/pkg/errors"
)

type parseCmd struct {
	Output string `kong:"name='output',help='File to write the JSON output to.'"`
}

func (c *parseCmd) Run(ctx *Context) error {
	pr, pw := io.Pipe()
	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, pw)

	errCh := make(chan error, 1)

	go func() {
		defer pr.Close()
		res, ex, err := gotest.Parse(gotest.ParseConfig{
			Stdout: pr,
			Logger: log.Writer(),
		})
		if err != nil {
			errCh <- errors.Wrap(err, "failed to scan test output")
			return
		}

		if ctx.Debug {
			b, _ := json.MarshalIndent(res, "", "  ")
			log.Printf("%s", string(b))
		}

		if len(ex.Failed()) > 0 {
			errCh <- errors.Errorf("%d test(s) failed", len(ex.Failed()))
			return
		}

		if c.Output != "" {
			if err := os.MkdirAll(filepath.Dir(c.Output), 0755); err != nil {
				errCh <- errors.Wrap(err, "failed to create output file directory")
				return
			}
			dt, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				errCh <- errors.Wrap(err, "failed to marshal result")
				return
			}
			if err := os.WriteFile(c.Output, dt, 0644); err != nil {
				errCh <- errors.Wrap(err, "failed to write result to output file")
				return
			}
		}

		errCh <- nil
	}()

	_, err := io.Copy(mw, os.Stdin)
	if err != nil {
		return errors.Wrap(err, "failed to copy test output")
	}

	pw.Close()

	_, err = io.Copy(os.Stdout, &buf)
	if err != nil {
		return errors.Wrap(err, "failed to write buffer to stdout")
	}

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}
