package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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
	// Stream a compressed copy of the complete output to disk. Only failed
	// runs retain the sidecar alongside their bounded failure excerpts.
	var output io.Writer
	var logFile *os.File
	var compressed *gzip.Writer
	digest := sha256.New()
	if c.Output != "" {
		if err := os.MkdirAll(filepath.Dir(c.Output), 0755); err != nil {
			return err
		}
		var err error
		logFile, err = os.CreateTemp(filepath.Dir(c.Output), ".output-*.log.gz")
		if err != nil {
			log.Printf("Warning: full output log is unavailable: %v", err)
		} else {
			defer os.Remove(logFile.Name())
			defer logFile.Close()
			compressed = gzip.NewWriter(io.MultiWriter(logFile, digest))
			output = compressed
		}
	}
	res, ex, err := gotest.Parse(gotest.ParseConfig{
		Stdout: os.Stdin,
		Logger: log.Writer(),
		Output: output,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to scan test output")
	}
	failed := len(ex.Failed()) > 0 || len(res.Failures) > 0
	if compressed != nil {
		logErr := compressed.Close()
		if res.OutputError != nil {
			logErr = res.OutputError
		}
		if failed && logErr == nil {
			// The root-owned container output is uploaded by the host user.
			logErr = logFile.Chmod(0644)
		}
		if err := logFile.Close(); logErr == nil {
			logErr = err
		}
		if failed && logErr == nil {
			name := "output-" + hex.EncodeToString(digest.Sum(nil)) + ".log.gz"
			logErr = os.Rename(logFile.Name(), filepath.Join(filepath.Dir(c.Output), name))
			if logErr == nil {
				res.OutputLog = name
			}
		}
		if logErr != nil {
			log.Printf("Warning: full output log is unavailable: %v", logErr)
		}
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

	if c.Output != "" {
		dt, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return errors.Wrap(err, "failed to marshal result")
		}
		if err := os.WriteFile(c.Output, dt, 0644); err != nil {
			return errors.Wrap(err, "failed to write result to output file")
		}
	}
	if failed {
		return errors.New("tests or benchmarks failed; results have been preserved")
	}
	return nil
}
