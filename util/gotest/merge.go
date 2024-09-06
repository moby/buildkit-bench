package gotest

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Result struct {
	Tests         map[string]Test
	Benchmarks    map[string]Benchmark
	BenchmarkInfo BenchmarkInfo
}

func MergeBenchmarks(files []string) (map[string]Benchmark, error) {
	var allFiles []string
	for _, pattern := range files {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand glob pattern %s", pattern)
		}
		allFiles = append(allFiles, matches...)
	}
	log.Printf("Merging %d benchmark results file(s)", len(allFiles))

	var benchmarks map[string]Benchmark
	for _, file := range allFiles {
		log.Printf("Loading benchmark results from file %s", file)
		fi, err := os.Stat(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to stat file %s", file)
		}
		if fi.IsDir() {
			continue
		}
		result, err := loadResult(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load result from file %s", file)
		}
		if len(result.Benchmarks) == 0 {
			log.Printf("No benchmarks found in file %s", file)
			continue
		}
		if benchmarks == nil {
			benchmarks = result.Benchmarks
		} else {
			for k, v := range result.Benchmarks {
				benchmarks[k] = v
			}
		}
	}
	return benchmarks, nil
}

func loadResult(filename string) (*Result, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open result file %s", filename)
	}
	defer file.Close()
	var res Result
	if err := json.NewDecoder(file).Decode(&res); err != nil {
		return nil, errors.Wrapf(err, "failed to decode result file %s", filename)
	}
	return &res, nil
}
